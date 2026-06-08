# Changelog

## [1.3.4](https://github.com/RefuseHQ/refuse-cli/compare/v1.3.3...v1.3.4) (2026-06-08)


### Bug Fixes

* **shim:** windows shims now intercept package managers ([#51](https://github.com/RefuseHQ/refuse-cli/issues/51)) ([c2621a4](https://github.com/RefuseHQ/refuse-cli/commit/c2621a4e3527f95d88a3a6ece2bb883f90db910f))

## [1.3.3](https://github.com/RefuseHQ/refuse-cli/compare/v1.3.2...v1.3.3) (2026-06-08)


### Bug Fixes

* **shim:** prevent libuv signal-handler race on Windows; cross-platform install hint ([#49](https://github.com/RefuseHQ/refuse-cli/issues/49)) ([b846f19](https://github.com/RefuseHQ/refuse-cli/commit/b846f19fb46270aefc8936cef7c991fc7b96de5b))

## [1.3.2](https://github.com/RefuseHQ/refuse-cli/compare/v1.3.1...v1.3.2) (2026-06-08)


### Bug Fixes

* **install.ps1:** update $env:PATH in current session so refuse runs immediately ([#47](https://github.com/RefuseHQ/refuse-cli/issues/47)) ([c75c68f](https://github.com/RefuseHQ/refuse-cli/commit/c75c68ff0e168d1fcc298e6bcc5ac148536d78aa))
* **install.sh:** write PATH export to shell rc so refuse runs after install ([#46](https://github.com/RefuseHQ/refuse-cli/issues/46)) ([5cf9df5](https://github.com/RefuseHQ/refuse-cli/commit/5cf9df566d7226276711be762ed5d4f8b13fcd6c))

## [Unreleased]

### Fixed
- **Shim no longer recurses through its own symlink.** `SelfDir()` resolved through the symlink and returned the dir of the real refuse binary (e.g. `/opt/homebrew/Caskroom/refuse/<version>`); `findReal` only skipped that dir, so when invoked via the canonical `~/.refuse/bin/<mgr>` symlink it would pick the shim itself as "real npm" and recurse. Visible symptoms: `--no-refuse` was stripped but the gate still fired on the recursive call; `pnpm build` / `go build` would hang. Fixed by skipping both the launched symlink dir and the resolved target dir, plus defence-in-depth against external symlinks pointing back at the refuse binary. ([#41](https://github.com/RefuseHQ/refuse-cli/pull/41))

## [1.3.1](https://github.com/RefuseHQ/refuse-cli/compare/v1.3.0...v1.3.1) (2026-06-06)


### Bug Fixes

* **shim:** --no-refuse and other PATH lookups recursed into the shim ([#41](https://github.com/RefuseHQ/refuse-cli/issues/41)) ([1f19e68](https://github.com/RefuseHQ/refuse-cli/commit/1f19e68eed47371be3782e1163a92afef4d1b57c))

## [1.3.0](https://github.com/RefuseHQ/refuse-cli/compare/v1.2.3...v1.3.0) (2026-06-03)


### Features

* **allowlist:** per-project .refuse.yaml for formally-accepted risks ([#37](https://github.com/RefuseHQ/refuse-cli/issues/37)) ([59a6984](https://github.com/RefuseHQ/refuse-cli/commit/59a6984edea32bf955d66db26bda991653f0a24d))
* **audit:** one-shot repo scan — lockfiles + Dockerfiles + GH workflows ([#35](https://github.com/RefuseHQ/refuse-cli/issues/35)) ([c9f0c01](https://github.com/RefuseHQ/refuse-cli/commit/c9f0c015694c8093560b0763959d34b2db9afabf))
* **enforcement:** pre-commit hook + GH Action + --filename override ([#33](https://github.com/RefuseHQ/refuse-cli/issues/33)) ([39dc986](https://github.com/RefuseHQ/refuse-cli/commit/39dc986cb56d6a23b7025687dfed232919d9c47b))
* **fix:** print the suggested safe version (pipeable) ([#38](https://github.com/RefuseHQ/refuse-cli/issues/38)) ([ba864f1](https://github.com/RefuseHQ/refuse-cli/commit/ba864f17a3043323eede441209c558f68a598ed6))
* **parsers:** wider shim coverage — uv, poetry, pipenv, pdm, pipx, bundle, npx, composer, dotnet ([#32](https://github.com/RefuseHQ/refuse-cli/issues/32)) ([be3ef9a](https://github.com/RefuseHQ/refuse-cli/commit/be3ef9adc0eed64d45c3e7b355d622fd7d372721))
* **parsers:** wire up cargo, gem, and go shims ([#28](https://github.com/RefuseHQ/refuse-cli/issues/28)) ([8524256](https://github.com/RefuseHQ/refuse-cli/commit/852425606c4c66b5cb1e47eee6343d2b7b45c064))
* **proxy:** scaffold + design — local registry proxy (preview) ([#39](https://github.com/RefuseHQ/refuse-cli/issues/39)) ([39663a3](https://github.com/RefuseHQ/refuse-cli/commit/39663a32465c310ab3d61a6b0cb487bdee7b18f5))
* **python-hook:** --all flag + list subcommand + interpreter discovery ([#36](https://github.com/RefuseHQ/refuse-cli/issues/36)) ([7294bf3](https://github.com/RefuseHQ/refuse-cli/commit/7294bf3ed0f34a8a237a91f7b996c1a20ef328aa))
* **python-hook:** close the python -m pip bypass via a .pth interceptor ([#34](https://github.com/RefuseHQ/refuse-cli/issues/34)) ([adb72f6](https://github.com/RefuseHQ/refuse-cli/commit/adb72f6d78143d2bb6a6846066547b04ed2395df))
* **shim:** add inline --no-refuse bypass flag ([#29](https://github.com/RefuseHQ/refuse-cli/issues/29)) ([47c74d3](https://github.com/RefuseHQ/refuse-cli/commit/47c74d328c3807aa7cb3e6346c47748e32dca183))


### Bug Fixes

* **client:** bump HTTP timeout 1.5s → 5s, REFUSE_TIMEOUT_MS override ([#27](https://github.com/RefuseHQ/refuse-cli/issues/27)) ([ebf1c1e](https://github.com/RefuseHQ/refuse-cli/commit/ebf1c1e67d44d574dac09baa948e44e65e86b86a))

## [1.2.3](https://github.com/RefuseHQ/refuse-cli/compare/v1.2.2...v1.2.3) (2026-05-19)


### Bug Fixes

* **sign:** also fetch the G2 Developer ID intermediate ([#25](https://github.com/RefuseHQ/refuse-cli/issues/25)) ([3f9442b](https://github.com/RefuseHQ/refuse-cli/commit/3f9442bb4bd8094a859bb258f4f799641fd55f1d))

## [1.2.2](https://github.com/RefuseHQ/refuse-cli/compare/v1.2.1...v1.2.2) (2026-05-19)


### Bug Fixes

* **sign:** embed Developer ID intermediate so Apple notary accepts ([#22](https://github.com/RefuseHQ/refuse-cli/issues/22)) ([26b67f7](https://github.com/RefuseHQ/refuse-cli/commit/26b67f780db442aa4bda7992843f5a60c2430ada))

## [1.2.1](https://github.com/RefuseHQ/refuse-cli/compare/v1.2.0...v1.2.1) (2026-05-19)


### Bug Fixes

* **release:** switch macOS sign+notarize to rcodesign hooks ([#20](https://github.com/RefuseHQ/refuse-cli/issues/20)) ([2fef566](https://github.com/RefuseHQ/refuse-cli/commit/2fef566dcb4e31c55b47cf2d65717f72ee55de34))

## [1.2.0](https://github.com/RefuseHQ/refuse-cli/compare/v1.1.1...v1.2.0) (2026-05-19)


### Features

* **release:** code-sign + notarize macOS binaries with Apple Developer ID ([#18](https://github.com/RefuseHQ/refuse-cli/issues/18)) ([aa0c512](https://github.com/RefuseHQ/refuse-cli/commit/aa0c5124e93a14513a4f28abd923b0c19ca88897))

## [1.1.1](https://github.com/RefuseHQ/refuse-cli/compare/v1.1.0...v1.1.1) (2026-05-19)


### Bug Fixes

* **release:** wire the cross-repo PAT into the cask push ([#16](https://github.com/RefuseHQ/refuse-cli/issues/16)) ([676df14](https://github.com/RefuseHQ/refuse-cli/commit/676df144ef27a0cf4f98873bbfb913a64272ca14))

## [1.1.0](https://github.com/RefuseHQ/refuse-cli/compare/v1.0.0...v1.1.0) (2026-05-19)


### Features

* **cli:** shim install, hook adapters, check / status / doctor / init ([3fb6ae5](https://github.com/RefuseHQ/refuse-cli/commit/3fb6ae5f37c308456edc0bb4a405c0a84482766b))
* **gate:** parsers + decision engine + stdin hook protocol ([48f7450](https://github.com/RefuseHQ/refuse-cli/commit/48f7450dea916cbd200546f3e4727ac278fec08b))
* **install:** add Windows one-line installer (PowerShell) ([#9](https://github.com/RefuseHQ/refuse-cli/issues/9)) ([d358005](https://github.com/RefuseHQ/refuse-cli/commit/d35800506c9a988d9d9739cb00e6c304da1cf2da))
* **release:** automate version + tag via release-please ([#8](https://github.com/RefuseHQ/refuse-cli/issues/8)) ([f8586b4](https://github.com/RefuseHQ/refuse-cli/commit/f8586b4e37519afd2b346ebad999ef6a66d51719))


### Bug Fixes

* **lint:** migrate .golangci.yaml to v2 schema + bump action to v9 ([#10](https://github.com/RefuseHQ/refuse-cli/issues/10)) ([8166850](https://github.com/RefuseHQ/refuse-cli/commit/81668502b451239917089779ab4365b2f36a2a95))

## 1.0.0 (2026-05-19)


### Features

* **cli:** shim install, hook adapters, check / status / doctor / init ([3fb6ae5](https://github.com/RefuseHQ/refuse-cli/commit/3fb6ae5f37c308456edc0bb4a405c0a84482766b))
* **gate:** parsers + decision engine + stdin hook protocol ([48f7450](https://github.com/RefuseHQ/refuse-cli/commit/48f7450dea916cbd200546f3e4727ac278fec08b))
* **install:** add Windows one-line installer (PowerShell) ([#9](https://github.com/RefuseHQ/refuse-cli/issues/9)) ([d358005](https://github.com/RefuseHQ/refuse-cli/commit/d35800506c9a988d9d9739cb00e6c304da1cf2da))
* **release:** automate version + tag via release-please ([#8](https://github.com/RefuseHQ/refuse-cli/issues/8)) ([f8586b4](https://github.com/RefuseHQ/refuse-cli/commit/f8586b4e37519afd2b346ebad999ef6a66d51719))


### Bug Fixes

* **lint:** migrate .golangci.yaml to v2 schema + bump action to v9 ([#10](https://github.com/RefuseHQ/refuse-cli/issues/10)) ([8166850](https://github.com/RefuseHQ/refuse-cli/commit/81668502b451239917089779ab4365b2f36a2a95))

## Changelog

All notable changes to refuse-cli are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once it reaches 1.0.

<!-- versions below this line are managed by release-please -->
