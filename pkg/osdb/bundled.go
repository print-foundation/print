package osdb

import (
	"bytes"
	_ "embed"
)

//go:embed catalog.json
var bundled []byte

func Bundled() (*Catalog, error) {
	return Load(bytes.NewReader(bundled))
}

func BundledJSON() []byte {
	out := make([]byte, len(bundled))
	copy(out, bundled)
	return out
}
