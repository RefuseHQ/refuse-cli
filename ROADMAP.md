# Roadmap

For the latest, see [milestones](https://github.com/RefuseHQ/refuse-cli/milestones) and the [project board](https://github.com/orgs/RefuseHQ/projects). Dates intentionally absent.

## Status: Alpha

The gate is the production-ready piece. Most other subcommands (`refuse init`, `refuse install`, `refuse hook install`) are wired up but thin — some still print "not yet implemented" for edge cases. Use at your own risk and please file issues.

## Recently shipped

- [x] Cosign-signed release binaries + checksums.
- [x] SLSA build provenance attestation.
- [x] golangci-lint configuration and CI step.
- [x] Cross-platform build matrix (GoReleaser).
- [x] Homebrew tap.

## Near-term — towards 0.1.0

- [ ] Finish all the CLI stubs: `refuse init`, `refuse install`, `refuse uninstall`, `refuse hook install/remove/list`, `refuse check`, `refuse check-lockfile`, `refuse status`, `refuse doctor`.
- [ ] `refuse doctor` actually inspects the environment (PATH, shim resolution, server reachability, hook presence).
- [ ] Better block message — multi-line, colored, with copy-paste upgrade command.
- [ ] Test coverage thresholds in CI (fail under 70%).

## Medium-term — towards 0.2.0

- [ ] **More agent hooks**: Cursor, Continue, Aider, Cline, Codex.
- [ ] **More managers**: poetry, uv, pipenv, mise, asdf.
- [ ] **`refuse explain <pkg>`** — print the full advisory + suggested upgrade with diff against current version.
- [ ] **Lockfile diff mode** — `refuse check-lockfile package-lock.json` should accept `--diff` to scan only added/upgraded entries.
- [ ] **Project-level allowlist** — `.refuse.yaml` can pin specific advisories to "acknowledged" if a team has decided to live with one.
- [ ] **Container image** — `ghcr.io/refusehq/refuse-cli`, useful for CI jobs that don't want to install via brew or install.sh.

## Longer-term — towards 1.0

- [ ] Stable CLI surface. Today subcommands and flag names may still change.
- [ ] Stable hook contract — a documented JSON shape that any agent can adopt, not just Claude Code.
- [ ] First-party Windows support (currently tier 2 — releases ship, but day-to-day testing is on macOS/Linux).
- [ ] Distro packages: `apt`, `dnf`, `pacman`, `apk`.
- [ ] LTS branch once 1.0 ships.

## Out of scope

- Becoming a full SCA tool — we're a gate, not a scanner. See [refuse](https://github.com/RefuseHQ/refuse) and [OSV-scanner](https://github.com/google/osv-scanner) for that.
- Replacing package managers. We exec real binaries; we won't re-implement npm.
- Network calls to anything other than the configured `server_url`. No telemetry, ever.

## How to influence this

- Open a [Discussion](https://github.com/RefuseHQ/refuse-cli/discussions) with the use case.
- Ship a PR. Concrete code moves things much faster than "wouldn't it be nice if".
