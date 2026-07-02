# genv — agent guide

A CLI for managing environment variables across a monorepo. Standard **Go** module
(Go ≥ 1.25) — a single statically-linked binary, no runtime deps. The **registry**
(`genv.json`) is the single source of truth for *structure*; **values live in
pluggable vaults**; plaintext `.env` files are *generated* from the vault on demand
and git-ignored.

Mental model: edit structure (registry) and values (vaults) via the CLI; never
hand-edit a generated `.env` — it is an output, rewritten by `genv generate`.

## Commands

```bash
go build -o genv ./cmd/genv      # compile the binary
go run ./cmd/genv <args>         # run from source
go run ./cmd/genv init           # create an empty registry + local vault config
go run ./cmd/genv generate       # vault → .env (+ compose) regenerate
go run ./cmd/genv check          # validate the repo (CI gate; exit 1 on findings)
go test ./...                    # whole suite
go test ./internal/generate/     # a single package
go vet ./...                     # static checks
gofmt -l .                       # formatting (should print nothing)
```

Full CLI grammar and concepts live in `README.md`. The command set (`init`;
`vault`/`consumer`/`group`/`global`/`compose` management; `var
define/update/remove/list/show`; `wire`/`unwire`/`enable`/`disable`; `set`/`get`;
`import`; `generate`; `check`; `tui`; `backup`/`restore`; `completion`) is built
in `internal/cli/program.go` (cobra).

## Structure

- `cmd/genv/` — entry point (`main.go`); wires the CLI and registers the `tui`
  command here (not in `internal/cli`) to avoid a cli→tui→cli import cycle
- `internal/cli/` — cobra program (`program.go`), command handlers, output/prompt/run plumbing, `check.go`
- `internal/tui/` — Bubble Tea (Elm architecture) TUI: state/update/views/modals, `keys.go` keymap (footer + help derive from it), the init wizard
- `internal/core/` — pure domain: `errors.go`, `interpolate.go`, `refs.go`, `plan.go`, and the op planners in `internal/core/ops/` (no I/O)
- `internal/registry/` — `genv.json` types, validation, load/save (`persist.go`)
- `internal/vault/` — `VaultProvider` contract, provider registry, `local/` (age encryption), auth resolution (`auth.go`)
- `internal/generate/` — ownership/disclaimer (`ownership.go`), renderers, consumer paths, compose splicing, the generate orchestrator, file-op applier
- `internal/io/` — atomic write, dotenv parse, root discovery, `.gitignore` block, backups
- `tests/helpers/fixtures.go` — builds registries/repos for tests

On-disk layout `init` creates and the CLI manages:

```text
genv.json              # registry — committed
.genv/vault.json       # genv-local store — committed IF encrypted, git-ignored if plaintext
.genv/auth.local.json  # per-machine vault auth — git-ignored, never committed
.genv/backups/         # `genv backup` snapshots — git-ignored
apps/*/.env            # GENERATED — git-ignored, never hand-edited
apps/*/.env.example    # GENERATED values-free template — committed (per-consumer opt-in)
.env.compose           # GENERATED compose interpolation values — git-ignored
```

## Security model

Values live in vaults, addressed by the keys in each variable's `vaultMapping`;
the registry never contains a value. `genv-local` optionally age-encrypts its JSON
(`encryption: true` ⇒ committable ciphertext; `false` ⇒ plaintext, must stay
git-ignored — `genv check` enforces both). The key resolves per vault:
`--vault-auth <vault>=…`, `GENV_VAULT_AUTH_<NAME>`, a `.genv/auth.local.json` hook
(`command`/`env`/`value`), then a TTY prompt. Anything in `internal/vault/` is
security-sensitive.

## Code style

- **Standard Go** — `gofmt`-clean, no lint config beyond `go vet`. Errors are
  `core.GenvError` with a code that maps to the process exit code.
- **`internal/` only** — the whole implementation is unexported; `cmd/genv` is the
  sole public entry.
- Keep `internal/io`/`internal/vault` effects out of `internal/core` (pure ops +
  interpolation + plan).

## Testing

Colocated `_test.go` files with `github.com/stretchr/testify` (`assert`/`require`);
fixtures in `tests/helpers`.

- Op planners are pure — assert `{ next, plan }` without I/O; secrets must never
  appear in serialized plan output.
- The `genv-local` vault provider is covered by its own tests under
  `internal/vault/local/`.
- New behavior needs a test; a fixed bug gets a regression test.

## Mutation ≠ generation

Registry/vault mutations (`set`, `wire`, `var define`, …) NEVER touch generated
files. `genv generate` is the only writer of outputs; `genv check` reports
staleness. Every mutating command supports `--dry-run` and `--output pretty|json`
with a uniform `{ ok, result }` / `{ ok, error }` envelope. Exit codes: 0 success
(incl. dry-run), 1 domain error / blockers / check findings, 2 usage, 3 auth, 4
vault I/O.

## The ownership rule

genv only overwrites or deletes a file whose FIRST line carries the disclaimer
marker (`internal/generate/ownership.go`). A generated path the user has taken over
(marker removed) is left untouched and reported by `check`. `consumer remove`
strips the marker (releasing the file) by default; `--delete-files` deletes it.
Registered compose files are the exception — user-owned, genv only rewrites the
lines between hand-authored `# <genv:consumer>` … `# </genv>` markers.

## Boundaries

- 🤝 **Git:** committing locally does **not** require per-commit approval — commit freely as work lands (still never commit a plaintext secret; see 🚫 below). Pushing to a remote is outward-facing — confirm first.
- ✅ **Always:** run `go test ./...` and `gofmt -l .` before claiming done. When a command/flag, on-disk layout, vault provider, or the registry schema changes, update `README.md` in the **same** change.
- ⚠️ **Ask first:** changing `genv.json`'s `schemaVersion` or shape (existing repos need migration). Bumping core deps (`bubbletea`, `cobra`, `age`). Anything in `internal/vault/` that changes encryption or the provider contract.
- 🚫 **Never:** commit a plaintext `.env`/`.env.*` or a plaintext vault file; print a real secret in code, tests, logs, or a plan. Hand-edit a generated file — it's an output.

---
Source of truth: `internal/` (behavior), `go.mod` (deps), `README.md` (user-facing),
`docs/superpowers/specs/2026-06-12-menv-v2-design.md` (design — note the design docs
retain the original **menv** naming; genv is the rebrand). Update when a
command/flag, on-disk layout, vault provider, or the registry schema changes; a core
dependency is bumped; or a directory moves.
