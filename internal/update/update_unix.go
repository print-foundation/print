//go:build !windows

package update

import (
	"fmt"
	"io"
	"os"
)

func replaceBinary(exe, newBinary string) error {
	info, err := os.Stat(exe)
	if err != nil {
		return err
	}
	if err := os.Rename(newBinary, exe); err != nil {
		if cpErr := copyFile(newBinary, exe, info.Mode()); cpErr != nil {
			return fmt.Errorf("replace binary: %w (copy fallback: %v)", err, cpErr)
		}
		_ = os.Remove(newBinary)
	}
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
