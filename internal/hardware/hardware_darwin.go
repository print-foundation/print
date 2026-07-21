//go:build darwin

package hardware

import (
	"context"
	"os"
	"strconv"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type darwinDetector struct {
	baseDetector
}

func newPlatformDetector(log logging.Logger) Detector {
	return &darwinDetector{baseDetector{log: log}}
}

func (d *darwinDetector) Detect(ctx context.Context) (domain.Hardware, error) {
	hw := domain.Hardware{
		Architecture: d.architecture(),
		Firmware:     domain.FirmwareUnknown,
		CPU:          d.cpu(ctx),
		Memory:       d.memory(ctx),
		Networks:     d.networkInterfaces(),
	}
	if name, err := os.Hostname(); err == nil {
		hw.Hostname = name
	}
	return hw, nil
}

func (d *darwinDetector) cpu(ctx context.Context) domain.CPU {
	cpu := domain.CPU{Architecture: d.architecture()}
	if v, err := runCommand(ctx, "sysctl", "-n", "machdep.cpu.brand_string"); err == nil {
		cpu.Model = v
	}
	if v, err := runCommand(ctx, "sysctl", "-n", "machdep.cpu.vendor"); err == nil {
		cpu.Vendor = v
	}
	if v, err := runCommand(ctx, "sysctl", "-n", "hw.physicalcpu"); err == nil {
		cpu.Cores = atoi(v)
	}
	if v, err := runCommand(ctx, "sysctl", "-n", "hw.logicalcpu"); err == nil {
		cpu.Threads = atoi(v)
	}
	if cpu.Vendor == "" {
		cpu.Vendor = "Apple"
	}
	return cpu
}

func (d *darwinDetector) memory(ctx context.Context) domain.Memory {
	v, err := runCommand(ctx, "sysctl", "-n", "hw.memsize")
	if err != nil {
		d.log.Warn("detect memory", "error", err)
		return domain.Memory{}
	}
	n, _ := strconv.ParseUint(strings.TrimSpace(v), 10, 64)
	return domain.Memory{Total: domain.ByteSize(n)}
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
