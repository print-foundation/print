package disk

import (
	"fmt"

	"github.com/print-foundation/print/internal/domain"
)

type PlanKind string

const (
	PlanWholeDisk PlanKind = "whole_disk"
)

type PartitionSpec struct {
	Role       PartitionRole
	Label      string
	Filesystem string
	Size       domain.ByteSize // zero means "fill the rest"
	Flags      []string
}

type PartitionRole string

const (
	RoleESP  PartitionRole = "esp"
	RoleBoot PartitionRole = "boot"
	RoleRoot PartitionRole = "root"
	RoleSwap PartitionRole = "swap"
)

type Plan struct {
	Disk        domain.Disk
	Scheme      domain.PartitionScheme
	Partitions  []PartitionSpec
	Destructive bool
}

type PlanOptions struct {
	Kind     PlanKind
	Firmware domain.Firmware
	SwapSize domain.ByteSize
	ESPSize  domain.ByteSize
}

const (
	defaultESPSize = 512 * domain.Mebibyte
	minRootSize    = 8 * domain.Gibibyte
)

func PlanWholeDiskLayout(target domain.Disk, opts PlanOptions) (Plan, error) {
	if target.System {
		return Plan{}, fmt.Errorf("refusing to plan over the running system disk %s", target.Path)
	}

	scheme := domain.SchemeGPT
	if opts.Firmware == domain.FirmwareBIOS {
		scheme = domain.SchemeMBR
	}

	espSize := opts.ESPSize
	if espSize == 0 {
		espSize = defaultESPSize
	}

	overhead := opts.SwapSize
	if opts.Firmware != domain.FirmwareBIOS {
		overhead += espSize
	}
	if target.Size <= overhead+minRootSize {
		return Plan{}, fmt.Errorf("%w: disk %s needs at least %s free for root",
			domain.ErrInsufficientSpace, target.Path, (overhead + minRootSize).String())
	}

	var parts []PartitionSpec
	switch scheme {
	case domain.SchemeGPT:
		parts = append(parts, PartitionSpec{
			Role:       RoleESP,
			Label:      "ESP",
			Filesystem: "vfat",
			Size:       espSize,
			Flags:      []string{"esp", "boot"},
		})
	case domain.SchemeMBR:
	}

	if opts.SwapSize > 0 {
		parts = append(parts, PartitionSpec{
			Role:       RoleSwap,
			Label:      "swap",
			Filesystem: "swap",
			Size:       opts.SwapSize,
		})
	}

	rootFlags := []string{}
	if scheme == domain.SchemeMBR {
		rootFlags = append(rootFlags, "boot")
	}
	parts = append(parts, PartitionSpec{
		Role:       RoleRoot,
		Label:      "root",
		Filesystem: "ext4",
		Size:       0, // fill remaining
		Flags:      rootFlags,
	})

	return Plan{
		Disk:        target,
		Scheme:      scheme,
		Partitions:  parts,
		Destructive: true,
	}, nil
}

func (p Plan) Summary() string {
	out := fmt.Sprintf("Target: %s (%s)\nScheme: %s\n",
		p.Disk.Path, p.Disk.Size.String(), p.Scheme)
	if p.Destructive {
		out += "WARNING: all existing data on this disk will be erased.\n"
	}
	out += "Partitions:\n"
	for _, part := range p.Partitions {
		size := "remaining space"
		if part.Size > 0 {
			size = part.Size.String()
		}
		out += fmt.Sprintf("  - %-5s %-6s %-6s %s\n", part.Role, part.Filesystem, sizeShort(size), part.Label)
	}
	return out
}

func sizeShort(s string) string {
	if len(s) > 6 {
		return s[:6]
	}
	return s
}
