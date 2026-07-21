# Testing guide

## Run everything

```sh
go test ./...
```

With race detector (needs cgo + C compiler):

```sh
CGO_ENABLED=1 go test -race ./...
```

## What is tested

- `internal/mirrors` — live-list parsers (Debian/Ubuntu/Fedora) against sample
  payloads, curated-country fallback, unsupported-distro rejection, live-source
  precedence. No real network.
- `internal/builder` — full `Build` against a TLS test server serving
  kernel/initrd; config embedding; unsupported-distro and missing-ISO-tool
  failures. Uses a fake `grub-mkrescue` so the assembler path is exercised
  without the real tool.
- `pkg/hashes`, `pkg/osdb`, `internal/verify`, `internal/download`,
  `internal/disk`, `internal/install`, `internal/hardware`, `internal/network`,
  `internal/locale`, `internal/plugin`, `internal/update`, `internal/logging`,
  `internal/config` — as before.

## Test philosophy

- Real over mocked where it matters: builder/mirror tests stand up `httptest` TLS
  servers and use genuine digests.
- Pure logic (parsers, resolution, seed generation) gets exhaustive unit tests.
- Platform code gets parser units; shell-outs are tested by running on the host.

## Coverage

```sh
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```
