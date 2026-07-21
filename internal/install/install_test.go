package install

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/print-foundation/print/internal/disk"
	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/download"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/internal/verify"
	"github.com/print-foundation/print/pkg/hashes"
	"github.com/print-foundation/print/pkg/osdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTestImage(t *testing.T) (path, digest string) {
	t.Helper()
	dir := t.TempDir()
	path = filepath.Join(dir, "image.iso")
	payload := make([]byte, 2*1024*1024+123)
	for i := range payload {
		payload[i] = byte(i % 251)
	}
	require.NoError(t, os.WriteFile(path, payload, 0o644))
	d, err := hashes.SumFile(hashes.SHA256, path)
	require.NoError(t, err)
	return path, d
}

func newEngine(t *testing.T, srv *httptest.Server) (*Engine, *download.Downloader, *verify.Verifier) {
	client := srv.Client()
	client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	d := download.New(download.WithHTTPClient(client), download.WithLogger(logging.NopLogger()))
	v := verify.New()
	return NewEngine(d, v,
		WithLogger(logging.NopLogger()),
		WithDeviceWriter(FileDeviceWriter{}),
	), d, v
}

func serveImage(t *testing.T, imgPath string) *httptest.Server {
	t.Helper()
	return httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.Open(imgPath)
		require.NoError(t, err)
		defer f.Close()
		io.Copy(w, f) //nolint:errcheck
	}))
}

func TestEngineRunWritesVerifiedImage(t *testing.T) {
	imgPath, digest := writeTestImage(t)
	srv := serveImage(t, imgPath)
	defer srv.Close()

	eng, _, _ := newEngine(t, srv)
	cache := t.TempDir()
	rel := osdb.Release{
		ID:       "test/os-1.0",
		URL:      srv.URL + "/image.iso",
		Size:     uint64(2*1024*1024 + 123),
		Checksum: osdb.Checksum{Algorithm: "sha256", Value: digest},
		Format:   "iso",
		Version:  "1.0",
	}

	out := filepath.Join(t.TempDir(), "out.iso")
	target := domain.Disk{Path: out, Size: domain.ByteSize(8 * 1024 * 1024 * 1024), Model: "Virtual"}
	conf := disk.Confirmation{DevicePath: target.Path, Acknowledged: true}

	var phases []Phase
	err := eng.Run(context.Background(), Request{
		Mode:     ModeAutomatic,
		Release:  rel,
		Target:   target,
		Confirm:  conf,
		CacheDir: cache,
		Firmware: domain.FirmwareUEFI,
	}, func(ev Event) { phases = append(phases, ev.Phase) })
	require.NoError(t, err)

	written, err := hashes.SumFile(hashes.SHA256, out)
	require.NoError(t, err)
	assert.Equal(t, digest, written)
	assert.Equal(t, PhaseDone, phases[len(phases)-1])
}

func TestEngineRefusesUnconfirmedWrite(t *testing.T) {
	imgPath, digest := writeTestImage(t)
	srv := serveImage(t, imgPath)
	defer srv.Close()

	eng, _, _ := newEngine(t, srv)
	cache := t.TempDir()
	rel := osdb.Release{ID: "x/y", URL: srv.URL + "/image.iso", Size: uint64(2*1024*1024 + 123), Checksum: osdb.Checksum{Algorithm: "sha256", Value: digest}, Format: "iso"}

	err := eng.Run(context.Background(), Request{
		Release:  rel,
		Target:   domain.Disk{Path: filepath.Join(t.TempDir(), "out"), Size: 1 << 30},
		Confirm:  disk.Confirmation{DevicePath: "wrong", Acknowledged: true},
		CacheDir: cache,
	}, nil)
	assert.ErrorIs(t, err, domain.ErrConfirmationRequired)
}

func TestEngineRefusesUnverifiableRelease(t *testing.T) {
	eng, _, _ := newEngine(t, httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))
	rel := osdb.Release{ID: "x", URL: "https://e/x", Size: 10, Format: "iso"} // no checksum
	err := eng.Run(context.Background(), Request{
		Release:  rel,
		Target:   domain.Disk{Path: "x"},
		Confirm:  disk.Confirmation{DevicePath: "x", Acknowledged: true},
		CacheDir: t.TempDir(),
	}, nil)
	assert.ErrorIs(t, err, domain.ErrVerificationFailed)
}

func TestCachePathIsDeterministic(t *testing.T) {
	rel := osdb.Release{ID: "ubuntu/24.04", Format: "iso"}
	a := filepath.Base(cachePath(t.TempDir(), rel))
	b := filepath.Base(cachePath(t.TempDir(), rel))
	assert.Equal(t, a, b)
	assert.Equal(t, "ubuntu_24.04.iso", a)
}

func TestRecoveryScanFindsCompleteImage(t *testing.T) {
	cache := t.TempDir()
	imgPath, digest := writeTestImage(t)
	rel := osdb.Release{ID: "test/os-1.0", Format: "iso", Size: uint64(2*1024*1024 + 123), Checksum: osdb.Checksum{Algorithm: "sha256", Value: digest}}

	dest := filepath.Join(cache, "images", cacheNameFor(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(dest), 0o755))
	src, err := os.Open(imgPath)
	require.NoError(t, err)
	defer src.Close()
	dst, err := os.Create(dest)
	require.NoError(t, err)
	_, err = io.Copy(dst, src)
	require.NoError(t, err)
	require.NoError(t, dst.Close())

	recs := RecoveryScan(cache, []osdb.Release{rel})
	require.Len(t, recs, 1)
	assert.True(t, recs[0].Complete)
}
