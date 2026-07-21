//go:build !linux && !darwin && !windows

package hardware

import (
	"context"
	"os"
	"runtime"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type fallbackDetector struct {
	baseDetector
}

func newPlatformDetector(log logging.Logger) Detector {
	return &fallbackDetector{baseDetector{log: log}}
}

func (d *fallbackDetector) Detect(ctx context.Context) (domain.Hardware, error) {
	hw := domain.Hardware{
		Architecture: d.architecture(),
		Firmware:     domain.FirmwareUnknown,
		CPU: domain.CPU{
			Architecture: d.architecture(),
			Threads:      runtime.NumCPU(),
			Cores:        runtime.NumCPU(),
		},
		Networks: d.networkInterfaces(),
	}
	if name, err := os.Hostname(); err == nil {
		hw.Hostname = name
	}
	return hw, nil
}
