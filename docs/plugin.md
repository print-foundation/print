# plugin.md — Writing a Pr!nt plugin

Pr!nt plugins are **compiled in**, not loaded from disk at runtime. This is a
deliberate security choice (see SECURITY.md): there is no arbitrary
shared-object execution surface. A plugin is just a Go package that registers an
`Extension` at `init`.

## Two things a plugin can do

- **Lifecycle hook** — observe or veto install phases.
- **Catalog source** — contribute extra OS releases merged with the bundled
  catalog.

## Minimal hook plugin

```go
package myplugin

import (
	"context"

	"github.com/kilo-org/print/internal/logging"
	"github.com/kilo-org/print/internal/plugin"
)

func init() {
	_ = plugin.Register(plugin.Extension{
		Meta: plugin.Meta{
			ID:          "example/notifier",
			Name:        "Notifier",
			Version:     "1.0",
			Description: "Logs each install phase.",
		},
		Hook: &hook{log: logging.NopLogger()},
	})
}

type hook struct{ log logging.Logger }

func (h *hook) OnPhase(ctx context.Context, phase plugin.Phase, req plugin.Request) error {
	h.log.Info("phase", "phase", phase, "release", req.ReleaseID)
	return nil // returning an error aborts the install
}
```

Then import the package from `cmd/print` (a blank import is enough) so its
`init` runs.

## A catalog source

```go
func init() {
	_ = plugin.Register(plugin.Extension{
		Meta: plugin.Meta{ID: "example/extra-os", Name: "Extra OS", Version: "1.0"},
		Sources: src{},
	})
}

type src struct{}

func (src) Sources(ctx context.Context) ([]osdb.Release, error) {
	return []osdb.Release{{
		ID: "example/os-9", Version: "9", URL: "https://example/os-9.iso",
		Format: "iso", Checksum: osdb.Checksum{Algorithm: "sha256", Value: "..."},
	}}, nil
}
```

## Rules

- IDs must be unique; duplicate registration is rejected.
- An empty `Meta.ID` is rejected.
- Hooks see only `plugin.Request` (release id + target path), never internals,
  so they can't weaken safety invariants.
- Returning an error from `OnPhase` aborts the installation.
