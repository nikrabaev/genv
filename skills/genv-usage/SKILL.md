---
name: genv-usage
description: >-
  Use when working in a repo managed by genv — the structure of environment
  variables lives in a committed registry (genv.json at the root) and the VALUES
  live in pluggable vaults; the .env files are GENERATED, not authored. Trigger
  whenever a task touches env vars, secrets, API keys, database URLs, or
  .env / .env.example / .env.compose files in such a repo: adding or changing a
  variable, wiring one to a consumer, reading or setting a value, or "my .env is
  empty / out of date". ESPECIALLY before editing any .env by hand — in a genv
  repo that edit is the wrong move and gets silently overwritten by
  `genv generate`. If you see a genv.json or a .genv/ directory, this skill
  applies.
---

# Using genv

`genv` keeps the *structure* of a repo's environment in one committed
**registry** (`genv.json`) and the *values* in pluggable **vaults** (the bundled
`genv-local` vault age-encrypts them). It **generates** the plaintext `.env`
files on demand. A repo uses it if there's a `genv.json` at the root. `genv` is a
single binary — run `genv --help` to confirm it's installed; if it isn't found,
install it with `go install github.com/nikrabaev/genv/cmd/genv@latest` or grab a
release binary (see the project README).

## The one rule

**The registry is the source of truth for structure; values live in vaults —
never in the registry, and never in a `.env`.** Every `.env`, `.env.example`, and
`.env.compose` is a generated *output*.

Never hand-edit a generated `.env`, and never `git add` one. A hand edit isn't a
real change — the next `genv generate` rewrites that file from the vault, so the
edit silently vanishes. Change *structure* through registry commands and *values*
through `genv set`, then run `genv generate`. `genv check` flags any drift between
the registry/vault and the files on disk.

## Gotchas that bite (the part worth reading)

- **Keep secret values out of shell history and logs.** Pipe them on stdin
  (`printf '%s' "$V" | genv set NAME`) or omit the value and let `genv set`
  prompt. Never pass a secret as a literal CLI argument. `genv get` prints the
  *raw* value (secrets included) — don't `echo`/log it; `var list` / `var show`
  and any `--output json` plan mask them.
- **Prefer machine-readable output and previews.** Use `--output json` for results
  you parse, and `--dry-run` to preview *any* mutation before applying it.
- **Run `genv check` after a batch of changes.** `genv check --output json`; exit
  **1** means problems (broken interpolation refs, stale generated files, a
  plaintext vault or `.env` tracked by git).
- **Vault auth must be non-interactive in scripts.** Supply it via
  `GENV_VAULT_AUTH_<VAULT>` (upper-cased vault name) or a `.genv/auth.local.json`
  entry. Off a TTY, genv never prompts — a missing key is a hard error (exit 3).
- **Values are single-line.** A multi-line value (e.g. a PEM key) isn't supported
  — put it on one line with escaped `\n`, or keep it out of the vault.
- **Mutations never write outputs.** `set`, `wire`, `var define`, … only touch the
  registry/vault. `genv generate` is the only writer of `.env` files.

## Command cheatsheet

```bash
# read (structure + masked values)
genv var list [--vault V] [--consumer C] [--output json]   # variables, secrets masked
genv var show NAME                                         # one variable, secrets masked
genv get NAME [--vault V]                                  # RAW value to stdout — don't log secrets

# change structure (registry)
genv var define NAME --secret --description "…"            # define a variable
genv wire NAME --vault V --consumers api,worker            # new per-consumer key each; --shared/--key K to share one value
genv unwire NAME --vault V --consumers worker              # remove that mapping

# change values (vault)
genv set NAME [value] [--vault V] [--consumer C]          # value via arg, stdin, or masked prompt
printf '%s' "$V" | genv set NAME                          # a secret, via stdin

# generate + validate (the only writer of outputs; the CI gate)
genv generate [--vault V] [--consumer C]                  # rewrite .env / .env.example / .env.compose
genv check --output json || exit 1                        # exit 1 ⇒ problems
```

Every mutating command takes the global `--dry-run`, `--output pretty|json`,
`--force`, and `--vault-auth <vault>=<secret>` flags. When unsure what exists or
how something is wired, `genv var list --output json` shows the picture before you
mutate anything.
