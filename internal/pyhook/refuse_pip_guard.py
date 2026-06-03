"""refuse_pip_guard — gate `python -m pip install` through the refuse CLI.

Installed by `refuse python-hook install` into a Python env's site-packages,
alongside a `.pth` file that imports this module at interpreter startup.
At import time we patch pip's InstallCommand.run() to first shell out to
`refuse pip-gate` with the same install args. If refuse blocks (exit 2),
the pip install is aborted before any code is fetched. On any other
condition we transparently delegate to pip's real install — fail-open
matches the shim's behavior and avoids breaking unrelated Python tooling.

Safe to no-op when:
  - `refuse` binary isn't on PATH (probably uninstalled)
  - REFUSE_NO_GATE=1 is set
  - pip can't be imported / its internals shifted under our feet
  - we're already inside the gated subprocess (REFUSE_PIP_GUARD_ACTIVE=1)
"""

from __future__ import annotations

import os
import shutil
import subprocess
import sys

_DISABLED_ENV = "REFUSE_NO_GATE"
_REENTRY_ENV = "REFUSE_PIP_GUARD_ACTIVE"


def _enable() -> None:
    if os.environ.get(_DISABLED_ENV) == "1":
        return
    if os.environ.get(_REENTRY_ENV) == "1":
        # Avoid recursion if the gate itself shells back into pip.
        return

    refuse = shutil.which("refuse")
    if not refuse:
        return

    try:
        from pip._internal.commands.install import InstallCommand  # type: ignore
    except Exception:
        return

    original_run = InstallCommand.run

    def gated_run(self, options, args):  # type: ignore[no-untyped-def]
        # `args` is the positional install spec list pip parsed
        # (`["requests==2.32.3"]` for `pip install requests==2.32.3`).
        # We pass them straight through to the refuse gate.
        try:
            env = dict(os.environ)
            env[_REENTRY_ENV] = "1"
            result = subprocess.run(
                [refuse, "pip-gate", "--"] + list(args),
                env=env,
                check=False,
            )
            if result.returncode == 2:
                # Block. refuse already printed the explanation to stderr.
                return 2
        except Exception:
            # Fail open — don't punish unrelated Python tooling for a
            # transient refuse problem.
            pass
        return original_run(self, options, args)

    InstallCommand.run = gated_run  # type: ignore[assignment]


_enable()
