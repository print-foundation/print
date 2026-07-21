# Build guide

## Requirements

- Go 1.24 or newer.
- For **ISO assembly**: `grub-mkrescue` (from `grub2`/`grub-common`) **or**
  `xorriso` on your `PATH`. Pr!nt shells out to one of these; if neither exists,
  the build fails with a clear message.
- For **building the in-ISO client**: a `go` toolchain at runtime (Pr!nt runs
  `go build` for `GOOS=linux` to produce the embedded client binary).

## Standard build

```sh
go build ./...
```

## Build an ISO

```sh
go run ./cmd/print -distro debian -country DE -out debian.iso
```

This compiles `cmd/print-client` for `GOOS=linux` internally, so it needs `go`
available. The resulting ISO is written to `-out`.

## Cross-compile Pr!nt itself

Pr!nt (the builder) runs on the host OS; only the embedded client targets Linux.

```sh
GOOS=linux   GOARCH=amd64 go build -o dist/print-linux   ./cmd/print
GOOS=windows GOARCH=amd64 go build -o dist/print.exe     ./cmd/print
GOOS=darwin  GOARCH=arm64 go build -o dist/print-darwin ./cmd/print
```

## Version stamping

```sh
go build -ldflags "-X main.version=$(git describe --tags --always)" -o dist/print ./cmd/print
```

## Build tags

Platform-specific detection (disk, hardware, network, writer) uses build
constraints with a `*_fallback.go`. Don't add OS-specific logic outside those.
