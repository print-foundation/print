//go:build windows

package hardware

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type windowsDetector struct {
	baseDetector
}

func newPlatformDetector(log logging.Logger) Detector {
	return &windowsDetector{baseDetector{log: log}}
}

func (d *windowsDetector) Detect(ctx context.Context) (domain.Hardware, error) {
	hw := domain.Hardware{
		Architecture: d.architecture(),
		Firmware:     d.firmware(ctx),
		SecureBoot:   d.secureBoot(ctx),
		CPU:          d.cpu(ctx),
		Memory:       d.memory(ctx),
		Networks:     d.networkInterfaces(),
	}
	if name, err := os.Hostname(); err == nil {
		hw.Hostname = name
	}
	return hw, nil
}

func (d *windowsDetector) firmware(ctx context.Context) domain.Firmware {
	out, err := runCommand(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"(Get-ComputerInfo -Property BiosFirmwareType).BiosFirmwareType")
	if err != nil {
		d.log.Warn("detect firmware", "error", err)
		return domain.FirmwareUnknown
	}
	switch strings.TrimSpace(out) {
	case "Uefi", "2":
		return domain.FirmwareUEFI
	case "Bios", "1":
		return domain.FirmwareBIOS
	default:
		return domain.FirmwareUnknown
	}
}

func (d *windowsDetector) secureBoot(ctx context.Context) bool {
	out, err := runCommand(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"try { Confirm-SecureBootUEFI } catch { 'False' }")
	if err != nil {
		return false
	}
	return strings.EqualFold(strings.TrimSpace(out), "True")
}

type cimProcessor struct {
	Manufacturer              string `json:"Manufacturer"`
	Name                      string `json:"Name"`
	NumberOfCores             int    `json:"NumberOfCores"`
	NumberOfLogicalProcessors int    `json:"NumberOfLogicalProcessors"`
}

func (d *windowsDetector) cpu(ctx context.Context) domain.CPU {
	cpu := domain.CPU{Architecture: d.architecture()}
	out, err := runCommand(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"Get-CimInstance Win32_Processor | Select-Object Manufacturer,Name,NumberOfCores,NumberOfLogicalProcessors | ConvertTo-Json -Compress")
	if err != nil {
		d.log.Warn("detect cpu", "error", err)
		return cpu
	}

	var procs []cimProcessor
	if err := unmarshalArrayOrObject([]byte(out), &procs); err != nil {
		d.log.Warn("parse cpu json", "error", err)
		return cpu
	}
	for _, p := range procs {
		if cpu.Vendor == "" {
			cpu.Vendor = strings.TrimSpace(p.Manufacturer)
		}
		if cpu.Model == "" {
			cpu.Model = strings.TrimSpace(p.Name)
		}
		cpu.Cores += p.NumberOfCores
		cpu.Threads += p.NumberOfLogicalProcessors
	}
	if cpu.Threads == 0 {
		cpu.Threads = runtime.NumCPU()
	}
	if cpu.Cores == 0 {
		cpu.Cores = cpu.Threads
	}
	return cpu
}

func (d *windowsDetector) memory(ctx context.Context) domain.Memory {
	out, err := runCommand(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory")
	if err != nil {
		d.log.Warn("detect memory", "error", err)
		return domain.Memory{}
	}
	n := parseUint(strings.TrimSpace(out))
	return domain.Memory{Total: domain.ByteSize(n)}
}

func unmarshalArrayOrObject(data []byte, out *[]cimProcessor) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "[") {
		return json.Unmarshal(data, out)
	}
	var single cimProcessor
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*out = append(*out, single)
	return nil
}

func parseUint(s string) uint64 {
	var n uint64
	for _, r := range s {
		if r < '0' || r > '9' {
			break
		}
		n = n*10 + uint64(r-'0')
	}
	return n
}
