package download

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type Progress struct {
	Downloaded     domain.ByteSize
	Total          domain.ByteSize
	BytesPerSecond float64
}

func (p Progress) Fraction() float64 {
	if p.Total == 0 {
		return 0
	}
	f := float64(p.Downloaded) / float64(p.Total)
	if f > 1 {
		return 1
	}
	return f
}

func (p Progress) ETA() time.Duration {
	if p.BytesPerSecond <= 0 || p.Total <= p.Downloaded {
		return 0
	}
	remaining := float64(p.Total - p.Downloaded)
	return time.Duration(remaining/p.BytesPerSecond) * time.Second
}

type ProgressFunc func(Progress)

type Request struct {
	URL          string
	Dest         string
	ExpectedSize domain.ByteSize
	OnProgress   ProgressFunc
}

type Downloader struct {
	client           *http.Client
	log              logging.Logger
	progressInterval time.Duration
}

type Option func(*Downloader)

func WithLogger(l logging.Logger) Option {
	return func(d *Downloader) { d.log = l }
}

func WithHTTPClient(c *http.Client) Option {
	return func(d *Downloader) { d.client = c }
}

func New(opts ...Option) *Downloader {
	d := &Downloader{
		client:           defaultClient(),
		log:              logging.NopLogger(),
		progressInterval: 200 * time.Millisecond,
	}
	for _, o := range opts {
		o(d)
	}
	return d
}

func defaultClient() *http.Client {
	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        16,
		IdleConnTimeout:     90 * time.Second,
		TLSHandshakeTimeout: 15 * time.Second,
	}
	return &http.Client{Transport: transport}
}

func (d *Downloader) Fetch(ctx context.Context, req Request) (string, error) {
	if !strings.HasPrefix(req.URL, "https://") {
		return "", fmt.Errorf("refusing non-https url %q", req.URL)
	}
	if err := os.MkdirAll(filepath.Dir(req.Dest), 0o755); err != nil {
		return "", fmt.Errorf("create dest dir: %w", err)
	}

	partPath := req.Dest + ".part"
	existing := resumeOffset(partPath)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, req.URL, nil)
	if err != nil {
		return "", err
	}
	if existing > 0 {
		httpReq.Header.Set("Range", fmt.Sprintf("bytes=%d-", existing))
	}

	resp, err := d.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	offset, total, err := d.resolveRanges(resp, existing, req.ExpectedSize)
	if err != nil {
		return "", err
	}

	flag := os.O_CREATE | os.O_WRONLY
	if offset > 0 {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}
	f, err := os.OpenFile(partPath, flag, 0o644)
	if err != nil {
		return "", fmt.Errorf("open part file: %w", err)
	}
	defer f.Close()

	if err := d.copyWithProgress(ctx, f, resp.Body, offset, total, req.OnProgress); err != nil {
		return "", err
	}
	if err := f.Sync(); err != nil {
		return "", fmt.Errorf("sync part file: %w", err)
	}
	return partPath, nil
}

func (d *Downloader) resolveRanges(resp *http.Response, requested int64, expected domain.ByteSize) (offset int64, total domain.ByteSize, err error) {
	switch resp.StatusCode {
	case http.StatusOK:
		if requested > 0 {
			d.log.Info("server ignored range, restarting download")
		}
		return 0, sizeFromContentLength(resp.ContentLength, expected, 0), nil
	case http.StatusPartialContent:
		return requested, sizeFromContentLength(resp.ContentLength, expected, requested), nil
	case http.StatusRequestedRangeNotSatisfiable:
		return requested, expected, errAlreadyComplete
	default:
		return 0, 0, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}
}

func sizeFromContentLength(cl int64, expected domain.ByteSize, offset int64) domain.ByteSize {
	if cl > 0 {
		return domain.ByteSize(offset + cl)
	}
	return expected
}

var errAlreadyComplete = errors.New("download already complete")

func (d *Downloader) copyWithProgress(ctx context.Context, dst io.Writer, src io.Reader, start int64, total domain.ByteSize, onProgress ProgressFunc) error {
	buf := make([]byte, 512*1024)
	downloaded := start
	startTime := time.Now()
	lastReport := time.Now()

	report := func() {
		if onProgress == nil {
			return
		}
		elapsed := time.Since(startTime).Seconds()
		var rate float64
		if elapsed > 0 {
			rate = float64(downloaded-start) / elapsed
		}
		onProgress(Progress{
			Downloaded:     domain.ByteSize(downloaded),
			Total:          total,
			BytesPerSecond: rate,
		})
	}

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", domain.ErrCancelled, ctx.Err())
		default:
		}

		n, readErr := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return fmt.Errorf("write: %w", werr)
			}
			downloaded += int64(n)
			if time.Since(lastReport) >= d.progressInterval {
				report()
				lastReport = time.Now()
			}
		}
		if readErr == io.EOF {
			report() // final tick so the UI lands on 100%
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read: %w", readErr)
		}
	}
}

func resumeOffset(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}
