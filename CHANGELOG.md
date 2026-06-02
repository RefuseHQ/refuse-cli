# Changelog

## [1.3.0](https://github.com/RefuseHQ/refuse-cli/compare/v1.2.3...v1.3.0) (2026-06-02)


### Features

* **parsers:** wire up cargo, gem, and go shims ([#28](https://github.com/RefuseHQ/refuse-cli/issues/28)) ([8524256](https://github.com/RefuseHQ/refuse-cli/commit/852425606c4c66b5cb1e47eee6343d2b7b45c064))
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
