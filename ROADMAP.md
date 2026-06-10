# Roadmap

For the latest, see [milestones](https://github.com/RefuseHQ/refuse-cli/milestones) and the [project board](https://github.com/orgs/RefuseHQ/projects). Dates intentionally absent.

## Recently shipped

- [x] All CLI subcommands wired up: `refuse init`, `install`, `uninstall`, `hook install/remove/list`, `allowlist add/remove/list`, `check`, `check-lockfile`, `audit`, `gate`, `pip-gate`, `python-hook install/status/uninstall`, `config show/set/get`, `status`, `doctor`.
- [x] `refuse doctor` inspects the environment — PATH, shim resolution, server reachability, hook presence.
- [x] Project-level allowlist via `.refuse.yaml` (`refuse allowlist add <CVE-ID>`).
- [x] Extended manager coverage to 18: `npm`, `pnpm`, `yarn`, `bun`, `npx`, `pip`, `pip3`, `uv`, `poetry`, `pipenv`, `pdm`, `pipx`, `cargo`, `gem`, `bundle`, `go`, `composer`, `dotnet`.
- [x] Self-hostable server ([RefuseHQ/refuse](https://github.com/RefuseHQ/refuse)) — bulk first-boot seed (~3 min) + parallel per-tick deltas. Image at `ghcr.io/refusehq/refuse`.
- [x] First-class Windows support — PATHEXT-aware shim resolution, libuv-safe SIGINT handling, PowerShell profile install path, hardlink fallback.
- [x] GitHub Actions composite action ([action.yml](./action.yml)) + pre-commit hooks (`refuse-check`, `refuse-check-installed-pip`).
- [x] Cosign-signed release binaries + SLSA build provenance + sha256 checksums.
- [x] golangci-lint + cross-platform build matrix (GoReleaser).
- [x] Homebrew tap.
- [x] `refuse python-hook` to close the `python -m pip` / `uv pip` / `conda install` bypass.

## In progress

- [ ] **More agent hooks**: Cursor, Continue, Aider, Cline, Codex CLI. Claude Code is the reference implementation; the others are tracked.
- [ ] **`refuse explain <pkg>`** — full advisory + suggested upgrade with diff against current version.
- [ ] **Lockfile diff mode** — `refuse check-lockfile --diff` to scan only added or upgraded entries.
- [ ] **Container image** — `ghcr.io/refusehq/refuse-cli` so CI jobs can drop in the CLI without `install.sh`.
- [ ] **Block-message polish** — multi-line, colored, with a copy-paste upgrade command.
- [ ] **Test coverage gate** in CI (fail under 70%).

## Towards 1.0

- [ ] Stable CLI surface — subcommand and flag names locked.
- [ ] Stable hook contract — documented JSON shape any agent can adopt, not just Claude Code.
- [ ] Distro packages: `apt`, `dnf`, `pacman`, `apk`.
- [ ] `mise` and `asdf` shim integration.
- [ ] LTS branch once 1.0 ships.

## Out of scope

- Becoming a full SCA tool — we're a gate, not a scanner. See [refuse](https://github.com/RefuseHQ/refuse) and [OSV-scanner](https://github.com/google/osv-scanner) for that.
- Replacing package managers. We exec real binaries; we won't re-implement npm.
- Network calls to anything other than the configured `server_url`. No telemetry, ever.

## How to influence this

- Open a [Discussion](https://github.com/RefuseHQ/refuse-cli/discussions) with the use case.
- Ship a PR. Concrete code moves things much faster than "wouldn't it be nice if".
