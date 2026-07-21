package install

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/print-foundation/print/internal/disk"
	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/download"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/internal/plugin"
	"github.com/print-foundation/print/internal/verify"
	"github.com/print-foundation/print/pkg/osdb"
)

type Mode string

const (
	ModeAutomatic Mode = "automatic"
	ModeAdvanced  Mode = "advanced"
)

type Phase string

const (
	PhasePreparing   Phase = "preparing"
	PhaseDownloading Phase = "downloading"
	PhaseVerifying   Phase = "verifying"
	PhaseWriting     Phase = "writing"
	PhaseFinalizing  Phase = "finalizing"
	PhaseDone        Phase = "done"
	PhaseFailed      Phase = "failed"
)

type Event struct {
	Phase    Phase
	Message  string
	Fraction float64
	Err      error
}

type EventFunc func(Event)

type Request struct {
	Mode    Mode
	Release osdb.Release
	Target  domain.Disk

	Confirm disk.Confirmation

	CacheDir string

	Plan *disk.Plan

	Firmware domain.Firmware
}

type Engine struct {
	downloader *download.Downloader
	verifier   *verify.Verifier
	writer     DeviceWriter
	log        logging.Logger
	clock      func() time.Time
}

type Option func(*Engine)

func WithLogger(l logging.Logger) Option { return func(e *Engine) { e.log = l } }

func WithDeviceWriter(w DeviceWriter) Option { return func(e *Engine) { e.writer = w } }

func NewEngine(d *download.Downloader, v *verify.Verifier, opts ...Option) *Engine {
	e := &Engine{
		downloader: d,
		verifier:   v,
		writer:     NewDeviceWriter(),
		log:        logging.NopLogger(),
		clock:      time.Now,
	}
	for _, o := range opts {
		o(e)
	}
	return e
}

func (e *Engine) Run(ctx context.Context, req Request, onEvent EventFunc) (err error) {
	emit := func(ev Event) {
		if onEvent != nil {
			onEvent(ev)
		}
	}
	defer func() {
		if err != nil {
			emit(Event{Phase: PhaseFailed, Message: err.Error(), Err: err, Fraction: -1})
		}
	}()

	if err = e.validate(req); err != nil {
		return err
	}

	emit(Event{Phase: PhasePreparing, Message: "preparing installation", Fraction: -1})

	if err := e.fireHooks(ctx, plugin.PhaseStart, req); err != nil {
		return err
	}

	want, err := e.verifier.ExpectedChecksum(ctx, req.Release)
	if err != nil {
		return fmt.Errorf("resolve checksum: %w", err)
	}

	imagePath, err := e.acquire(ctx, req, want, emit)
	if err != nil {
		return err
	}

	if err := e.fireHooks(ctx, plugin.PhaseDownloaded, req); err != nil {
		return err
	}

	emit(Event{Phase: PhaseVerifying, Message: "verifying image", Fraction: -1})
	if _, verr := e.verifier.VerifyFile(imagePath, want); verr != nil {
		return verr
	}

	if err := e.fireHooks(ctx, plugin.PhaseVerified, req); err != nil {
		return err
	}

	if guardErr := disk.Guard(req.Target, req.Confirm); guardErr != nil {
		return guardErr
	}

	emit(Event{Phase: PhaseWriting, Message: "writing to " + req.Target.Path, Fraction: 0})
	if werr := e.write(ctx, imagePath, req.Target, emit); werr != nil {
		return werr
	}

	if err := e.fireHooks(ctx, plugin.PhaseWritten, req); err != nil {
		return err
	}

	emit(Event{Phase: PhaseFinalizing, Message: "flushing buffers", Fraction: -1})
	if ferr := e.writer.Finalize(ctx, req.Target.Path); ferr != nil {
		return fmt.Errorf("finalize: %w", ferr)
	}

	_ = e.fireHooks(ctx, plugin.PhaseEnd, req)

	emit(Event{Phase: PhaseDone, Message: "installation complete", Fraction: 1})
	return nil
}

func (e *Engine) fireHooks(ctx context.Context, phase plugin.Phase, req Request) error {
	pr := plugin.Request{ReleaseID: req.Release.ID, Target: req.Target.Path}
	return plugin.RunHooks(ctx, phase, pr)
}

func (e *Engine) validate(req Request) error {
	if req.Release.Checksum.IsZero() && req.Release.ChecksumURL == "" {
		return fmt.Errorf("%w: release %s has no verification source", domain.ErrVerificationFailed, req.Release.ID)
	}
	if req.Target.Path == "" {
		return errors.New("no target disk selected")
	}
	if req.CacheDir == "" {
		return errors.New("cache directory required")
	}
	return nil
}

func (e *Engine) acquire(ctx context.Context, req Request, want domain.Checksum, emit EventFunc) (string, error) {
	dest := cachePath(req.CacheDir, req.Release)

	res, err := e.downloader.FetchVerified(ctx, download.Request{
		URL:          req.Release.URL,
		Dest:         dest,
		ExpectedSize: domain.ByteSize(req.Release.Size),
		OnProgress: func(p download.Progress) {
			emit(Event{
				Phase:    PhaseDownloading,
				Message:  fmt.Sprintf("downloading %s", req.Release.Version),
				Fraction: p.Fraction(),
			})
		},
	}, e.verifier, want)
	if err != nil {
		return "", fmt.Errorf("download: %w", err)
	}
	return res.Path, nil
}

func (e *Engine) write(ctx context.Context, imagePath string, target domain.Disk, emit EventFunc) error {
	err := e.writer.Write(ctx, imagePath, target.Path, func(written, total domain.ByteSize) {
		frac := -1.0
		if total > 0 {
			frac = float64(written) / float64(total)
		}
		emit(Event{Phase: PhaseWriting, Message: "writing image", Fraction: frac})
	})
	if err != nil {
		return fmt.Errorf("write image: %w", err)
	}
	return nil
}
