package domain

import (
	"net/url"
	"strings"
)

type Publisher struct {
	Name    string
	Website string
}

type Checksum struct {
	Algorithm HashAlgorithm
	Value     string
}

func (c Checksum) Equal(other Checksum) bool {
	return c.Algorithm == other.Algorithm &&
		strings.EqualFold(c.Value, other.Value)
}

func (c Checksum) IsZero() bool {
	return c.Value == ""
}

type Image struct {
	ID string

	OSID string

	Name    string
	Version string
	Edition string

	Architecture Architecture
	Publisher    Publisher

	URL string

	Size ByteSize

	Checksum Checksum

	SignatureURL string

	LicenseURL string

	Kind ImageKind
}

type ImageKind string

const (
	ImageISO   ImageKind = "iso"
	ImageRaw   ImageKind = "raw"
	ImageRawXZ ImageKind = "raw.xz"
	ImageRawGZ ImageKind = "raw.gz"
)

func (i Image) Validate() error {
	if strings.TrimSpace(i.ID) == "" {
		return &ValidationError{Field: "ID", Msg: "must not be empty"}
	}
	if strings.TrimSpace(i.Name) == "" {
		return &ValidationError{Field: "Name", Msg: "must not be empty"}
	}
	if i.Architecture == ArchUnknown || i.Architecture == "" {
		return &ValidationError{Field: "Architecture", Msg: "must be known"}
	}
	u, err := url.Parse(i.URL)
	if err != nil {
		return &ValidationError{Field: "URL", Msg: "is not a valid url"}
	}
	if u.Scheme != "https" {
		return &ValidationError{Field: "URL", Msg: "must use https"}
	}
	if i.SignatureURL != "" {
		su, err := url.Parse(i.SignatureURL)
		if err != nil || su.Scheme != "https" {
			return &ValidationError{Field: "SignatureURL", Msg: "must use https"}
		}
	}
	return nil
}

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return "invalid " + e.Field + ": " + e.Msg
}
