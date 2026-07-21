# Developer guide

## Layout

```
cmd/print         CLI: pick distro -> country -> mirror -> build/flash ISO
cmd/print-client  program embedded in the ISO (get online -> install from cloud)
docs/             guides
internal/
  domain          shared value types (no internal deps)
  config          user config load/save
  logging         slog wrapper, file sink, crash reporter
  hashes (pkg)    digests + checksum files
  osdb (pkg)      generic release metadata type (verify/install use it)
  verify          checksum + PGP verification
  download        resumable verifying downloader
  disk            disk detection + guard
  hardware        hardware detection (build-tagged per OS)
  network         connectivity + WiFi (build-tagged per OS)
  install         flash a built ISO to USB (guarded) + recovery
  mirrors         resolve country -> mirror
  builder         download netboot assets + verify + assemble ISO
  locale          i18n strings
  plugin          extension registry
  update          self-update
```

## Local dev loop

```sh
go build ./...
go vet ./...
go test ./...
gofmt -w .
```

## Editing netboot asset paths

`internal/builder/spec.go` holds the per-distro netboot layout (kernel/initrd
paths relative to the mirror). These are the publishers' documented locations and
drift between releases. If a build 404s on an asset, update the path there and
add a test case if the layout changed.

## The in-ISO client

`cmd/print-client` is compiled for `GOOS=linux` at build time and embedded. It
depends only on Pr!nt's own packages so it stays small. Edit its seed generation
(`seed.go`) when a distro changes its preseed/kickstart format.

## Debugging without a TTY

Use the CLI flags: `-list-distros`, `-distro ... -country ... -out ...` (build
only, no flashing), and `-write -iso ... -device ... -yes` (flash only).
