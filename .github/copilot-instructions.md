# Copilot Instructions

This file contains the project instructions for GitHub Copilot.
For the full, up-to-date instructions, see `AGENTS.md` in the repository root.

## Project

Pr!nt is a cross-platform open-source OS installer in Go
(`github.com/print-foundation/print`). It downloads official OS images over HTTPS,
verifies them, and writes them to a disk behind explicit confirmation.

## Hard rules (do not violate)

- Keep import direction clean: `internal/domain` depends on nothing internal;
  `pkg/*` depend on nothing internal. No cycles.
- HTTPS-only for every download/verify/update fetch. Never add plaintext URLs.
- Verify before writing. Destructive ops must go through `disk.Guard`. No silent
  overwrite paths. Never execute downloaded content.
- No placeholders/TODOs/stubs presented as complete. Isolate and document genuine
  external deps (the OS catalog and the update manifest URL are the two honest
  trust anchors).
- Platform-specific code lives in build-tagged files (`linux`, `windows`,
  `darwin`, `freebsd`, `fallback`). Add a fallback.
- Plugins are compiled in (registered at `init`), never loaded from disk.

## Commands (run from repo root)

- `go build ./...`
- `go vet ./...`
- `go test ./...` (race needs `CGO_ENABLED=1` + C toolchain; skip `-race` if
  none)
- `gofmt -l .` (must be empty)
- `scripts/test.sh` and `scripts/build.sh`

## Before finishing a task

Run build, vet, test, and gofmt. Cross-compile if you touched platform files:
`GOOS=linux GOARCH=amd64 go build ./...` etc.

## Git conventions

- Use conventional commits: `feat:`, `fix:`, `docs:`, `refactor:`, `test:`,
  `chore:`, etc.
- Lowercase, imperative, no trailing period.
- Never mention that you are an AI assistant in commit messages or PR
  descriptions.
- Keep commits focused; one logical change per commit.

## Pull request conventions

- Title follows conventional commits format.
- Description explains *why* the change is needed, not just what changed.
- Include test evidence (`go test ./...`, `gofmt -l .`, relevant platform
  builds).
- Link related issues.

## Style

gofmt-clean; lowercase why-focused comments; preserve `UEFI`/`GPT`/`SHA-256`/
`PGP`. Match existing package conventions.
