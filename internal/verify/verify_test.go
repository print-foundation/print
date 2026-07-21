package verify

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/pkg/osdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "artifact.bin")
	require.NoError(t, os.WriteFile(path, data, 0o600))
	return path
}

func TestVerifyFileSuccess(t *testing.T) {
	data := []byte("operating system image")
	path := writeTemp(t, data)
	sum := sha256.Sum256(data)

	v := New()
	res, err := v.VerifyFile(path, domain.Checksum{
		Algorithm: domain.SHA256,
		Value:     hex.EncodeToString(sum[:]),
	})
	require.NoError(t, err)
	assert.True(t, res.ChecksumVerified)
}

func TestVerifyFileMismatch(t *testing.T) {
	path := writeTemp(t, []byte("real content"))
	v := New()
	_, err := v.VerifyFile(path, domain.Checksum{
		Algorithm: domain.SHA256,
		Value:     hex.EncodeToString(make([]byte, 32)),
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrVerificationFailed)
}

func TestVerifyFileNoExpected(t *testing.T) {
	path := writeTemp(t, []byte("x"))
	v := New()
	_, err := v.VerifyFile(path, domain.Checksum{})
	assert.ErrorIs(t, err, domain.ErrVerificationFailed)
}

func TestExpectedChecksumEmbedded(t *testing.T) {
	rel := osdb.Release{
		ID:       "x-1",
		URL:      "https://example.com/x.iso",
		Checksum: osdb.Checksum{Algorithm: "sha256", Value: "ABCDEF"},
	}
	v := New()
	got, err := v.ExpectedChecksum(context.Background(), rel)
	require.NoError(t, err)
	assert.Equal(t, domain.SHA256, got.Algorithm)
	assert.Equal(t, "abcdef", got.Value) // lowercased
}

func TestExpectedChecksumFromFile(t *testing.T) {
	data := []byte("the image bytes")
	sum := sha256.Sum256(data)
	digest := hex.EncodeToString(sum[:])

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s  the-image.iso\n", digest)
	}))
	defer srv.Close()

	rel := osdb.Release{
		ID:          "x-1",
		URL:         "https://example.com/the-image.iso",
		ChecksumURL: "https://example.com/SHA256SUMS", // scheme checked, host from client
	}

	v := New(WithHTTPClient(rewriteClient{to: srv.URL}))
	got, err := v.ExpectedChecksum(context.Background(), rel)
	require.NoError(t, err)
	assert.Equal(t, digest, got.Value)
	assert.Equal(t, domain.SHA256, got.Algorithm)
}

func TestExpectedChecksumRefusesHTTP(t *testing.T) {
	rel := osdb.Release{ID: "x", URL: "https://e.com/x.iso", ChecksumURL: "http://e.com/sums"}
	v := New()
	_, err := v.ExpectedChecksum(context.Background(), rel)
	require.Error(t, err)
}

type rewriteClient struct {
	to string
}

func (c rewriteClient) Do(req *http.Request) (*http.Response, error) {
	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, c.to+req.URL.Path, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(newReq)
}

func TestPGPRoundTrip(t *testing.T) {
	entity, err := openpgp.NewEntity("Test Publisher", "test key", "test@example.com", nil)
	require.NoError(t, err)

	message := []byte("d41d8cd98f00b204e9800998ecf8427e  image.iso\n")
	var sigBuf bytes.Buffer
	require.NoError(t, openpgp.DetachSign(&sigBuf, entity, bytes.NewReader(message), nil))

	var pubBuf bytes.Buffer
	require.NoError(t, entity.Serialize(&pubBuf))

	v, err := NewPGPVerifier(bytes.NewReader(pubBuf.Bytes()))
	require.NoError(t, err)
	assert.Equal(t, 1, v.KeyCount())

	require.NoError(t, v.Verify(context.Background(), message, sigBuf.Bytes()))

	err = v.Verify(context.Background(), []byte("tampered"), sigBuf.Bytes())
	assert.Error(t, err)
}

func TestPGPNoTrustedKey(t *testing.T) {
	v, err := NewPGPVerifier()
	require.NoError(t, err)
	err = v.Verify(context.Background(), []byte("x"), []byte("sig"))
	assert.ErrorIs(t, err, ErrNoTrustedKey)
}

func TestLoadPGPKeyDirMissing(t *testing.T) {
	v, err := LoadPGPKeyDir(filepath.Join(t.TempDir(), "nope"))
	require.NoError(t, err)
	assert.Equal(t, 0, v.KeyCount())
}
