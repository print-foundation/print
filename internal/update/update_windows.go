//go:build windows

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
	old := exe + ".old"
	_ = os.Remove(old)
	if err := os.Rename(exe, old); err != nil {
		return fmt.Errorf("back up current binary: %w", err)
	}
	if err := copyFile(newBinary, exe, info.Mode()); err != nil {
		_ = os.Rename(old, exe)
		return fmt.Errorf("copy new binary: %w", err)
	}
	_ = os.Remove(old)
	return nil
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, mode)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
