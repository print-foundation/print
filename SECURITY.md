# Security

Pr!nt builds small cloud-netboot ISOs. This document is honest about what is
guaranteed versus what is best-effort.

## Guarantees

- **HTTPS only.** `mirrors`, `builder`, and the in-ISO client refuse any
  non-`https://` URL. TLS minimum is 1.2 with validation on.
- **Verify before trust (when possible).** Netboot assets are checked against the
  publisher's `SHA256SUMS` when the distro publishes one. If none exists, the
  build completes but is flagged `verified=false` and logged — we never claim a
  file was verified when it wasn't.
- **Confirmation required for destructive ops.** Flashing a built ISO to a USB
  stick goes through `disk.Guard`: explicit, device-path-matched acknowledgement.
  The running system disk is never a valid target.
- **No execution of untrusted content.** The in-ISO client writes a seed and
  invokes the distro's *own* installer; it does not run downloaded code directly.
- **No fake artifacts.** Windows/macOS are reported as unsupported; Pr!nt never
  produces a broken or misleading ISO for them.

## Trust anchors

1. **The distro mirror** (HTTPS) — where netboot assets and the full OS come
   from. Pr!nt picks it from the distro's official mirror list or a curated map.
2. **The publisher `SHA256SUMS`** — best-effort checksum of netboot assets.
3. **`grub-mkrescue`/`xorriso`** — the external ISO assembler. Pr!nt shells out to
   it; if it's missing, the build fails with a clear message rather than writing
   a corrupt file.

## Best-effort / honest limits

- Some distros (Arch, Alpine, FreeBSD, OpenBSD) publish netboot assets without a
  convenient `SHA256SUMS`; those are downloaded but not checksum-verified. The
  build reports this.
- The "install from cloud" step runs the *distro's* installer inside the booted
  ISO. Pr!nt's guarantees end at handing off; the installer's own supply chain is
  the distro's responsibility.
- The netboot asset *paths* in `builder/spec.go` are the publishers' documented
  layouts and do drift between releases; refresh them when a fetch 404s.

## Reporting

Report security issues privately to the maintainers; do not open public issues
with exploit code.
