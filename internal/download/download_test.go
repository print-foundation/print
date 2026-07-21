package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serveTLS(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewTLSServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func contentHandler(data []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "image.iso", time.Time{}, strings.NewReader(string(data)))
	}
}

func TestFetchWritesPartFile(t *testing.T) {
	data := []byte(strings.Repeat("A", 4096))
	srv := serveTLS(t, contentHandler(data))

	dest := filepath.Join(t.TempDir(), "image.iso")
	d := New(WithHTTPClient(srv.Client()))

	part, err := d.Fetch(context.Background(), Request{URL: srv.URL, Dest: dest})
	require.NoError(t, err)
	assert.Equal(t, dest+".part", part)

	got, err := os.ReadFile(part)
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestFetchRefusesHTTP(t *testing.T) {
	d := New()
	_, err := d.Fetch(context.Background(), Request{URL: "http://example.com/x", Dest: filepath.Join(t.TempDir(), "x")})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "https")
}

func TestFetchResumes(t *testing.T) {
	data := []byte(strings.Repeat("B", 8192))
	dest := filepath.Join(t.TempDir(), "image.iso")
	require.NoError(t, os.WriteFile(dest+".part", data[:2000], 0o644))

	var served string
	srv := serveTLS(t, func(w http.ResponseWriter, r *http.Request) {
		served = r.Header.Get("Range")
		http.ServeContent(w, r, "image.iso", time.Time{}, strings.NewReader(string(data)))
	})

	d := New(WithHTTPClient(srv.Client()))
	part, err := d.Fetch(context.Background(), Request{URL: srv.URL, Dest: dest})
	require.NoError(t, err)

	assert.Equal(t, "bytes=2000-", served)
	got, err := os.ReadFile(part)
	require.NoError(t, err)
	assert.Equal(t, data, got)
}

func TestProgressReported(t *testing.T) {
	data := []byte(strings.Repeat("C", 100000))
	srv := serveTLS(t, contentHandler(data))

	dest := filepath.Join(t.TempDir(), "image.iso")
	d := New(WithHTTPClient(srv.Client()))

	var last Progress
	var called bool
	_, err := d.Fetch(context.Background(), Request{
		URL:  srv.URL,
		Dest: dest,
		OnProgress: func(p Progress) {
			called = true
			last = p
		},
	})
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, domain.ByteSize(len(data)), last.Downloaded)
	assert.InDelta(t, 1.0, last.Fraction(), 0.0001)
}

func TestFetchVerifiedPromotesOnMatch(t *testing.T) {
	data := []byte("verified operating system image")
	srv := serveTLS(t, contentHandler(data))

	sum := sha256.Sum256(data)
	dest := filepath.Join(t.TempDir(), "image.iso")

	d := New(WithHTTPClient(srv.Client()))
	v := verify.New()

	res, err := d.FetchVerified(context.Background(), Request{URL: srv.URL, Dest: dest}, v, domain.Checksum{
		Algorithm: domain.SHA256,
		Value:     hex.EncodeToString(sum[:]),
	})
	require.NoError(t, err)
	assert.Equal(t, dest, res.Path)
	assert.True(t, res.Verify.ChecksumVerified)

	_, err = os.Stat(dest + ".part")
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(dest)
	assert.NoError(t, err)
}

func TestFetchVerifiedRemovesOnMismatch(t *testing.T) {
	data := []byte("bad image")
	srv := serveTLS(t, contentHandler(data))

	dest := filepath.Join(t.TempDir(), "image.iso")
	d := New(WithHTTPClient(srv.Client()))
	v := verify.New()

	_, err := d.FetchVerified(context.Background(), Request{URL: srv.URL, Dest: dest}, v, domain.Checksum{
		Algorithm: domain.SHA256,
		Value:     hex.EncodeToString(make([]byte, 32)),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrVerificationFailed)

	_, statErr := os.Stat(dest + ".part")
	assert.True(t, os.IsNotExist(statErr))
	_, statErr = os.Stat(dest)
	assert.True(t, os.IsNotExist(statErr))
}

func TestFetchVerifiedRequiresChecksum(t *testing.T) {
	d := New()
	v := verify.New()
	_, err := d.FetchVerified(context.Background(), Request{URL: "https://e.com/x", Dest: "x"}, v, domain.Checksum{})
	assert.ErrorIs(t, err, domain.ErrVerificationFailed)
}

func TestProgressHelpers(t *testing.T) {
	p := Progress{Downloaded: 50, Total: 100, BytesPerSecond: 10}
	assert.InDelta(t, 0.5, p.Fraction(), 0.0001)
	assert.Equal(t, 5, int(p.ETA().Seconds()))

	unknown := Progress{Downloaded: 50, Total: 0}
	assert.Equal(t, 0.0, unknown.Fraction())
	assert.Equal(t, 0, int(unknown.ETA()))
}
