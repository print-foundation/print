package domain

import (
	"fmt"
	"strings"
)

type ByteSize uint64

const (
	Byte     ByteSize = 1
	Kilobyte          = 1000 * Byte
	Megabyte          = 1000 * Kilobyte
	Gigabyte          = 1000 * Megabyte
	Terabyte          = 1000 * Gigabyte

	Kibibyte = 1024 * Byte
	Mebibyte = 1024 * Kibibyte
	Gibibyte = 1024 * Mebibyte
	Tebibyte = 1024 * Gibibyte
)

func (b ByteSize) String() string {
	switch {
	case b >= Tebibyte:
		return fmt.Sprintf("%.2f TiB", float64(b)/float64(Tebibyte))
	case b >= Gibibyte:
		return fmt.Sprintf("%.2f GiB", float64(b)/float64(Gibibyte))
	case b >= Mebibyte:
		return fmt.Sprintf("%.2f MiB", float64(b)/float64(Mebibyte))
	case b >= Kibibyte:
		return fmt.Sprintf("%.2f KiB", float64(b)/float64(Kibibyte))
	default:
		return fmt.Sprintf("%d B", uint64(b))
	}
}

type Architecture string

const (
	ArchAMD64   Architecture = "amd64"
	ArchARM64   Architecture = "arm64"
	ArchRISCV64 Architecture = "riscv64"
	ArchUnknown Architecture = "unknown"
)

func NormalizeArch(s string) Architecture {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "amd64", "x86_64", "x64", "x86-64":
		return ArchAMD64
	case "arm64", "aarch64":
		return ArchARM64
	case "riscv64", "riscv":
		return ArchRISCV64
	default:
		return ArchUnknown
	}
}

type Firmware string

const (
	FirmwareUEFI    Firmware = "uefi"
	FirmwareBIOS    Firmware = "bios"
	FirmwareUnknown Firmware = "unknown"
)

type PartitionScheme string

const (
	SchemeGPT     PartitionScheme = "gpt"
	SchemeMBR     PartitionScheme = "mbr"
	SchemeUnknown PartitionScheme = "unknown"
)

type HashAlgorithm string

const (
	SHA256 HashAlgorithm = "sha256"
	SHA512 HashAlgorithm = "sha512"
	SHA1   HashAlgorithm = "sha1" // some older publishers still only ship sha1
)
