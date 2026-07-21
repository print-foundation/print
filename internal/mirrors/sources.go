package mirrors

import (
	"regexp"
	"strings"
)

type liveSource struct {
	url   string
	parse func(raw []byte, country string) ([]Mirror, error)
}

var liveSources = map[Distro]liveSource{
	DistroDebian: {
		url:   "https://www.debian.org/mirror/list-full",
		parse: parseDebianList,
	},
	DistroUbuntu: {
		url:   "https://mirrors.ubuntu.com/mirrors.txt",
		parse: parseUbuntuMirrorsTxt,
	},
	DistroFedora: {
		url:   "https://mirrors.fedoraproject.org/mirrorlist?repo=fedora-40&arch=x86_64",
		parse: parseFedoraMirrorlist,
	},
}

func parseDebianList(raw []byte, country string) ([]Mirror, error) {
	text := string(raw)
	blocks := strings.Split(text, "\n\n")
	var out []Mirror
	reSite := regexp.MustCompile(`(?m)^Site:\s*(\S+)`)
	reCountry := regexp.MustCompile(`(?m)^Country:\s*([A-Z]{2})`)
	reHTTP := regexp.MustCompile(`(?m)^Archive-http:\s*(\S+)`)
	reSponsor := regexp.MustCompile(`(?m)^Sponsor:\s*(.+)`)

	for _, b := range blocks {
		site := firstGroup(reSite, b)
		if site == "" {
			continue
		}
		httpPath := firstGroup(reHTTP, b)
		if httpPath == "" {
			continue
		}
		cc := firstGroup(reCountry, b)
		sponsor := firstGroup(reSponsor, b)
		base := "https://" + site + httpPath
		m := Mirror{Host: site, BaseURL: strings.TrimRight(base, "/"), Country: cc, Sponsor: sponsor}
		if country == "" || cc == country {
			out = append(out, m)
		}
	}
	return out, nil
}

func parseUbuntuMirrorsTxt(raw []byte, _ string) ([]Mirror, error) {
	var out []Mirror
	for _, line := range strings.Split(string(raw), "\n") {
		u := strings.TrimSpace(line)
		if !strings.HasPrefix(u, "https://") {
			continue
		}
		host := hostOf(u)
		if host == "" {
			continue
		}
		out = append(out, Mirror{Host: host, BaseURL: strings.TrimRight(u, "/"), Country: ""})
	}
	return out, nil
}

func parseFedoraMirrorlist(raw []byte, _ string) ([]Mirror, error) {
	var out []Mirror
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		u := line
		if i := strings.Index(line, "https://"); i > 0 {
			u = line[i:]
		}
		if !strings.HasPrefix(u, "https://") {
			continue
		}
		host := hostOf(u)
		out = append(out, Mirror{Host: host, BaseURL: strings.TrimRight(u, "/"), Country: ""})
	}
	return out, nil
}

func firstGroup(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}

func hostOf(u string) string {
	u = strings.TrimPrefix(u, "https://")
	u = strings.TrimPrefix(u, "http://")
	if i := strings.Index(u, "/"); i >= 0 {
		u = u[:i]
	}
	return u
}
