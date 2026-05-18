# Changelog

All notable changes to refuse-cli are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once it reaches 1.0.

## [Unreleased]

### Added
- Community + governance: `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `CODEOWNERS`, issue + PR templates, `.editorconfig`, `ARCHITECTURE.md`, `ROADMAP.md`.
- `.golangci.yaml` and a `lint` workflow.
- Dependabot, CodeQL.
- Release signing (cosign) and SLSA provenance attestations.

### Changed
- README: hero, badges, ecosystem matrix, cross-links to `refuse` and `refuse.dev`.

## [0.0.1] — initial

- Multicall binary: invoking the binary as `npm` / `pip` / `yarn` / `cargo` / etc. routes to the shim, invoking it as `refuse` opens the CLI.
- `refuse gate` decision engine: parses an install command, queries the server, and exits 0 (allow) or 2 (block).
- Argv parsers: npm, pnpm, yarn (classic + Berry), pip, pip3, cargo, gem, bun, go.
- Claude Code PreToolUse hook integration.
- Goreleaser-based releases for darwin/linux/windows × amd64/arm64.
- Homebrew tap (`refusehq/tap/refuse`).
- `scripts/install.sh` with sha256 checksum verification.

[Unreleased]: https://github.com/RefuseHQ/refuse-cli/compare/v0.0.1...HEAD
[0.0.1]: https://github.com/RefuseHQ/refuse-cli/releases/tag/v0.0.1
