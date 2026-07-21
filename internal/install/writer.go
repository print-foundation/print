package install

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/print-foundation/print/internal/domain"
)

type WriteProgress func(written, total domain.ByteSize)

type DeviceWriter interface {
	Write(ctx context.Context, imagePath, devicePath string, onProgress WriteProgress) error
	Finalize(ctx context.Context, devicePath string) error
}

func copyImage(ctx context.Context, dst io.Writer, src io.Reader, total domain.ByteSize, onProgress WriteProgress) error {
	buf := make([]byte, 4*1024*1024) // 4 MiB blocks write efficiently to most media
	var written domain.ByteSize
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %v", domain.ErrCancelled, ctx.Err())
		default:
		}
		n, readErr := src.Read(buf)
		if n > 0 {
			if _, werr := dst.Write(buf[:n]); werr != nil {
				return werr
			}
			written += domain.ByteSize(n)
			if onProgress != nil {
				onProgress(written, total)
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return readErr
		}
	}
}

type FileDeviceWriter struct{}

func (FileDeviceWriter) Write(ctx context.Context, imagePath, devicePath string, onProgress WriteProgress) error {
	src, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer src.Close()

	info, err := src.Stat()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(devicePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer dst.Close()

	if err := copyImage(ctx, dst, src, domain.ByteSize(info.Size()), onProgress); err != nil {
		return err
	}
	return dst.Sync()
}

func (FileDeviceWriter) Finalize(_ context.Context, _ string) error {
	return nil
}
