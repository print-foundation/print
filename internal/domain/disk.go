package domain

type Disk struct {
	Path string

	Model  string
	Serial string

	Size        ByteSize
	LogicalSize uint32 // logical sector size in bytes, usually 512 or 4096

	Removable bool
	System    bool

	Scheme     PartitionScheme
	Partitions []Partition
}

func (d Disk) Free() ByteSize {
	var used ByteSize
	for _, p := range d.Partitions {
		used += p.Size
	}
	if used >= d.Size {
		return 0
	}
	return d.Size - used
}

type Partition struct {
	Path string

	Number     int
	Label      string
	Filesystem string

	Start ByteSize
	Size  ByteSize

	Flags []string

	Mountpoint string
}

func (p Partition) IsESP() bool {
	for _, f := range p.Flags {
		if f == "esp" || f == "efi" {
			return true
		}
	}
	return p.Filesystem == "vfat" && p.hasFlag("boot")
}

func (p Partition) hasFlag(name string) bool {
	for _, f := range p.Flags {
		if f == name {
			return true
		}
	}
	return false
}
