package hashes

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAlgorithm(t *testing.T) {
	cases := map[string]Algorithm{
		"sha256":    SHA256,
		"SHA-256":   SHA256,
		"sha256sum": SHA256,
		"SHA512":    SHA512,
		"sha1":      SHA1,
	}
	for in, want := range cases {
		got, err := ParseAlgorithm(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got)
	}

	_, err := ParseAlgorithm("md5")
	assert.ErrorIs(t, err, ErrUnknownAlgorithm)
}

func TestSumMatchesStdlib(t *testing.T) {
	data := []byte("the quick brown fox")
	want := sha256.Sum256(data)

	got, err := Sum(SHA256, bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, hex.EncodeToString(want[:]), got)
}

func TestSumFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.bin")
	require.NoError(t, os.WriteFile(path, []byte("hello world"), 0o600))

	got, err := SumFile(SHA256, path)
	require.NoError(t, err)

	want, err := Sum(SHA256, strings.NewReader("hello world"))
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestEqual(t *testing.T) {
	assert.True(t, Equal("ABCDEF12", "abcdef12"))
	assert.True(t, Equal(" abcdef12 ", "abcdef12"))
	assert.False(t, Equal("abcdef12", "abcdef34"))
	assert.False(t, Equal("abcd", "abcdef12"))
	assert.False(t, Equal("nothex", "nothex"))
}

func TestWriter(t *testing.T) {
	var buf bytes.Buffer
	w, err := NewWriter(SHA256, &buf)
	require.NoError(t, err)

	_, err = w.Write([]byte("hello world"))
	require.NoError(t, err)

	assert.Equal(t, "hello world", buf.String())

	want, err := Sum(SHA256, strings.NewReader("hello world"))
	require.NoError(t, err)
	assert.Equal(t, want, w.Sum())
}

func TestParseChecksumFile(t *testing.T) {
	content := `# a comment
d2d2d2  ubuntu-24.04-desktop-amd64.iso
abcabc *ubuntu-24.04-server-amd64.iso

`
	cf, err := ParseChecksumFile(strings.NewReader(content), SHA256)
	require.NoError(t, err)

	d, ok := cf.Lookup("ubuntu-24.04-desktop-amd64.iso")
	assert.True(t, ok)
	assert.Equal(t, "d2d2d2", d)

	d, ok = cf.Lookup("ubuntu-24.04-server-amd64.iso")
	assert.True(t, ok)
	assert.Equal(t, "abcabc", d)

	d, ok = cf.Lookup("https://example.com/path/ubuntu-24.04-desktop-amd64.iso")
	assert.True(t, ok)
	assert.Equal(t, "d2d2d2", d)

	_, ok = cf.Lookup("missing.iso")
	assert.False(t, ok)
}

func TestParseChecksumFileInfersAlgorithm(t *testing.T) {
	line := strings.Repeat("a", 64) + "  file.iso"
	cf, err := ParseChecksumFile(strings.NewReader(line), "")
	require.NoError(t, err)
	assert.Equal(t, SHA256, cf.Algorithm)
}
