package domain

type Hardware struct {
	Hostname     string
	Firmware     Firmware
	Architecture Architecture

	SecureBoot bool

	CPU    CPU
	Memory Memory

	Disks    []Disk
	Networks []NetworkInterface
}

type CPU struct {
	Vendor       string
	Model        string
	Cores        int
	Threads      int
	Architecture Architecture
}

type Memory struct {
	Total ByteSize
}
