package download

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/verify"
)

type VerifiedResult struct {
	Path   string
	Verify verify.Result
}

func (d *Downloader) FetchVerified(ctx context.Context, req Request, v *verify.Verifier, want domain.Checksum) (VerifiedResult, error) {
	if want.IsZero() {
		return VerifiedResult{}, fmt.Errorf("%w: refusing to download without an expected checksum", domain.ErrVerificationFailed)
	}

	partPath, err := d.Fetch(ctx, req)
	if errors.Is(err, errAlreadyComplete) {
		partPath = req.Dest + ".part"
	} else if err != nil {
		return VerifiedResult{}, err
	}

	res, verr := v.VerifyFile(partPath, want)
	if verr != nil {
		if rmErr := os.Remove(partPath); rmErr != nil {
			d.log.Warn("failed to remove unverified file", "path", partPath, "error", rmErr)
		}
		return VerifiedResult{}, verr
	}

	if err := os.Rename(partPath, req.Dest); err != nil {
		return VerifiedResult{}, fmt.Errorf("promote verified file: %w", err)
	}
	d.log.Info("download verified", "path", req.Dest, "algorithm", res.Algorithm)
	return VerifiedResult{Path: req.Dest, Verify: res}, nil
}
