//go:build linux

package hardware

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type linuxDetector struct {
	baseDetector
}

func newPlatformDetector(log logging.Logger) Detector {
	return &linuxDetector{baseDetector{log: log}}
}

func (d *linuxDetector) Detect(ctx context.Context) (domain.Hardware, error) {
	hw := domain.Hardware{
		Architecture: d.architecture(),
		Firmware:     detectFirmware(),
		SecureBoot:   detectSecureBoot(),
		CPU:          d.cpu(),
		Memory:       d.memory(),
		Networks:     d.networkInterfaces(),
	}
	if name, err := os.Hostname(); err == nil {
		hw.Hostname = name
	}
	return hw, nil
}

func detectFirmware() domain.Firmware {
	if _, err := os.Stat("/sys/firmware/efi"); err == nil {
		return domain.FirmwareUEFI
	}
	return domain.FirmwareBIOS
}

func detectSecureBoot() bool {
	matches, _ := filepath.Glob("/sys/firmware/efi/efivars/SecureBoot-*")
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			continue
		}
		if len(data) > 0 && data[len(data)-1] == 1 {
			return true
		}
	}
	return false
}

func (d *linuxDetector) cpu() domain.CPU {
	cpu := domain.CPU{Architecture: d.architecture()}
	f, err := os.Open("/proc/cpuinfo")
	if err != nil {
		d.log.Warn("read cpuinfo", "error", err)
		return cpu
	}
	defer f.Close()

	physical := map[string]bool{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		key, val, ok := splitProc(scanner.Text())
		if !ok {
			continue
		}
		switch key {
		case "vendor_id":
			cpu.Vendor = val
		case "model name":
			if cpu.Model == "" {
				cpu.Model = val
			}
		case "processor":
			cpu.Threads++
		case "physical id":
			physical[val] = true
		case "cpu cores":
			if n, err := strconv.Atoi(val); err == nil && cpu.Cores == 0 {
				cpu.Cores = n
			}
		}
	}
	if len(physical) > 1 && cpu.Cores > 0 {
		cpu.Cores *= len(physical)
	}
	if cpu.Cores == 0 {
		cpu.Cores = cpu.Threads
	}
	return cpu
}

func (d *linuxDetector) memory() domain.Memory {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		d.log.Warn("read meminfo", "error", err)
		return domain.Memory{}
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "MemTotal:") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			break
		}
		kb, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			break
		}
		return domain.Memory{Total: domain.ByteSize(kb) * domain.Kibibyte}
	}
	return domain.Memory{}
}

func splitProc(line string) (key, val string, ok bool) {
	idx := strings.IndexByte(line, ':')
	if idx < 0 {
		return "", "", false
	}
	return strings.TrimSpace(line[:idx]), strings.TrimSpace(line[idx+1:]), true
}
