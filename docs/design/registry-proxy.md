# Design — Local Registry Proxy

Status: **scaffolded** — CLI surface reserved, implementation in flight.
Tracks the "airtight bypass solution" line of the roadmap.

## Problem

Every interception layer refuse has today is bypassable:

| Layer | Caught | Bypassed by |
| --- | --- | --- |
| PATH shim | `pip install`, `npm install`, … | `python -m pip`, absolute-path calls, `uv pip`, `conda install`, anything that doesn't resolve via PATH |
| `python -m pip` hook (`refuse python-hook`) | module-form pip in envs we've installed into | uv, poetry, conda, ad-hoc envs |
| Agent PreToolUse hook (Claude Code) | the Bash tool calls of one agent | every other agent |
| Lockfile scan (`refuse audit`, pre-commit, CI Action) | anything that ended up resolved | not bypassable, but only *post-hoc* (the install already happened) |

The only architecturally **airtight** interception is at the **transport** —
the moment the package manager fetches the package from a registry. Every
install, no matter how invoked, fetches from a registry over HTTP.

## Architecture

```
┌─────────────┐    HTTPS    ┌─────────────────────┐    HTTPS    ┌─────────────┐
│  pip / npm  │ ──────────► │  refuse proxy       │ ──────────► │  pypi.org   │
│  /cargo/etc │             │  proxy.refuse.dev   │             │  npmjs.org  │
└─────────────┘             │  /<eco>/<...>       │             │  crates.io  │
                            │                     │             │  …          │
                            │  gate.Decide() ────►│             │             │
                            │  if block: 403      │             │             │
                            │  if allow: forward  │             │             │
                            └─────────────────────┘             └─────────────┘
```

Managers are pointed at the proxy via standard config:

```sh
pip   config set global.index-url      https://proxy.refuse.dev/pypi/simple/
npm   config set registry               https://proxy.refuse.dev/npm
cargo config set source.crates-io.replace-with refuse
```

The proxy:

1. Parses the requested path → `(ecosystem, package, version?)`.
2. Calls the existing gate engine.
3. On allow: HTTP 301/307 to the upstream URL, or stream the upstream response through.
4. On block: HTTP 403 with a refuse error body (JSON + a human-readable
   `WWW-Authenticate`-style header) so the manager's error message
   surfaces the CVE / suggested fix.

This catches **everything** — `python -m pip`, `uv`, `conda`, Docker
build layers, IDE installs, arbitrary scripts. There is no
"PATH-shadowing" or "module-import-order" trick that can dodge it.

## Why not just do this everywhere

Three real costs to weigh:

1. **Latency.** Every install now does an extra round-trip through
   refuse before reaching the registry. The hosted proxy must run in
   front of OSV-already-warm tables; the keep-warm cron mitigates D1
   cold starts but the *minimum* path length is now `manager → refuse →
   upstream → refuse → manager`.

2. **Trust.** The user's machine is now trusting refuse with every
   package download. Compromise of the proxy is a supply-chain
   vulnerability. Mitigations: cosign-signed proxy responses for the
   blob path; SRI-verified content; the proxy *only* gates `(name,
   version)` lookups and 301-redirects the actual blob fetch to the
   upstream CDN, so the bytes never touch refuse infra.

3. **Operator burden.** Self-hosters now need to run the proxy as well
   as the gate. For the OSS edition this is fine (it's another endpoint
   on the same Hono server); for the hosted edition the proxy needs its
   own SLA and CDN footprint.

For these reasons the proxy is **opt-in**, not the default. The
PATH-shim + python-hook + agent-hook story is the default; the proxy is
the "I want airtight" upgrade path.

## Ecosystem coverage (priority order)

Implementation order is driven by user demand × registry-protocol
complexity:

1. **PyPI (simple index).** Static HTML/JSON; trivial to parse. `pip
   config set global.index-url`. Highest user value (closes the `python
   -m pip` + `uv pip` + `conda install` gap in one shot).
2. **npm.** JSON registry. `npm config set registry`. Second highest
   value (npx, every Node tool).
3. **crates.io.** `[source.crates-io.replace-with]` redirection.
4. **RubyGems.** `gem sources` + `Gemfile`-level `source` directive.
5. **Go modules.** `GOPROXY` env var. Trivial config; Go's protocol is
   well documented.
6. **NuGet.** Multi-source config; trickier.
7. **Packagist.** Composer's `composer config repo` mechanism.

Each is a separate PR. The CLI command surface is shared:

```sh
refuse proxy enable <ecosystem> [--url=...]   # configure the manager
refuse proxy disable <ecosystem>              # revert; state is saved
refuse proxy status                           # which ecosystems are routed
```

## CLI design

State persists at `~/.refuse/proxy-state.json`:

```json
{
  "version": 1,
  "ecosystems": {
    "pypi": {
      "enabled_at": "2026-06-03T00:00:00Z",
      "previous_index_url": "https://pypi.org/simple/",
      "proxy_url": "https://proxy.refuse.dev/pypi/simple/"
    },
    "npm": {
      "enabled_at": "2026-06-03T00:00:00Z",
      "previous_registry": "https://registry.npmjs.org/",
      "proxy_url": "https://proxy.refuse.dev/npm"
    }
  }
}
```

`enable` reads + records the manager's current setting, then runs the
manager's own config CLI (`pip config set`, `npm config set`) so the
proxy URL lands in the same precedence chain a human would have used.
`disable` restores the recorded value (or runs the `unset` variant if
there was no prior value).

`refuse doctor` gains a check that the configured proxy URL is
reachable and returns the expected refuse-flavored health response.

## Server side

The proxy lives in `RefuseHQ/refuse` (OSS) and `RefuseHQ/core` (hosted)
as a sibling to the existing `/api/v1/check/*` endpoints under
`/proxy/<ecosystem>/...`. Implementation notes for each ecosystem live
alongside the route handlers in the server repo, not here.

Health: `GET /proxy/<ecosystem>/__refuse_health` returns a small JSON
body confirming the proxy is up. This is what `refuse doctor` probes.

## Out of scope for v1

- Caching upstream responses at the proxy (latency / cost; do later).
- Authenticated upstream registries (pip extras, private npm scopes).
- Subresource integrity beyond what the upstream already provides.
- Web-UI for proxy state.

## Status today

- This document.
- `refuse proxy` CLI command tree scaffolded (`enable`, `disable`,
  `status` subcommands) but currently print "coming in a follow-up
  release" — the reservation makes the design discoverable in
  `refuse --help` without shipping broken behavior.

Next PRs:

1. PyPI proxy endpoint + `refuse proxy enable pypi` real implementation.
2. npm proxy endpoint + `refuse proxy enable npm`.
3. `refuse doctor` proxy probe.
4. The remaining ecosystems, one per PR.
