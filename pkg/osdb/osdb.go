package osdb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
)

type Catalog struct {
	SchemaVersion int `json:"schema_version"`

	Updated string `json:"updated"`

	OperatingSystems []OperatingSystem `json:"operating_systems"`
}

type OperatingSystem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Publisher   Publisher `json:"publisher"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Releases    []Release `json:"releases"`
}

type Publisher struct {
	Name    string `json:"name"`
	Website string `json:"website"`
}

type Release struct {
	ID           string `json:"id"`
	Version      string `json:"version"`
	Edition      string `json:"edition"`
	Architecture string `json:"architecture"`
	LTS          bool   `json:"lts"`

	URL string `json:"url"`

	Size uint64 `json:"size"`

	Format string `json:"format"`

	Checksum Checksum `json:"checksum"`

	ChecksumURL string `json:"checksum_url"`

	SignatureURL string `json:"signature_url"`

	LicenseURL string `json:"license_url"`
}

type Checksum struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

func (c Checksum) IsZero() bool { return strings.TrimSpace(c.Value) == "" }

func Load(r io.Reader) (*Catalog, error) {
	var c Catalog
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&c); err != nil {
		return nil, fmt.Errorf("decode catalog: %w", err)
	}
	if err := c.Validate(); err != nil {
		return nil, err
	}
	return &c, nil
}

func (c *Catalog) Validate() error {
	if c.SchemaVersion != SchemaVersion {
		return fmt.Errorf("unsupported schema version %d (want %d)", c.SchemaVersion, SchemaVersion)
	}
	seen := map[string]bool{}
	for _, os := range c.OperatingSystems {
		if os.ID == "" {
			return fmt.Errorf("operating system with empty id")
		}
		for _, rel := range os.Releases {
			if rel.ID == "" {
				return fmt.Errorf("%s: release with empty id", os.ID)
			}
			if seen[rel.ID] {
				return fmt.Errorf("duplicate release id %q", rel.ID)
			}
			seen[rel.ID] = true
			if err := requireHTTPS(rel.URL); err != nil {
				return fmt.Errorf("%s: url: %w", rel.ID, err)
			}
			for _, optional := range []string{rel.ChecksumURL, rel.SignatureURL, rel.LicenseURL} {
				if optional == "" {
					continue
				}
				if err := requireHTTPS(optional); err != nil {
					return fmt.Errorf("%s: %w", rel.ID, err)
				}
			}
		}
	}
	return nil
}

func requireHTTPS(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid url %q", raw)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("url %q must use https", raw)
	}
	if u.Host == "" {
		return fmt.Errorf("url %q missing host", raw)
	}
	return nil
}

const SchemaVersion = 1

func (c *Catalog) FindRelease(id string) (OperatingSystem, Release, bool) {
	for _, os := range c.OperatingSystems {
		for _, rel := range os.Releases {
			if rel.ID == id {
				return os, rel, true
			}
		}
	}
	return OperatingSystem{}, Release{}, false
}

func (c *Catalog) Releases() []Release {
	var out []Release
	for _, os := range c.OperatingSystems {
		out = append(out, os.Releases...)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Version != out[j].Version {
			return out[i].Version > out[j].Version // newer first
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func (c *Catalog) FilterByArch(arch string) []OperatingSystem {
	arch = strings.ToLower(arch)
	var out []OperatingSystem
	for _, os := range c.OperatingSystems {
		var matching []Release
		for _, rel := range os.Releases {
			if strings.ToLower(rel.Architecture) == arch {
				matching = append(matching, rel)
			}
		}
		if len(matching) > 0 {
			cp := os
			cp.Releases = matching
			out = append(out, cp)
		}
	}
	return out
}
