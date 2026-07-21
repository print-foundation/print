//go:build linux

package disk

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/internal/logging"
)

type linuxDetector struct {
	log logging.Logger
}

func newPlatformDetector(log logging.Logger) Detector {
	return &linuxDetector{log: log}
}

type lsblkOutput struct {
	Blockdevices []lsblkDevice `json:"blockdevices"`
}

type lsblkDevice struct {
	Name       string        `json:"name"`
	Path       string        `json:"path"`
	Type       string        `json:"type"`
	Size       uint64        `json:"size"`
	Model      string        `json:"model"`
	Serial     string        `json:"serial"`
	Rota       bool          `json:"rota"`
	RM         bool          `json:"rm"`
	Mountpoint string        `json:"mountpoint"`
	PartTypeN  string        `json:"parttypename"`
	Fstype     string        `json:"fstype"`
	Label      string        `json:"label"`
	PhySec     uint32        `json:"phy-sec"`
	Children   []lsblkDevice `json:"children"`
}

func (d *linuxDetector) Detect(ctx context.Context) ([]domain.Disk, error) {
	out, err := runCommand(ctx, "lsblk", "-J", "-b", "-o",
		"NAME,PATH,TYPE,SIZE,MODEL,SERIAL,ROTA,RM,MOUNTPOINT,PARTTYPENAME,FSTYPE,LABEL,PHY-SEC")
	if err != nil {
		return nil, err
	}

	var parsed lsblkOutput
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		return nil, err
	}

	rootDev := d.rootDevice(ctx)

	var disks []domain.Disk
	for _, dev := range parsed.Blockdevices {
		if dev.Type != "disk" {
			continue
		}
		disk := domain.Disk{
			Path:        dev.Path,
			Model:       strings.TrimSpace(dev.Model),
			Serial:      strings.TrimSpace(dev.Serial),
			Size:        domain.ByteSize(dev.Size),
			LogicalSize: dev.PhySec,
			Removable:   dev.RM,
			Scheme:      domain.SchemeUnknown,
		}
		for _, child := range dev.Children {
			part := domain.Partition{
				Path:       child.Path,
				Label:      child.Label,
				Filesystem: child.Fstype,
				Size:       domain.ByteSize(child.Size),
				Mountpoint: child.Mountpoint,
			}
			if isESPName(child.PartTypeN) {
				part.Flags = append(part.Flags, "esp")
			}
			if child.Mountpoint == "/" || strings.HasPrefix(rootDev, child.Path) {
				disk.System = true
			}
			disk.Partitions = append(disk.Partitions, part)
		}
		if dev.Mountpoint == "/" || (rootDev != "" && strings.HasPrefix(rootDev, dev.Path)) {
			disk.System = true
		}
		disks = append(disks, disk)
	}
	return disks, nil
}

func (d *linuxDetector) rootDevice(ctx context.Context) string {
	out, err := runCommand(ctx, "findmnt", "-n", "-o", "SOURCE", "/")
	if err != nil {
		d.log.Debug("findmnt root failed", "error", err)
		return ""
	}
	return strings.TrimSpace(out)
}

func isESPName(partType string) bool {
	return strings.Contains(strings.ToLower(partType), "efi system")
}
