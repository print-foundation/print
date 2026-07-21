package plugin

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/print-foundation/print/pkg/osdb"
)

type Phase string

const (
	PhaseStart      Phase = "start"
	PhaseDownloaded Phase = "downloaded"
	PhaseVerified   Phase = "verified"
	PhaseWritten    Phase = "written"
	PhaseEnd        Phase = "end"
)

type Hook interface {
	OnPhase(ctx context.Context, phase Phase, req Request) error
}

type Request struct {
	ReleaseID string
	Target    string
}

type CatalogSource interface {
	Sources(ctx context.Context) ([]osdb.Release, error)
}

type Extension struct {
	Meta    Meta
	Hook    Hook
	Sources CatalogSource
}

type Meta struct {
	ID          string
	Name        string
	Version     string
	Description string
	MinEngine   string
}

type NoOpHook struct{}

func (NoOpHook) OnPhase(_ context.Context, _ Phase, _ Request) error { return nil }

func Register(ext Extension) error {
	if ext.Meta.ID == "" {
		return fmt.Errorf("plugin: missing id")
	}
	reg.mu.Lock()
	defer reg.mu.Unlock()
	if _, dup := reg.exts[ext.Meta.ID]; dup {
		return fmt.Errorf("plugin: duplicate id %q", ext.Meta.ID)
	}
	if ext.Hook == nil {
		ext.Hook = NoOpHook{}
	}
	reg.exts[ext.Meta.ID] = ext
	return nil
}

func List() []Extension {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	out := make([]Extension, 0, len(reg.exts))
	for _, e := range reg.exts {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Meta.ID < out[j].Meta.ID })
	return out
}

func Hooks() []Hook {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	out := make([]Hook, 0, len(reg.exts))
	for _, e := range reg.exts {
		out = append(out, e.Hook)
	}
	return out
}

func AllSources() []CatalogSource {
	reg.mu.RLock()
	defer reg.mu.RUnlock()
	var out []CatalogSource
	for _, e := range reg.exts {
		if e.Sources != nil {
			out = append(out, e.Sources)
		}
	}
	return out
}

func notify(ctx context.Context, phase Phase, req Request) error {
	for _, h := range Hooks() {
		if err := h.OnPhase(ctx, phase, req); err != nil {
			return fmt.Errorf("plugin %s phase %s: %w", phase, req.ReleaseID, err)
		}
	}
	return nil
}

func RunHooks(ctx context.Context, phase Phase, req Request) error {
	return notify(ctx, phase, req)
}

var reg = struct {
	mu   sync.RWMutex
	exts map[string]Extension
}{
	exts: make(map[string]Extension),
}
