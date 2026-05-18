# Security Policy

## Reporting a vulnerability

refuse-cli sits in front of every `npm install`, `pip install`, etc. on machines that install it — a vulnerability here has unusually high blast radius. Please report it privately.

**Email:** [hello@refuse.dev](mailto:hello@refuse.dev) with the subject `[security] <short description>`.

Or use GitHub's [private vulnerability reporting](https://github.com/RefuseHQ/refuse-cli/security/advisories/new).

Please include:

- A description of the issue and the realistic impact.
- A reproduction — exact `refuse` command + state of `~/.refuse/`.
- The version (`refuse --version`) and OS/arch.
- A suggested fix, if you have one.

We aim to acknowledge within **48 hours** and ship a fix or mitigation within **90 days** for most issues. Issues that allow:

- Bypassing the gate (running a known-bad install without detection),
- Command injection through the wrapped package manager,
- Local privilege escalation via the shim's PATH manipulation,
- Arbitrary file write via the install or uninstall flow,

are priority-1 and we'll move on them inside a week.

Please do not:

- File public issues for security problems.
- Test exploits against any machine you don't own.
- Publish details before we've shipped a fix.

## Supported versions

refuse-cli is pre-1.0. We currently support security fixes only on the **latest tagged release** and `main`. Older versions should be upgraded.

## Verifying a release

Each release ships a `checksums.txt` plus a cosign keyless signature (`checksums.txt.sig` + `checksums.txt.pem`) tied to this repo's GitHub Actions OIDC identity. To verify before installing:

```sh
cosign verify-blob \
  --certificate-identity-regexp '^https://github.com/RefuseHQ/refuse-cli/' \
  --certificate-oidc-issuer 'https://token.actions.githubusercontent.com' \
  --certificate checksums.txt.pem \
  --signature checksums.txt.sig \
  checksums.txt
```

Then `shasum -a 256 -c checksums.txt` against the downloaded archive. SLSA build provenance is also attested via `actions/attest-build-provenance` and visible on the release page.

## In scope

- The shim install / uninstall flow (`refuse install`, `refuse uninstall`).
- PATH and shell-rc manipulation (`internal/shim/shellrc.go`).
- The gate's decision logic (`internal/gate/`).
- The agent hook installer (`internal/hook/`).
- The release pipeline — signed binaries, Homebrew formula, install script.
- `scripts/install.sh` — particularly checksum verification.

## Out of scope

- Reports against the upstream `refuse` server — those go to that [repo's SECURITY.md](https://github.com/RefuseHQ/refuse/blob/main/SECURITY.md).
- Issues that require the user to deliberately set `REFUSE_FAIL_CLOSED=0` (the default), `REFUSE_ALLOW_VULNERABLE=1`, or otherwise disable the gate.
- Findings against third-party dependencies that don't affect refuse-cli's behavior. Open a PR bumping the dep.

## Disclosure

Once a fix is shipped we publish a GitHub Security Advisory with a CVE where appropriate. Reporters are credited unless they ask to remain anonymous.

Thanks for keeping this safe.
