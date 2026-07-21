package disk

import (
	"github.com/print-foundation/print/internal/domain"
)

type Confirmation struct {
	DevicePath   string
	Acknowledged bool
}

func Guard(target domain.Disk, c Confirmation) error {
	if !c.Acknowledged || c.DevicePath != target.Path {
		return domain.ErrConfirmationRequired
	}
	if target.System {
		return domain.ErrConfirmationRequired
	}
	return nil
}

func RecommendTarget(disks []domain.Disk) (domain.Disk, bool) {
	var best domain.Disk
	found := false
	for _, d := range disks {
		if d.System || d.Removable {
			continue
		}
		if !found || d.Size > best.Size {
			best = d
			found = true
		}
	}
	return best, found
}
