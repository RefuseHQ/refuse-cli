# refuse-cli

> Wraps `npm`, `pip`, `cargo`, `yarn`, `pnpm`, `gem`, `bun`, and `go` as a PATH shim — refuses to install packages with known CVEs.

`refuse` sits in front of your package managers. Each `install` call is vetted against the [`refuse`](https://github.com/RefuseHQ/refuse) server (self-hosted or hosted) before the real binary runs; if the package has a known advisory above your severity threshold, the install is blocked with the CVE and a suggested safe version.

Works on a developer's laptop, in a CI job, or inside a Docker build stage. Optional integration with coding-agent pre-tool-use hooks (Claude Code today, more soon) so an agent's autonomous installs hit the same gate.

## Install

Homebrew:

```sh
brew install refusehq/tap/refuse
```

Direct binary:

```sh
curl -sSL https://raw.githubusercontent.com/RefuseHQ/refuse-cli/main/scripts/install.sh | sh
```

Or from source:

```sh
go install github.com/RefuseHQ/refuse-cli/cmd/refuse@latest
```

## Quickstart

```sh
refuse init                       # interactive: server URL + API key
refuse install                    # drop shims into ~/.refuse/bin + update PATH
refuse hook install claude-code   # PreToolUse hook in ~/.claude/settings.json
```

Now when Claude Code (or you) runs `npm install lodash@4.17.10`, the shim consults the refuse server, blocks the install, and reports the CVE.

## Commands

| Command | What it does |
| --- | --- |
| `refuse init` | First-time setup wizard |
| `refuse install` | Install shims for the supported package managers |
| `refuse uninstall` | Remove shims + revert shell-rc edits |
| `refuse hook install <agent>` | Write a pre-tool-use hook for `<agent>` |
| `refuse hook remove <agent>` | Remove that hook |
| `refuse hook list` | Show all installed hooks |
| `refuse check <eco> <pkg>[@<ver>]` | One-off check |
| `refuse check-lockfile <path>` | Scan an entire lockfile |
| `refuse gate` | The decision engine — shims + hooks call this on stdin |
| `refuse config show \| set \| get` | Manage `~/.refuse/config.yaml` |
| `refuse status` | Diagnose install state |
| `refuse doctor` | Verify PATH / hooks / server reachability |

## Pointing at a self-hosted server

```sh
refuse config set server_url http://localhost:8080
```

The companion self-hostable server lives at [RefuseHQ/refuse](https://github.com/RefuseHQ/refuse) — `docker run --rm -p 8080:8080 ghcr.io/refusehq/refuse:latest` and you're up.

## Status

This is an alpha release — the scaffolding is in place, but most subcommands are still stubs (`refuse <cmd>` prints "not implemented yet"). See open issues for the v0.1.0 plan.

## Licence

Apache-2.0 — see [LICENSE](LICENSE).
