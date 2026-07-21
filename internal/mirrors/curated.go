package mirrors

func defaultCurated() map[Distro][]Mirror {
	return map[Distro][]Mirror{
		DistroDebian: {
			{Host: "deb.debian.org", BaseURL: "https://deb.debian.org/debian", Country: "", Sponsor: "Debian"},
		},
		DistroUbuntu: {
			{Host: "archive.ubuntu.com", BaseURL: "https://archive.ubuntu.com/ubuntu", Country: "", Sponsor: "Ubuntu"},
		},
		DistroFedora: {
			{Host: "download.fedoraproject.org", BaseURL: "https://download.fedoraproject.org/pub/fedora", Country: "", Sponsor: "Fedora"},
		},
		DistroArch: {
			{Host: "mirror.rackspace.com", BaseURL: "https://mirror.rackspace.com/archlinux", Country: "US", Sponsor: "Rackspace"},
			{Host: "de.arch.niran.org", BaseURL: "https://de.arch.niran.org", Country: "DE", Sponsor: "niran.org"},
			{Host: "mirror.uk.domainwood.net", BaseURL: "https://mirror.uk.domainwood.net/archlinux", Country: "GB", Sponsor: "Domainwood"},
		},
		DistroAlpine: {
			{Host: "dl-cdn.alpinelinux.org", BaseURL: "https://dl-cdn.alpinelinux.org/alpine", Country: "", Sponsor: "Alpine"},
			{Host: "mirror.rackspace.com", BaseURL: "https://mirror.rackspace.com/alpine", Country: "US", Sponsor: "Rackspace"},
			{Host: "de.alpinelinux.org", BaseURL: "https://de.alpinelinux.org/alpine", Country: "DE", Sponsor: "Alpine DE"},
		},
		DistroOpenSUSE: {
			{Host: "download.opensuse.org", BaseURL: "https://download.opensuse.org", Country: "", Sponsor: "OpenSUSE"},
			{Host: "mirror.rackspace.com", BaseURL: "https://mirror.rackspace.com/opensuse", Country: "US", Sponsor: "Rackspace"},
		},
		DistroNixOS: {
			{Host: "releases.nixos.org", BaseURL: "https://releases.nixos.org", Country: "", Sponsor: "NixOS"},
		},
		DistroRocky: {
			{Host: "dl.rockylinux.org", BaseURL: "https://dl.rockylinux.org", Country: "", Sponsor: "Rocky"},
			{Host: "mirror.rackspace.com", BaseURL: "https://mirror.rackspace.com/rocky-linux", Country: "US", Sponsor: "Rackspace"},
		},
		DistroAlma: {
			{Host: "mirrors.almalinux.org", BaseURL: "https://mirrors.almalinux.org", Country: "", Sponsor: "AlmaLinux"},
			{Host: "mirror.rackspace.com", BaseURL: "https://mirror.rackspace.com/almalinux", Country: "US", Sponsor: "Rackspace"},
		},
		DistroOracle: {
			{Host: "yum.oracle.com", BaseURL: "https://yum.oracle.com", Country: "", Sponsor: "Oracle"},
		},
		DistroVoid: {
			{Host: "alpha.de.repo.voidlinux.org", BaseURL: "https://alpha.de.repo.voidlinux.org", Country: "DE", Sponsor: "Void"},
			{Host: "alpha.us.repo.voidlinux.org", BaseURL: "https://alpha.us.repo.voidlinux.org", Country: "US", Sponsor: "Void"},
		},
		DistroGentoo: {
			{Host: "gentoo.osuosl.org", BaseURL: "https://gentoo.osuosl.org", Country: "", Sponsor: "Gentoo USOSL"},
			{Host: "mirror.rackspace.com", BaseURL: "https://mirror.rackspace.com/gentoo", Country: "US", Sponsor: "Rackspace"},
		},
		DistroClear: {
			{Host: "download.clearlinux.org", BaseURL: "https://download.clearlinux.org", Country: "", Sponsor: "Intel"},
		},
		DistroFreeBSD: {
			{Host: "download.freebsd.org", BaseURL: "https://download.freebsd.org", Country: "", Sponsor: "FreeBSD"},
			{Host: "mirror.rackspace.com", BaseURL: "https://mirror.rackspace.com/freebsd", Country: "US", Sponsor: "Rackspace"},
		},
		DistroOpenBSD: {
			{Host: "cdn.openbsd.org", BaseURL: "https://cdn.openbsd.org/pub/OpenBSD", Country: "", Sponsor: "OpenBSD"},
			{Host: "ftp.eu.openbsd.org", BaseURL: "https://ftp.eu.openbsd.org/pub/OpenBSD", Country: "EU", Sponsor: "OpenBSD EU"},
		},
	}
}
