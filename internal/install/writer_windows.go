//go:build windows

package install

import (
	"context"
	"fmt"
	"os"

	"github.com/print-foundation/print/internal/domain"
)

func NewDeviceWriter() DeviceWriter {
	return &windowsDeviceWriter{}
}

type windowsDeviceWriter struct{}

func (w *windowsDeviceWriter) Write(ctx context.Context, imagePath, devicePath string, onProgress WriteProgress) error {
	src, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("open %s (run as administrator): %w", devicePath, err)
	}
	defer dst.Close()

	if err := copyImage(ctx, dst, src, domain.ByteSize(info.Size()), onProgress); err != nil {
		return err
	}
	return dst.Sync()
}

func (w *windowsDeviceWriter) Finalize(ctx context.Context, devicePath string) error {
	f, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
