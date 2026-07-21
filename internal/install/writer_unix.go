//go:build !windows

package install

import (
	"context"
	"fmt"
	"os"

	"github.com/print-foundation/print/internal/domain"
)

func NewDeviceWriter() DeviceWriter {
	return &unixDeviceWriter{}
}

type unixDeviceWriter struct{}

func (w *unixDeviceWriter) Write(ctx context.Context, imagePath, devicePath string, onProgress WriteProgress) error {
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
		return fmt.Errorf("open device %s (need root, unmounted): %w", devicePath, err)
	}
	defer dst.Close()

	if err := copyImage(ctx, dst, src, domain.ByteSize(info.Size()), onProgress); err != nil {
		return err
	}
	return dst.Sync()
}

func (w *unixDeviceWriter) Finalize(_ context.Context, devicePath string) error {
	f, err := os.OpenFile(devicePath, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	return f.Sync()
}
