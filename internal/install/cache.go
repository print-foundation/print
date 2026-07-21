package install

import (
	"path/filepath"
	"regexp"

	"github.com/print-foundation/print/pkg/osdb"
)

var unsafeChars = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func cachePath(cacheDir string, rel osdb.Release) string {
	name := unsafeChars.ReplaceAllString(rel.ID, "_")
	ext := extensionFor(rel.Format)
	return filepath.Join(cacheDir, "images", name+ext)
}

func extensionFor(format string) string {
	switch format {
	case "raw.xz":
		return ".raw.xz"
	case "raw.gz":
		return ".raw.gz"
	case "raw":
		return ".raw"
	default:
		return ".iso"
	}
}
