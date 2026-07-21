package install

import (
	"context"
	"fmt"
	"os"

	"github.com/print-foundation/print/internal/disk"
	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/pkg/hashes"
)

func (e *Engine) FlashLocal(ctx context.Context, imagePath, devicePath string, confirm disk.Confirmation, onProgress WriteProgress) error {
	info, err := os.Stat(imagePath)
	if err != nil {
		return fmt.Errorf("stat image: %w", err)
	}
	target := domain.Disk{Path: devicePath, Size: domain.ByteSize(info.Size())}

	if err := disk.Guard(target, confirm); err != nil {
		return err
	}

	digest, err := hashes.SumFile(hashes.SHA256, imagePath)
	if err != nil {
		return fmt.Errorf("hash image: %w", err)
	}
	e.log.Info("flashing local image", "path", imagePath, "device", devicePath, "digest", digest)

	if err := e.writer.Write(ctx, imagePath, devicePath, onProgress); err != nil {
		return fmt.Errorf("write image: %w", err)
	}
	if err := e.writer.Finalize(ctx, devicePath); err != nil {
		return fmt.Errorf("finalize: %w", err)
	}
	return nil
}
