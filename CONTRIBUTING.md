# Contributing to refuse-cli

Thanks for thinking about contributing. refuse-cli is a small Go binary — the kind of codebase where a careful 50-line PR can move the project meaningfully forward.

For the **server** side, see [`RefuseHQ/refuse`](https://github.com/RefuseHQ/refuse). The hosted service lives at [refuse.dev](https://refuse.dev).

## Ways to help

- **A package manager we don't wrap yet.** `internal/parsers/` has one file per ecosystem (`npm.go`, `pip.go`, `yarn.go`, …). Adding a new one means: parser + tests + register it in `internal/shim/shim.go`.
- **A coding agent we don't hook yet.** `internal/hook/` is where Claude Code lives. Cursor, Continue, Aider, Cline, Codex — all welcome.
- **A platform we don't release for.** Today we ship darwin/linux/windows × amd64/arm64 via goreleaser. FreeBSD, OpenBSD, and Linux/armv7 are plausible additions.
- **Better error messages.** When we block an install, the message has to teach the developer something. Improvements to wording are first-class PRs.
- **The install script.** `scripts/install.sh` is plain shell with checksum verification. If it breaks on your platform, please report it.

## Development setup

Requirements:

- Go ≥ 1.21
- `make` (or just call `go` directly)
- A running [refuse](https://github.com/RefuseHQ/refuse) server. The default gate fails open with a warning if the server is unreachable, so iteration without one is fine — but most decision paths can't be exercised that way.

```sh
git clone https://github.com/RefuseHQ/refuse-cli.git
cd refuse-cli
go mod download
make build           # → dist/refuse
make test            # → go test ./...
make typecheck       # → go vet ./...
```

Once built, drop the binary on your PATH and run `refuse --version`. For interactive use:

```sh
./dist/refuse config set server_url http://localhost:8080
./dist/refuse install       # writes shims to ~/.refuse/bin
./dist/refuse status
```

## Repo layout

```
cmd/refuse/main.go         # multicall entry: argv[0] decides shim vs CLI mode
internal/
  cli/                     # Cobra commands (gate is the real one; others are stubs)
  config/                  # layered config: defaults → user → project → env
  gate/                    # decision engine — vets a parsed install against the server
    decide.go              #   "should this install be allowed?"
    protocol.go            #   Claude Code PreToolUse hook input parsing
  hook/                    # agent-hook integration (claude-code today)
  parsers/                 # per-ecosystem argv parsers (npm, pip, yarn, …)
  server/                  # HTTP client for the refuse REST API
  shim/                    # PATH shim install + real-binary exec
  version/                 # build-time metadata, injected by ldflags
scripts/install.sh         # curl-based installer with sha256 verification
.goreleaser.yaml           # release config (darwin/linux/windows × amd64/arm64)
testdata/                  # fixtures for parser tests (currently empty)
```

See [ARCHITECTURE.md](./ARCHITECTURE.md) for a deeper walkthrough.

## Coding conventions

- **`gofmt` and `go vet` must pass.** CI enforces this. We're adding `golangci-lint` — see `.golangci.yaml`.
- **No third-party logging libraries.** stdlib `log/slog` is enough.
- **Dependencies are precious.** Today the entire dependency tree is Cobra, shlex, yaml. Adding a new one needs a justification in the PR description.
- **Errors carry context.** Use `fmt.Errorf("doing X: %w", err)`. The decision message a user sees ultimately comes from one of these chains.
- **Parsers are pure.** They take `[]string` (argv) and return `ParseResult`. No I/O, no env reads. This is what makes them testable.
- **Don't import `internal/cli/` from anywhere else.** Cobra commands are the leaf of the import graph.

## Tests

- All tests are stdlib `testing`, table-driven where it makes sense. We deliberately avoid `testify`.
- Each parser has a sibling `_test.go` file. New parsers must come with one.
- `internal/gate/decide_test.go` covers the decision matrix — severity tiers, fail-open/closed, env overrides. New decision behavior gets new rows in the table.

```sh
go test ./...
go test ./internal/parsers/ -run TestParseNpm -v
go test -race -coverprofile=cover.out ./...
go tool cover -html=cover.out
```

## Commit & PR style

- One logical change per PR. The codebase is small enough that this is enforceable.
- Conventional commits encouraged (`feat:`, `fix:`, `docs:`, `chore:`).
- Reference the issue you close: `Closes #N`.
- Add a row to `CHANGELOG.md` under `[Unreleased]` for any user-visible change.

CI runs `go vet`, `go test`, and `go build` on every push. Releases (on `v*` tags) run goreleaser and publish to GitHub Releases + the Homebrew tap.

## Adding a new package manager

Concrete checklist, in order:

1. `internal/shim/shim.go` — add to `KnownManagers` map with its ecosystem identifier.
2. `internal/parsers/<manager>.go` — implement the `Parser` interface (`Parse([]string) (ParseResult, error)`).
3. `internal/parsers/<manager>_test.go` — table-driven cases covering: passthrough (non-install), direct install, lockfile install, version pinning, multiple packages, edge flags.
4. `internal/parsers/registry.go` — register the parser.
5. Manual test: `./dist/refuse install`, then run `<manager> install <some-package>` and confirm the gate kicks in.
6. Update the table in `README.md`.

PRs that add a new manager and ship with tests have been merged within a day historically.

## Adding a new agent hook

1. `internal/hook/<agent>.go` — implement the install/uninstall logic. The reference is `claudecode.go`.
2. `internal/hook/registry.go` — register it.
3. `internal/cli/hook.go` — should pick it up automatically via the registry.
4. Document the agent in `README.md`.

## Reporting security issues

Don't open a public issue. See [SECURITY.md](./SECURITY.md).

## Code of conduct

By participating you agree to abide by the [Code of Conduct](./CODE_OF_CONDUCT.md).

## License

By contributing you agree your changes are licensed under [Apache License 2.0](./LICENSE).
