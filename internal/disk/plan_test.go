package disk

import (
	"testing"

	"github.com/print-foundation/print/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPlanWholeDiskUEFI(t *testing.T) {
	target := domain.Disk{Path: "/dev/sdb", Size: 256 * domain.Gibibyte}
	plan, err := PlanWholeDiskLayout(target, PlanOptions{Firmware: domain.FirmwareUEFI})
	require.NoError(t, err)

	assert.Equal(t, domain.SchemeGPT, plan.Scheme)
	assert.True(t, plan.Destructive)
	require.Len(t, plan.Partitions, 2)
	assert.Equal(t, RoleESP, plan.Partitions[0].Role)
	assert.Equal(t, RoleRoot, plan.Partitions[1].Role)
	assert.Equal(t, domain.ByteSize(0), plan.Partitions[1].Size, "root fills the rest")
}

func TestPlanWholeDiskBIOS(t *testing.T) {
	target := domain.Disk{Path: "/dev/sdb", Size: 128 * domain.Gibibyte}
	plan, err := PlanWholeDiskLayout(target, PlanOptions{Firmware: domain.FirmwareBIOS})
	require.NoError(t, err)

	assert.Equal(t, domain.SchemeMBR, plan.Scheme)
	require.Len(t, plan.Partitions, 1)
	assert.Equal(t, RoleRoot, plan.Partitions[0].Role)
	assert.Contains(t, plan.Partitions[0].Flags, "boot")
}

func TestPlanWithSwap(t *testing.T) {
	target := domain.Disk{Path: "/dev/sdb", Size: 256 * domain.Gibibyte}
	plan, err := PlanWholeDiskLayout(target, PlanOptions{
		Firmware: domain.FirmwareUEFI,
		SwapSize: 8 * domain.Gibibyte,
	})
	require.NoError(t, err)
	require.Len(t, plan.Partitions, 3)
	assert.Equal(t, RoleSwap, plan.Partitions[1].Role)
	assert.Equal(t, 8*domain.Gibibyte, plan.Partitions[1].Size)
}

func TestPlanRefusesSystemDisk(t *testing.T) {
	target := domain.Disk{Path: "/dev/sda", Size: 256 * domain.Gibibyte, System: true}
	_, err := PlanWholeDiskLayout(target, PlanOptions{Firmware: domain.FirmwareUEFI})
	require.Error(t, err)
}

func TestPlanRefusesTinyDisk(t *testing.T) {
	target := domain.Disk{Path: "/dev/sdb", Size: 2 * domain.Gibibyte}
	_, err := PlanWholeDiskLayout(target, PlanOptions{Firmware: domain.FirmwareUEFI})
	require.ErrorIs(t, err, domain.ErrInsufficientSpace)
}

func TestPlanSummaryMentionsErase(t *testing.T) {
	target := domain.Disk{Path: "/dev/sdb", Size: 256 * domain.Gibibyte}
	plan, err := PlanWholeDiskLayout(target, PlanOptions{Firmware: domain.FirmwareUEFI})
	require.NoError(t, err)
	assert.Contains(t, plan.Summary(), "erased")
	assert.Contains(t, plan.Summary(), "/dev/sdb")
}

func TestGuardRequiresMatchingConfirmation(t *testing.T) {
	target := domain.Disk{Path: "/dev/sdb"}

	assert.ErrorIs(t, Guard(target, Confirmation{}), domain.ErrConfirmationRequired)
	assert.ErrorIs(t, Guard(target, Confirmation{DevicePath: "/dev/sdb", Acknowledged: false}), domain.ErrConfirmationRequired)
	assert.ErrorIs(t, Guard(target, Confirmation{DevicePath: "/dev/sdc", Acknowledged: true}), domain.ErrConfirmationRequired)

	require.NoError(t, Guard(target, Confirmation{DevicePath: "/dev/sdb", Acknowledged: true}))
}

func TestGuardNeverAllowsSystemDisk(t *testing.T) {
	target := domain.Disk{Path: "/dev/sda", System: true}
	err := Guard(target, Confirmation{DevicePath: "/dev/sda", Acknowledged: true})
	assert.ErrorIs(t, err, domain.ErrConfirmationRequired)
}

func TestRecommendTarget(t *testing.T) {
	disks := []domain.Disk{
		{Path: "/dev/sda", Size: 256 * domain.Gibibyte, System: true},
		{Path: "/dev/sdb", Size: 512 * domain.Gibibyte},
		{Path: "/dev/sdc", Size: 1 * domain.Tebibyte, Removable: true},
		{Path: "/dev/sdd", Size: 128 * domain.Gibibyte},
	}
	best, ok := RecommendTarget(disks)
	require.True(t, ok)
	assert.Equal(t, "/dev/sdb", best.Path)
}

func TestRecommendTargetNoneSuitable(t *testing.T) {
	disks := []domain.Disk{
		{Path: "/dev/sda", System: true},
		{Path: "/dev/sdc", Removable: true},
	}
	_, ok := RecommendTarget(disks)
	assert.False(t, ok)
}
