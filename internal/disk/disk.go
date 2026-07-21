package disk

import (
	"context"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type Detector interface {
	Detect(ctx context.Context) ([]domain.Disk, error)
}

func New(log logging.Logger) Detector {
	return newPlatformDetector(log)
}
