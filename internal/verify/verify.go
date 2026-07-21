package verify

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
	"github.com/print-foundation/print/pkg/hashes"
	"github.com/print-foundation/print/pkg/osdb"
)

type Result struct {
	ChecksumVerified bool
	Algorithm        hashes.Algorithm
	Digest           string

	SignatureVerified bool
	SignatureChecked  bool
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type SignatureVerifier interface {
	Verify(ctx context.Context, data, sig []byte) error
}

type Verifier struct {
	client HTTPClient
	sig    SignatureVerifier
	log    logging.Logger
}

type Option func(*Verifier)

func WithHTTPClient(c HTTPClient) Option {
	return func(v *Verifier) { v.client = c }
}

func WithSignatureVerifier(s SignatureVerifier) Option {
	return func(v *Verifier) { v.sig = s }
}

func WithLogger(l logging.Logger) Option {
	return func(v *Verifier) { v.log = l }
}

func New(opts ...Option) *Verifier {
	v := &Verifier{
		client: http.DefaultClient,
		log:    logging.NopLogger(),
	}
	for _, o := range opts {
		o(v)
	}
	return v
}

func (v *Verifier) ExpectedChecksum(ctx context.Context, rel osdb.Release) (domain.Checksum, error) {
	if !rel.Checksum.IsZero() {
		algo, err := hashes.ParseAlgorithm(rel.Checksum.Algorithm)
		if err != nil {
			return domain.Checksum{}, err
		}
		return domain.Checksum{
			Algorithm: domain.HashAlgorithm(algo),
			Value:     strings.ToLower(rel.Checksum.Value),
		}, nil
	}

	if rel.ChecksumURL == "" {
		return domain.Checksum{}, fmt.Errorf("%w: no checksum available for %s", domain.ErrVerificationFailed, rel.ID)
	}

	cf, raw, err := v.fetchChecksumFile(ctx, rel.ChecksumURL)
	if err != nil {
		return domain.Checksum{}, err
	}

	if rel.SignatureURL != "" && v.sig != nil {
		if err := v.verifyChecksumSignature(ctx, raw, rel.SignatureURL); err != nil {
			return domain.Checksum{}, err
		}
	}

	digest, ok := cf.Lookup(rel.URL)
	if !ok {
		return domain.Checksum{}, fmt.Errorf("%w: %s not listed in checksum file", domain.ErrVerificationFailed, rel.ID)
	}
	return domain.Checksum{
		Algorithm: domain.HashAlgorithm(cf.Algorithm),
		Value:     strings.ToLower(digest),
	}, nil
}

func (v *Verifier) VerifyFile(path string, want domain.Checksum) (Result, error) {
	if want.IsZero() {
		return Result{}, fmt.Errorf("%w: no expected checksum", domain.ErrVerificationFailed)
	}
	algo, err := hashes.ParseAlgorithm(string(want.Algorithm))
	if err != nil {
		return Result{}, err
	}
	got, err := hashes.SumFile(algo, path)
	if err != nil {
		return Result{}, fmt.Errorf("hash file: %w", err)
	}
	if !hashes.Equal(got, want.Value) {
		v.log.Error("checksum mismatch", "path", path, "want", want.Value, "got", got)
		return Result{}, fmt.Errorf("%w: %s digest mismatch", domain.ErrVerificationFailed, algo)
	}
	return Result{
		ChecksumVerified: true,
		Algorithm:        algo,
		Digest:           got,
	}, nil
}

func (v *Verifier) verifyChecksumSignature(ctx context.Context, checksumData []byte, sigURL string) error {
	sig, err := v.fetchBytes(ctx, sigURL)
	if err != nil {
		return fmt.Errorf("fetch signature: %w", err)
	}
	err = v.sig.Verify(ctx, checksumData, sig)
	if errors.Is(err, ErrNoTrustedKey) {
		v.log.Warn("checksum signature not verified: no trusted key", "url", sigURL)
		return nil
	}
	if err != nil {
		return fmt.Errorf("%w: signature invalid: %v", domain.ErrVerificationFailed, err)
	}
	return nil
}

func (v *Verifier) fetchChecksumFile(ctx context.Context, url string) (hashes.ChecksumFile, []byte, error) {
	raw, err := v.fetchBytes(ctx, url)
	if err != nil {
		return hashes.ChecksumFile{}, nil, fmt.Errorf("fetch checksum file: %w", err)
	}
	cf, err := hashes.ParseChecksumFile(strings.NewReader(string(raw)), "")
	if err != nil {
		return hashes.ChecksumFile{}, nil, fmt.Errorf("parse checksum file: %w", err)
	}
	return cf, raw, nil
}

func (v *Verifier) fetchBytes(ctx context.Context, url string) ([]byte, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("refusing non-https url %q", url)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d fetching %s", resp.StatusCode, url)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 8*1024*1024))
}
