package builder

import (
	"github.com/print-foundation/print/internal/mirrors"
)

type Arch string

const (
	ArchAmd64 Arch = "amd64"
	ArchArm64 Arch = "arm64"
)

type Spec struct {
	Distro  mirrors.Distro
	Arch    Arch
	Mirror  mirrors.Mirror
	WiFi    WiFiConfig
	Output  string
	WorkDir string
	Tools   ToolPaths
}

type WiFiConfig struct {
	SSID       string
	Passphrase string
}

type ToolPaths struct {
	Xorriso      string
	GrubMkRescue string
	GrubMkImage  string
}

type Asset struct {
	Kernel string
	Initrd string
	Extra  []string
}

var netbootAssets = map[mirrors.Distro]map[Arch]Asset{
	mirrors.DistroDebian: {
		ArchAmd64: {Kernel: "/dists/stable/main/installer-amd64/current/images/netboot/debian-installer/amd64/linux",
			Initrd: "/dists/stable/main/installer-amd64/current/images/netboot/debian-installer/amd64/initrd.gz"},
		ArchArm64: {Kernel: "/dists/stable/main/installer-arm64/current/images/netboot/debian-installer/arm64/linux",
			Initrd: "/dists/stable/main/installer-arm64/current/images/netboot/debian-installer/arm64/initrd.gz"},
	},
	mirrors.DistroUbuntu: {
		ArchAmd64: {Kernel: "/dists/noble/main/installer-amd64/current/legacy-images/netboot/ubuntu-installer/amd64/linux",
			Initrd: "/dists/noble/main/installer-amd64/current/legacy-images/netboot/ubuntu-installer/amd64/initrd.gz"},
		ArchArm64: {Kernel: "/dists/noble/main/installer-arm64/current/legacy-images/netboot/ubuntu-installer/arm64/linux",
			Initrd: "/dists/noble/main/installer-arm64/current/legacy-images/netboot/ubuntu-installer/arm64/initrd.gz"},
	},
	mirrors.DistroFedora: {
		ArchAmd64: {Kernel: "/pub/fedora/linux/releases/40/Everything/x86_64/os/images/pxeboot/vmlinuz",
			Initrd: "/pub/fedora/linux/releases/40/Everything/x86_64/os/images/pxeboot/initrd.img"},
		ArchArm64: {Kernel: "/pub/fedora/linux/releases/40/Everything/aarch64/os/images/pxeboot/vmlinuz",
			Initrd: "/pub/fedora/linux/releases/40/Everything/aarch64/os/images/pxeboot/initrd.img"},
	},
	mirrors.DistroArch: {
		ArchAmd64: {Kernel: "/iso/latest/boot/x86_64/vmlinuz-linux",
			Initrd: "/iso/latest/boot/x86_64/initramfs-linux.img"},
		ArchArm64: {Kernel: "/iso/latest/boot/aarch64/vmlinuz-linux",
			Initrd: "/iso/latest/boot/aarch64/initramfs-linux.img"},
	},
	mirrors.DistroAlpine: {
		ArchAmd64: {Kernel: "/v3.20/releases/x86_64/netboot/vmlinuz-lts",
			Initrd: "/v3.20/releases/x86_64/netboot/initramfs-lts"},
		ArchArm64: {Kernel: "/v3.20/releases/aarch64/netboot/vmlinuz-lts",
			Initrd: "/v3.20/releases/aarch64/netboot/initramfs-lts"},
	},
	mirrors.DistroFreeBSD: {
		ArchAmd64: {Kernel: "/releases/amd64/amd64/14.2-RELEASE/base.txz", Initrd: ""},
		ArchArm64: {Kernel: "/releases/arm64/arm64/14.2-RELEASE/base.txz", Initrd: ""},
	},
	mirrors.DistroOpenBSD: {
		ArchAmd64: {Kernel: "/amd64/bsd.rd", Initrd: ""},
		ArchArm64: {Kernel: "/arm64/bsd.rd", Initrd: ""},
	},
	mirrors.DistroOpenSUSE: {
		ArchAmd64: {Kernel: "/tumbleweed/repo/oss/boot/x86_64/loader/linux",
			Initrd: "/tumbleweed/repo/oss/boot/x86_64/loader/initrd"},
		ArchArm64: {Kernel: "/tumbleweed/repo/oss/boot/aarch64/loader/linux",
			Initrd: "/tumbleweed/repo/oss/boot/aarch64/loader/initrd"},
	},
	mirrors.DistroNixOS: {
		ArchAmd64: {Kernel: "/nixos/24.05/nixos-24.05-x86_64-linux/netboot/vmlinuz",
			Initrd: "/nixos/24.05/nixos-24.05-x86_64-linux/netboot/initrd"},
		ArchArm64: {Kernel: "/nixos/24.05/nixos-24.05-aarch64-linux/netboot/vmlinuz",
			Initrd: "/nixos/24.05/nixos-24.05-aarch64-linux/netboot/initrd"},
	},
	mirrors.DistroRocky: {
		ArchAmd64: {Kernel: "/pub/rocky/9/BaseOS/x86_64/kickstart/images/vmlinuz",
			Initrd: "/pub/rocky/9/BaseOS/x86_64/kickstart/images/initrd.img"},
		ArchArm64: {Kernel: "/pub/rocky/9/BaseOS/aarch64/kickstart/images/vmlinuz",
			Initrd: "/pub/rocky/9/BaseOS/aarch64/kickstart/images/initrd.img"},
	},
	mirrors.DistroAlma: {
		ArchAmd64: {Kernel: "/almalinux/9/BaseOS/x86_64/kickstart/images/vmlinuz",
			Initrd: "/almalinux/9/BaseOS/x86_64/kickstart/images/initrd.img"},
		ArchArm64: {Kernel: "/almalinux/9/BaseOS/aarch64/kickstart/images/vmlinuz",
			Initrd: "/almalinux/9/BaseOS/aarch64/kickstart/images/initrd.img"},
	},
	mirrors.DistroOracle: {
		ArchAmd64: {Kernel: "/yun/9/ol9_baseos_latest/x86_64/boot/loader/vmlinuz",
			Initrd: "/yun/9/ol9_baseos_latest/x86_64/boot/loader/initrd.img"},
		ArchArm64: {Kernel: "/yun/9/ol9_baseos_latest/aarch64/boot/loader/vmlinuz",
			Initrd: "/yun/9/ol9_baseos_latest/aarch64/boot/loader/initrd.img"},
	},
	mirrors.DistroVoid: {
		ArchAmd64: {Kernel: "/live/current/void-current-x86_64-20240606.iso", Initrd: "", Extra: []string{"/live/current/void-current-x86_64-20240606.iso"}},
		ArchArm64: {Kernel: "/live/current/void-current-aarch64-20240606.iso", Initrd: "", Extra: []string{"/live/current/void-current-aarch64-20240606.iso"}},
	},
	mirrors.DistroGentoo: {
		ArchAmd64: {Kernel: "/releases/amd64/autobuilds/current-install-amd64-minimal/install-amd64-minimal-latest.iso", Initrd: ""},
		ArchArm64: {Kernel: "/releases/arm64/autobuilds/current-install-arm64-minimal/install-arm64-minimal-latest.iso", Initrd: ""},
	},
	mirrors.DistroClear: {
		ArchAmd64: {Kernel: "/releases/39700/clear/linux-4.19.292-1398.tar.xz", Initrd: ""},
		ArchArm64: {Kernel: "/releases/39700/clear/linux-aarch64-4.19.292-1398.tar.xz", Initrd: ""},
	},
}

func AssetFor(d mirrors.Distro, a Arch) (Asset, bool) {
	m, ok := netbootAssets[d]
	if !ok {
		return Asset{}, false
	}
	a2, ok := m[a]
	return a2, ok
}
