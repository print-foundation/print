package install

import (
	"context"
	"os"
	"path/filepath"

	"github.com/print-foundation/print/internal/disk"
	"github.com/print-foundation/print/internal/domain"
	"github.com/print-foundation/print/pkg/osdb"
)

type Recovery struct {
	ImagePath  string
	Release    osdb.Release
	TargetPath string
	Complete   bool
}

func RecoveryScan(cacheDir string, releases []osdb.Release) []Recovery {
	var out []Recovery
	dir := filepath.Join(cacheDir, "images")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out
	}
	byFile := make(map[string]osdb.Release, len(releases))
	for _, r := range releases {
		byFile[cacheNameFor(r)] = r
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		rel, ok := byFile[e.Name()]
		if !ok {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Recovery{
			ImagePath:  filepath.Join(dir, e.Name()),
			Release:    rel,
			TargetPath: "",
			Complete:   domain.ByteSize(info.Size()) == domain.ByteSize(rel.Size) && rel.Size > 0,
		})
	}
	return out
}

func (e *Engine) Resume(ctx context.Context, rec Recovery, target domain.Disk, confirm disk.Confirmation, onEvent EventFunc) error {
	req := Request{
		Mode:     ModeAdvanced,
		Release:  rec.Release,
		Target:   target,
		Confirm:  confirm,
		CacheDir: filepath.Dir(filepath.Dir(rec.ImagePath)),
	}
	return e.Run(ctx, req, onEvent)
}

func cacheNameFor(rel osdb.Release) string {
	return unsafeChars.ReplaceAllString(rel.ID, "_") + extensionFor(rel.Format)
}
