package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestByteSizeString(t *testing.T) {
	cases := []struct {
		in   ByteSize
		want string
	}{
		{500, "500 B"},
		{2 * Kibibyte, "2.00 KiB"},
		{3 * Mebibyte, "3.00 MiB"},
		{4 * Gibibyte, "4.00 GiB"},
		{5 * Tebibyte, "5.00 TiB"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, c.in.String())
	}
}

func TestNormalizeArch(t *testing.T) {
	assert.Equal(t, ArchAMD64, NormalizeArch("x86_64"))
	assert.Equal(t, ArchAMD64, NormalizeArch("AMD64"))
	assert.Equal(t, ArchARM64, NormalizeArch("aarch64"))
	assert.Equal(t, ArchRISCV64, NormalizeArch("riscv64"))
	assert.Equal(t, ArchUnknown, NormalizeArch("sparc"))
}

func TestChecksumEqual(t *testing.T) {
	a := Checksum{Algorithm: SHA256, Value: "ABCDEF"}
	b := Checksum{Algorithm: SHA256, Value: "abcdef"}
	assert.True(t, a.Equal(b))

	c := Checksum{Algorithm: SHA512, Value: "abcdef"}
	assert.False(t, a.Equal(c))
}

func TestImageValidate(t *testing.T) {
	valid := Image{
		ID:           "ubuntu-24.04-amd64",
		Name:         "Ubuntu 24.04",
		Architecture: ArchAMD64,
		URL:          "https://releases.ubuntu.com/24.04/ubuntu.iso",
	}
	require.NoError(t, valid.Validate())

	insecure := valid
	insecure.URL = "http://releases.ubuntu.com/24.04/ubuntu.iso"
	assert.Error(t, insecure.Validate())

	noArch := valid
	noArch.Architecture = ArchUnknown
	assert.Error(t, noArch.Validate())

	empty := Image{}
	assert.Error(t, empty.Validate())
}

func TestDiskFree(t *testing.T) {
	d := Disk{
		Size: 100 * Gibibyte,
		Partitions: []Partition{
			{Size: 30 * Gibibyte},
			{Size: 20 * Gibibyte},
		},
	}
	assert.Equal(t, 50*Gibibyte, d.Free())

	over := Disk{Size: 10, Partitions: []Partition{{Size: 20}}}
	assert.Equal(t, ByteSize(0), over.Free())
}

func TestPartitionIsESP(t *testing.T) {
	esp := Partition{Flags: []string{"esp"}}
	assert.True(t, esp.IsESP())

	vfatBoot := Partition{Filesystem: "vfat", Flags: []string{"boot"}}
	assert.True(t, vfatBoot.IsESP())

	plain := Partition{Filesystem: "ext4"}
	assert.False(t, plain.IsESP())
}
