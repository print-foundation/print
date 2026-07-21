package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/print-foundation/print/internal/builder"
)

func generateSeed(cfg builder.ClientConfig) (string, error) {
	mirror := strings.TrimRight(cfg.Mirror, "/")
	switch cfg.Distro {
	case "debian", "ubuntu":
		return debianPreseed(mirror), nil
	case "fedora":
		return fedoraKickstart(mirror), nil
	case "arch":
		return archSeed(mirror), nil
	case "alpine":
		return alpineSeed(mirror), nil
	case "freebsd":
		return freebsdSeed(mirror), nil
	case "openbsd":
		return openbsdSeed(mirror), nil
	default:
		return "", fmt.Errorf("no seed generator for %s", cfg.Distro)
	}
}

func debianPreseed(mirror string) string {
	return fmt.Sprintf(`# Pr!nt-generated preseed for %s
d-i mirror/country string manual
d-i mirror/http/hostname string %s
d-i mirror/http/directory string /debian
d-i mirror/http/proxy string
d-i pkgsel/install-language-support boolean false
d-i pkgsel/update-policy select none
d-i finish-install/reboot_in_progress note
`, mirror, hostOf(mirror))
}

func fedoraKickstart(mirror string) string {
	return fmt.Sprintf(`# Pr!nt-generated kickstart for fedora
url --url="%s"
text
reboot
clearpart --all --initlabel
autopart
%%packages
@core
%%end
`, mirror)
}

func archSeed(mirror string) string {
	cfg := map[string]any{
		"mirror_regions": map[string]any{
			"World": []string{mirror + "/$repo/os/$arch"},
		},
		"script": "guided",
		"disk_layouts": map[string]any{
			"__default__": "wipe_and_default_layout",
		},
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Sprintf("# Pr!nt arch seed (mirror %s)\n", mirror)
	}
	return string(b)
}

func alpineSeed(mirror string) string {
	return fmt.Sprintf(`# Pr!nt-generated alpine setup answer
KEYMAP=us
HOSTNAME=print
INTERFACES=auto
MIRROR=%s
`, mirror)
}

func freebsdSeed(mirror string) string {
	return fmt.Sprintf(`# Pr!nt-generated freebsd install config
# fetch sets from: %s/releases/amd64/amd64/14.2-RELEASE
mirror=%s
`, mirror, mirror)
}

func openbsdSeed(mirror string) string {
	return fmt.Sprintf(`# Pr!nt-generated openbsd install config
# install from: %s/amd64
mirror=%s
`, mirror, mirror)
}

func hostOf(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	if i := strings.Index(u, "/"); i >= 0 {
		u = u[:i]
	}
	return u
}
