package hashes

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"
)

type Algorithm string

const (
	SHA1   Algorithm = "sha1"
	SHA256 Algorithm = "sha256"
	SHA512 Algorithm = "sha512"
)

func ParseAlgorithm(s string) (Algorithm, error) {
	norm := strings.ToLower(strings.TrimSpace(s))
	norm = strings.ReplaceAll(norm, "-", "")
	norm = strings.TrimSuffix(norm, "sum")
	switch norm {
	case "sha1":
		return SHA1, nil
	case "sha256":
		return SHA256, nil
	case "sha512":
		return SHA512, nil
	default:
		return "", fmt.Errorf("%w: %q", ErrUnknownAlgorithm, s)
	}
}

func (a Algorithm) new() (hash.Hash, error) {
	switch a {
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	case SHA512:
		return sha512.New(), nil
	default:
		return nil, fmt.Errorf("%w: %q", ErrUnknownAlgorithm, a)
	}
}

func Sum(a Algorithm, r io.Reader) (string, error) {
	h, err := a.new()
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(h, r); err != nil {
		return "", fmt.Errorf("hash stream: %w", err)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func SumFile(a Algorithm, path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	return Sum(a, f)
}

func Equal(a, b string) bool {
	da, err := hex.DecodeString(strings.TrimSpace(a))
	if err != nil {
		return false
	}
	db, err := hex.DecodeString(strings.TrimSpace(b))
	if err != nil {
		return false
	}
	if len(da) != len(db) {
		return false
	}
	return subtle.ConstantTimeCompare(da, db) == 1
}

type Writer struct {
	w io.Writer
	h hash.Hash
}

func NewWriter(a Algorithm, w io.Writer) (*Writer, error) {
	h, err := a.new()
	if err != nil {
		return nil, err
	}
	return &Writer{w: w, h: h}, nil
}

func (w *Writer) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if n > 0 {
		w.h.Write(p[:n])
	}
	return n, err
}

func (w *Writer) Sum() string {
	return hex.EncodeToString(w.h.Sum(nil))
}
