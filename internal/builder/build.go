package builder

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/internal/mirrors"
	"github.com/print-foundation/print/pkg/hashes"
)

type ClientConfig struct {
	Distro  mirrors.Distro `json:"distro"`
	Arch    Arch           `json:"arch"`
	Mirror  string         `json:"mirror"`
	WiFi    WiFiConfig     `json:"wifi,omitempty"`
	Version string         `json:"version"`
}

type Result struct {
	ISO      string
	Verified bool
}

type Builder struct {
	log    logging.Logger
	client *http.Client
}

func New(opts ...Option) *Builder {
	b := &Builder{log: logging.NopLogger(), client: http.DefaultClient}
	for _, o := range opts {
		o(b)
	}
	return b
}

type Option func(*Builder)

func WithLogger(l logging.Logger) Option { return func(b *Builder) { b.log = l } }

func WithHTTPClient(c *http.Client) Option { return func(b *Builder) { b.client = c } }

func (b *Builder) Build(ctx context.Context, spec Spec, clientBinary []byte) (Result, error) {
	if !mirrors.Supported(spec.Distro) {
		return Result{}, fmt.Errorf("%w: %s", errUnsupported, spec.Distro)
	}
	asset, ok := AssetFor(spec.Distro, spec.Arch)
	if !ok {
		return Result{}, fmt.Errorf("%w: %s/%s", errNoAsset, spec.Distro, spec.Arch)
	}

	work := spec.WorkDir
	if work == "" {
		d, err := os.MkdirTemp("", "print-build-")
		if err != nil {
			return Result{}, err
		}
		defer os.RemoveAll(d)
		work = d
	} else {
		if err := os.MkdirAll(work, 0o755); err != nil {
			return Result{}, err
		}
	}

	base := strings.TrimRight(spec.Mirror.BaseURL, "/")
	files := map[string]string{}
	if asset.Kernel != "" {
		files["kernel"] = base + asset.Kernel
	}
	if asset.Initrd != "" {
		files["initrd"] = base + asset.Initrd
	}
	for i, e := range asset.Extra {
		files[fmt.Sprintf("extra%d", i)] = base + e
	}

	for name, url := range files {
		dest := filepath.Join(work, name)
		if err := b.fetch(ctx, url, dest); err != nil {
			return Result{}, fmt.Errorf("fetch %s: %w", name, err)
		}
	}

	verified := false
	if b.maybeVerify(ctx, base+"/SHA256SUMS", work) {
		verified = true
		b.log.Info("netboot assets verified against publisher SHA256SUMS")
	} else {
		b.log.Warn("no publisher SHA256SUMS found; assets not checksum-verified", "distro", spec.Distro)
	}

	cfg := ClientConfig{
		Distro:  spec.Distro,
		Arch:    spec.Arch,
		Mirror:  base,
		WiFi:    spec.WiFi,
		Version: "cloud-netboot",
	}
	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return Result{}, err
	}
	if err := os.WriteFile(filepath.Join(work, "print-client.json"), cfgBytes, 0o644); err != nil {
		return Result{}, err
	}
	clientPath := filepath.Join(work, "print-client")
	if err := os.WriteFile(clientPath, clientBinary, 0o755); err != nil {
		return Result{}, err
	}

	isoPath := spec.Output
	if isoPath == "" {
		isoPath = filepath.Join(work, "print-"+string(spec.Distro)+".iso")
	}
	if err := assembleISO(work, isoPath, spec, b.log); err != nil {
		return Result{}, err
	}

	info, err := os.Stat(isoPath)
	if err != nil {
		return Result{}, err
	}
	if info.Size() == 0 {
		return Result{}, fmt.Errorf("assembled iso is empty")
	}
	return Result{ISO: isoPath, Verified: verified}, nil
}

func (b *Builder) fetch(ctx context.Context, url, dest string) error {
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("refusing non-https url %q", url)
	}
	var existing int64
	if fi, err := os.Stat(dest); err == nil {
		existing = fi.Size()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if existing > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existing))
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("download %s: status %d", url, resp.StatusCode)
	}
	f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	return err
}

func (b *Builder) maybeVerify(ctx context.Context, sumURL, work string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, sumURL, nil)
	if err != nil {
		return false
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return false
	}
	cf, err := hashes.ParseChecksumFile(strings.NewReader(string(raw)), "")
	if err != nil {
		return false
	}
	for _, local := range []string{"kernel", "initrd"} {
		path := filepath.Join(work, local)
		if _, err := os.Stat(path); err != nil {
			continue
		}
		if _, ok := cf.Lookup(filepath.Base(path)); !ok {
			return false
		}
		digest, _ := cf.Lookup(filepath.Base(path))
		got, err := hashes.SumFile(cf.Algorithm, path)
		if err != nil || !hashes.Equal(got, digest) {
			return false
		}
	}
	return true
}

var (
	errUnsupported = fmt.Errorf("unsupported distro")
	errNoAsset     = fmt.Errorf("no netboot asset")
)
