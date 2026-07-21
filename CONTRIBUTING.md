# Contributing

Thanks for considering a contribution to Pr!nt. This is a production-minded,
open-source project: we care about correctness, safety, and tests more than
feature count.

## Ground rules

- **No placeholders, TODOs, or stubs presented as complete.** If you can't make
  something real yet, isolate and document the external dependency — don't fake
  it. (Example: Windows/macOS ISOs are *not* built; we say so, we don't fake one.)
- **Every fetch must be HTTPS-only and verify before use.** This is the project's
  core promise, including the ISO builder and in-ISO client.
- **Destructive operations must go through `disk.Guard`.** Do not add a way to
  skip confirmation when flashing an ISO to a device.
- **Keep packages dependency-direction clean.** `internal/domain` depends on
  nothing internal; `pkg/*` depend on nothing internal. No cycles.
- **Tests are required for new behavior.** Mirror parsers, builder logic, and
  seed generation need unit tests. Platform shell-outs get parser tests.

## Code style

- `gofmt` clean. CI fails on unformatted code.
- Lowercase, why-focused comments. Preserve names like `UEFI`, `GPT`, `SHA-256`,
  `PGP`.

## Workflow

1. Fork and branch from `main`.
2. Make the change with tests.
3. Run locally:
   ```sh
   go build ./...
   go vet ./...
   go test ./...
   gofmt -l .
   ```
4. Open a PR describing the *why*.

## Editing netboot/mirror data

- `internal/builder/spec.go` — per-distro netboot asset paths (drift between
  releases; refresh when a fetch 404s).
- `internal/mirrors/curated.go` — curated country→mirror map (refresh if a host
  changes). Live mirror lists take precedence when reachable.
- Validate with `go test ./internal/builder/... ./internal/mirrors/...`. These
  changes are security-relevant because they affect what gets downloaded.

## Adding a plugin

Register it at `init` in a package that `cmd/print` imports, implementing the
`plugin.Hook` and/or `plugin.CatalogSource` interface. Do not add runtime
shared-object loading.

## Code of conduct

Be respectful. Assume good faith. Keep discussions focused on the code and the
user's safety.
