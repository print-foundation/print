//go:build !linux && !darwin && !windows

package disk

import (
	"context"
	"fmt"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type unsupportedDetector struct {
	log logging.Logger
}

func newPlatformDetector(log logging.Logger) Detector {
	return &unsupportedDetector{log: log}
}

func (d *unsupportedDetector) Detect(ctx context.Context) ([]domain.Disk, error) {
	return nil, fmt.Errorf("%w: disk detection on this platform", domain.ErrUnsupported)
}
