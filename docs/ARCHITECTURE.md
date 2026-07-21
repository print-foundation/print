# Architecture

Pr!nt is now a **cloud-netboot ISO builder**. The dependency direction stays
clean: leaf libraries at the bottom, orchestration in `internal/*`, wiring in
`cmd/print`.

## Layers

```
cmd/print
   -> internal/mirrors   (country -> mirror resolution)
   -> internal/builder  (download netboot assets + verify + assemble ISO)
   -> internal/install  (flash built ISO to USB, guarded)
   -> internal/network  (WiFi/connectivity, reused by the in-ISO client)
   -> pkg/osdb, internal/config, internal/logging

cmd/print-client  (embedded in the ISO)
   -> internal/network  (get online)
   -> builder.ClientConfig (which mirror / wifi to use)
   -> optional: generates a distro-specific unattended seed
```

`internal/domain` is the bottom of the graph; `pkg/*` depend on nothing
internal. No cycles.

## The build pipeline (`internal/builder.Build`)

1. **Resolve mirror** via `internal/mirrors` (live list when available, curated
   fallback otherwise).
2. **Resolve netboot asset** layout for the distro/arch (`spec.go`).
3. **Download** kernel + initrd (+ extras) over HTTPS, with resume.
4. **Verify** against the publisher `SHA256SUMS` when present; otherwise degrade
   to "unverified" and report it (never pretend).
5. **Embed** `print-client.json` (mirror + wifi + distro) and the compiled
   `print-client` binary.
6. **Assemble** a bootable ISO with `grub-mkrescue` (preferred) or `xorriso`.
   A clear error explains the missing tool rather than faking output.

The grub config boots the netboot kernel+initrd and passes the client location
on the kernel cmdline.

## The in-ISO client (`cmd/print-client`)

The client does **not** auto-install. On boot (`prepare`):
1. Loads `print-client.json`.
2. Connects to the preconfigured WiFi if set (else assumes wired DHCP).
3. Confirms it can reach the mirror (`network.Connectivity`).
4. Writes an *optional* unattended seed (preseed/kickstart/answer) to disk.
5. Returns, leaving the user in the distro's **live environment**.

The unattended path is a separate, explicit subcommand — `print-client
autoinstall` (the "Arch Autoinstall" style option). Only that path invokes the
distro installer with the seed; it is never run on boot.

This is why the ISO stays ~300–500 MB: it carries only the boot environment and
the client, not the OS.

## Mirror resolution (`internal/mirrors`)

Each distro exposes mirrors differently. `liveSources` maps distros with a
machine-readable list (Debian `list-full`, Ubuntu `mirrors.txt`, Fedora
mirrorlist) to a parser; the rest use a curated country→mirror map. All over
HTTPS; no URL is invented.

## Why Windows/macOS are excluded

They have no redistributable netboot installer, so a "small cloud ISO" isn't
something Pr!nt can legally or technically produce. `mirrors.Supported` returns
false for them and the CLI shows a clear "requires licensed media" message.

## Flashing (`internal/install`)

`FlashLocal` writes a built ISO to a USB device through `DeviceWriter`, gated by
`disk.Guard` (explicit, device-path-matched confirmation). The same engine that
once flashed full images now flashes the small one.

## Plugins

`internal/plugin` still defines the compile-in extension registry; lifecycle
hooks fire during installation. No runtime shared-object loading.
