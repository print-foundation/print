package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/pkg/hashes"
)

type Manifest struct {
	Version string           `json:"version"`
	Notes   string           `json:"notes"`
	Assets  map[string]Asset `json:"assets"`
}

type Asset struct {
	URL      string `json:"url"`
	Checksum string `json:"checksum"`
	Size     uint64 `json:"size"`
}

func platformKey() string {
	return runtime.GOOS + "/" + runtime.GOARCH
}

type Client struct {
	http *http.Client
	log  logging.Logger
}

func New(opts ...Option) *Client {
	c := &Client{http: http.DefaultClient, log: logging.NopLogger()}
	for _, o := range opts {
		o(c)
	}
	return c
}

type Option func(*Client)

func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

func WithLogger(l logging.Logger) Option { return func(c *Client) { c.log = l } }

func (c *Client) Latest(ctx context.Context, url string) (*Manifest, error) {
	if len(url) < 8 || url[:8] != "https://" {
		return nil, fmt.Errorf("refusing non-https manifest url %q", url)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest status %d", resp.StatusCode)
	}
	var m Manifest
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&m); err != nil {
		return nil, fmt.Errorf("decode manifest: %w", err)
	}
	return &m, nil
}

func (m *Manifest) AssetFor(platform string) (Asset, bool) {
	a, ok := m.Assets[platform]
	return a, ok
}

func (c *Client) Download(ctx context.Context, asset Asset, dest string, onProgress func(float64)) error {
	if len(asset.URL) < 8 || asset.URL[:8] != "https://" {
		return fmt.Errorf("refusing non-https asset url %q", asset.URL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, nil)
	if err != nil {
		return err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("asset status %d", resp.StatusCode)
	}

	tmp := dest + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	total := resp.ContentLength
	var written int64

	buf := make([]byte, 1<<16)
	for {
		select {
		case <-ctx.Done():
			f.Close()
			return fmt.Errorf("%w: %v", domain.ErrCancelled, ctx.Err())
		default:
		}
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := f.Write(buf[:n]); werr != nil {
				f.Close()
				return werr
			}
			written += int64(n)
			if onProgress != nil && total > 0 {
				onProgress(float64(written) / float64(total))
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			f.Close()
			return rerr
		}
	}
	if err := f.Close(); err != nil {
		return err
	}

	got, err := hashes.SumFile(hashes.SHA256, tmp)
	if err != nil {
		return err
	}
	if !hashes.Equal(got, asset.Checksum) {
		return fmt.Errorf("%w: update asset checksum mismatch", domain.ErrVerificationFailed)
	}
	return os.Rename(tmp, dest)
}

func Apply(newBinary string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	return replaceBinary(exe, newBinary)
}

var CurrentVersion = "dev"
