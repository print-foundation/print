package mirrors

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
)

type Distro string

const (
	DistroDebian   Distro = "debian"
	DistroUbuntu   Distro = "ubuntu"
	DistroFedora   Distro = "fedora"
	DistroArch     Distro = "arch"
	DistroAlpine   Distro = "alpine"
	DistroFreeBSD  Distro = "freebsd"
	DistroOpenBSD  Distro = "openbsd"
	DistroOpenSUSE Distro = "opensuse"
	DistroNixOS    Distro = "nixos"
	DistroRocky    Distro = "rocky"
	DistroAlma     Distro = "almalinux"
	DistroOracle   Distro = "oracle"
	DistroVoid     Distro = "void"
	DistroGentoo   Distro = "gentoo"
	DistroClear    Distro = "clear"
)

const (
	DistroWindows Distro = "windows"
	DistroDarwin  Distro = "darwin"
	DistroMacOS   Distro = "macos"
)

type Mirror struct {
	Host    string
	BaseURL string
	Country string
	Sponsor string
}

type Fetcher interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

type httpFetcher struct{ c *http.Client }

func (h httpFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	if !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("refusing non-https url %q", url)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := h.c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("mirror list status %d", resp.StatusCode)
	}
	return readAll(resp.Body)
}

type Resolver struct {
	fetch   Fetcher
	curated map[Distro][]Mirror
}

func NewResolver() *Resolver {
	return &Resolver{fetch: httpFetcher{c: http.DefaultClient}, curated: defaultCurated()}
}

func (r *Resolver) WithFetcher(f Fetcher) *Resolver {
	r.fetch = f
	return r
}

func Supported(d Distro) bool {
	switch d {
	case DistroDebian, DistroUbuntu, DistroFedora, DistroArch, DistroAlpine, DistroFreeBSD, DistroOpenBSD,
		DistroOpenSUSE, DistroNixOS, DistroRocky, DistroAlma, DistroOracle, DistroVoid, DistroGentoo, DistroClear:
		return true
	default:
		return false
	}
}

func ListDistros() []Distro {
	out := []Distro{
		DistroAlpine, DistroAlma, DistroArch, DistroClear, DistroDebian, DistroFedora,
		DistroGentoo, DistroNixOS, DistroOpenBSD, DistroOpenSUSE, DistroOracle, DistroRocky,
		DistroUbuntu, DistroVoid,
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (r *Resolver) Resolve(ctx context.Context, d Distro, country string) ([]Mirror, error) {
	if !Supported(d) {
		return nil, fmt.Errorf("%w: %s cannot be built as a cloud ISO", ErrUnsupportedDistro, d)
	}
	country = strings.ToUpper(strings.TrimSpace(country))
	if live, ok := liveSources[d]; ok {
		raw, err := r.fetch.Fetch(ctx, live.url)
		if err == nil {
			if m, perr := live.parse(raw, country); perr == nil && len(m) > 0 {
				return m, nil
			}
		}
	}
	if cur, ok := r.curated[d]; ok {
		var in, out []Mirror
		for _, m := range cur {
			if m.Country == country {
				in = append(in, m)
			} else {
				out = append(out, m)
			}
		}
		sort.Slice(in, func(i, j int) bool { return in[i].Host < in[j].Host })
		sort.Slice(out, func(i, j int) bool { return out[i].Host < out[j].Host })
		all := append(in, out...)
		if len(all) > 0 {
			return all, nil
		}
	}
	return nil, fmt.Errorf("%w: no mirror for %s in %s", ErrNoMirror, d, country)
}
