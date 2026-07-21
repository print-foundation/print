# Release guide

Pr!nt's release artifact is the builder binary plus the in-ISO client; the ISOs
themselves are produced per-run by end users (they choose distro + mirror).

## Steps

1. **Refresh netboot asset paths** in `internal/builder/spec.go` if a distro
   changed its netboot layout. Run `go test ./internal/builder/...`.
2. **Refresh curated mirrors** in `internal/mirrors/curated.go` if a host
   changed. Run `go test ./internal/mirrors/...`.
3. **Bump version** (`main.version`), stamped via `-ldflags "-X main.version"`.
4. **Tag and push:**
   ```sh
   git tag v1.2.3
   git push origin v1.2.3
   ```
5. **Build artifacts:**
   ```sh
   go build -ldflags "-X main.version=v1.2.3" -o dist/print ./cmd/print
   ```
   (Requires `grub-mkrescue`/`xorriso` and `go` for client compilation.)
6. **Publish the update manifest** (optional, for `internal/update`) at the URL
   `cmd/print` checks. The manifest must be HTTPS with per-platform SHA-256.

## Why the catalog.json is not the release artifact

`pkg/osdb/catalog.json` still exists as generic release *metadata*, but the
product no longer flashes prebuilt images. The shippable thing is the builder;
ISOs are generated on demand from live mirrors. Keep `catalog.json` accurate only
if you use it for anything; otherwise it is legacy.

## Signing

If PGP signatures are added for the builder binary or netboot assets, document
the key fingerprint in SECURITY.md and wire `verify.WithSignatureVerifier`.
