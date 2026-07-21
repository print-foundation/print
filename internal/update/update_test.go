package update

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/print-foundation/print/pkg/hashes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatestAndDownloadVerifies(t *testing.T) {
	payload := []byte("print-binary-bytes")
	digest, err := hashes.Sum(hashes.SHA256, bytes.NewReader(payload))
	require.NoError(t, err)

	manifest := Manifest{
		Version: "1.2.3",
		Assets: map[string]Asset{
			platformKey(): {URL: "", Checksum: digest, Size: uint64(len(payload))},
		},
	}

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/manifest.json":
			json.NewEncoder(w).Encode(manifest) //nolint:errcheck
		case "/asset":
			w.Write(payload) //nolint:errcheck
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	client := srv.Client()
	client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c := New(WithHTTPClient(client))

	m, err := c.Latest(context.Background(), srv.URL+"/manifest.json")
	require.NoError(t, err)
	assert.Equal(t, "1.2.3", m.Version)

	asset, ok := m.AssetFor(platformKey())
	require.True(t, ok)
	asset.URL = srv.URL + "/asset"

	dest := filepath.Join(t.TempDir(), "print.new")
	require.NoError(t, c.Download(context.Background(), asset, dest, nil))

	got, err := hashes.SumFile(hashes.SHA256, dest)
	require.NoError(t, err)
	assert.Equal(t, digest, got)
}

func TestLatestRejectsPlaintext(t *testing.T) {
	c := New()
	_, err := c.Latest(context.Background(), "http://example/manifest.json")
	assert.Error(t, err)
}

func TestDownloadRejectsChecksumMismatch(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("real-bytes")) //nolint:errcheck
	}))
	defer srv.Close()
	client := srv.Client()
	client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
	c := New(WithHTTPClient(client))

	asset := Asset{URL: srv.URL + "/x", Checksum: "deadbeef", Size: 10}
	dest := filepath.Join(t.TempDir(), "bad")
	err := c.Download(context.Background(), asset, dest, nil)
	assert.Error(t, err)
}

func TestApplyReplaceOnWindowsKeepsOld(t *testing.T) {
	dir := t.TempDir()
	exe := filepath.Join(dir, "print.exe")
	require.NoError(t, os.WriteFile(exe, []byte("old"), 0o755))
	newBin := filepath.Join(dir, "print.new")
	require.NoError(t, os.WriteFile(newBin, []byte("new"), 0o755))

	require.NoError(t, replaceBinary(exe, newBin))
	got, err := os.ReadFile(exe)
	require.NoError(t, err)
	assert.Equal(t, "new", string(got))
}
