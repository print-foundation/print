//go:build darwin

package disk

import (
	"context"
	"strconv"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type darwinDetector struct {
	log logging.Logger
}

func newPlatformDetector(log logging.Logger) Detector {
	return &darwinDetector{log: log}
}

func (d *darwinDetector) Detect(ctx context.Context) ([]domain.Disk, error) {
	out, err := runCommand(ctx, "diskutil", "list")
	if err != nil {
		return nil, err
	}

	var disks []domain.Disk
	for _, block := range strings.Split(out, "\n\n") {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) == 0 || !strings.HasPrefix(lines[0], "/dev/disk") {
			continue
		}
		devPath := strings.Fields(lines[0])[0]
		disk := domain.Disk{
			Path:      devPath,
			Removable: strings.Contains(lines[0], "external"),
			Scheme:    domain.SchemeUnknown,
		}
		if size, err := d.diskSize(ctx, devPath); err == nil {
			disk.Size = size
		}
		if strings.HasSuffix(devPath, "disk0") {
			disk.System = true
		}
		disks = append(disks, disk)
	}
	return disks, nil
}

func (d *darwinDetector) diskSize(ctx context.Context, dev string) (domain.ByteSize, error) {
	out, err := runCommand(ctx, "diskutil", "info", dev)
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(out, "\n") {
		if !strings.Contains(line, "Disk Size") {
			continue
		}
		open := strings.Index(line, "(")
		if open < 0 {
			continue
		}
		rest := line[open+1:]
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		if n, err := strconv.ParseUint(fields[0], 10, 64); err == nil {
			return domain.ByteSize(n), nil
		}
	}
	return 0, nil
}
