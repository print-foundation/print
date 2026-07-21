//go:build windows

package disk

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type windowsDetector struct {
	log logging.Logger
}

func newPlatformDetector(log logging.Logger) Detector {
	return &windowsDetector{log: log}
}

type winDisk struct {
	Number             int    `json:"Number"`
	FriendlyName       string `json:"FriendlyName"`
	SerialNumber       string `json:"SerialNumber"`
	Size               uint64 `json:"Size"`
	BusType            int    `json:"BusType"`
	IsBoot             bool   `json:"IsBoot"`
	IsSystem           bool   `json:"IsSystem"`
	PartitionStyle     string
	LogicalSectorB     uint32 `json:"LogicalSectorSize"`
	NumberOfPartitions int    `json:"NumberOfPartitions"`
}

func (d *windowsDetector) Detect(ctx context.Context) ([]domain.Disk, error) {
	out, err := runCommand(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"Get-Disk | Select-Object Number,FriendlyName,SerialNumber,Size,BusType,IsBoot,IsSystem,PartitionStyle,LogicalSectorSize,NumberOfPartitions | ConvertTo-Json -Compress")
	if err != nil {
		return nil, err
	}

	var raw []winDisk
	if err := decodeJSONArrayOrObject([]byte(out), &raw); err != nil {
		return nil, err
	}

	disks := make([]domain.Disk, 0, len(raw))
	for _, wd := range raw {
		disk := domain.Disk{
			Path:        physicalDrivePath(wd.Number),
			Model:       strings.TrimSpace(wd.FriendlyName),
			Serial:      strings.TrimSpace(wd.SerialNumber),
			Size:        domain.ByteSize(wd.Size),
			LogicalSize: wd.LogicalSectorB,
			Removable:   wd.BusType == 7, // windows reports removable drives differently
			System:      wd.IsBoot || wd.IsSystem,
			Scheme:      partitionStyle(wd.PartitionStyle),
		}
		if parts, err := d.partitions(ctx, wd.Number); err != nil {
			d.log.Warn("read partitions", "disk", wd.Number, "error", err)
		} else {
			disk.Partitions = parts
		}
		disks = append(disks, disk)
	}
	return disks, nil
}

type winPartition struct {
	PartitionNumber int    `json:"PartitionNumber"`
	DriveLetter     string `json:"DriveLetter"`
	Size            uint64 `json:"Size"`
	Offset          uint64 `json:"Offset"`
	Type            string `json:"Type"`
	GptType         string `json:"GptType"`
	IsBoot          bool   `json:"IsBoot"`
	IsSystem        bool   `json:"IsSystem"`
}

func (d *windowsDetector) partitions(ctx context.Context, diskNumber int) ([]domain.Partition, error) {
	out, err := runCommand(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		"Get-Partition -DiskNumber "+itoa(diskNumber)+" | Select-Object PartitionNumber,DriveLetter,Size,Offset,Type,GptType,IsBoot,IsSystem | ConvertTo-Json -Compress")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var raw []winPartition
	if err := decodeJSONArrayOrObject([]byte(out), &raw); err != nil {
		return nil, err
	}
	parts := make([]domain.Partition, 0, len(raw))
	for _, wp := range raw {
		p := domain.Partition{
			Number:     wp.PartitionNumber,
			Start:      domain.ByteSize(wp.Offset),
			Size:       domain.ByteSize(wp.Size),
			Mountpoint: driveLetter(wp.DriveLetter),
		}
		if isESPType(wp.GptType, wp.Type) {
			p.Flags = append(p.Flags, "esp")
			p.Filesystem = "vfat"
		}
		if wp.IsBoot {
			p.Flags = append(p.Flags, "boot")
		}
		parts = append(parts, p)
	}
	return parts, nil
}

func physicalDrivePath(number int) string {
	return `\\.\PhysicalDrive` + itoa(number)
}

func partitionStyle(s string) domain.PartitionScheme {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "GPT":
		return domain.SchemeGPT
	case "MBR":
		return domain.SchemeMBR
	default:
		return domain.SchemeUnknown
	}
}

func isESPType(gptType, mbrType string) bool {
	return strings.EqualFold(strings.Trim(gptType, "{}"), "c12a7328-f81f-11d2-ba4b-00a0c93ec93b") ||
		strings.EqualFold(mbrType, "System")
}

func driveLetter(l string) string {
	l = strings.TrimSpace(l)
	if l == "" {
		return ""
	}
	return l + ":"
}

func decodeJSONArrayOrObject[T any](data []byte, out *[]T) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "" {
		return nil
	}
	if strings.HasPrefix(trimmed, "[") {
		return json.Unmarshal(data, out)
	}
	var single T
	if err := json.Unmarshal(data, &single); err != nil {
		return err
	}
	*out = append(*out, single)
	return nil
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
