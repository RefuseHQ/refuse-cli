package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/gate"
	"github.com/RefuseHQ/refuse-cli/internal/parsers"
	"github.com/RefuseHQ/refuse-cli/internal/server"
)

// realPipGateCmd is the entry point the Python hook
// (internal/pyhook/refuse_pip_guard.py) shells out to from inside the
// monkey-patched pip.InstallCommand.run(). The trailing `--` separates
// our flags (none currently) from pip's positional install specs.
//
//	refuse pip-gate -- requests==2.32.3 pyyaml==5.3
//
// exits 2 on block, 0 on allow, fail-open + 0 on transient errors so the
// Python tooling isn't broken by a flaky refuse.
func realPipGateCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "pip-gate [-- <pip install args>...]",
		Short:              "Gate a `python -m pip install` invocation (called by refuse_pip_guard.py)",
		Hidden:             true, // not for direct user invocation
		DisableFlagParsing: true, // pass argv straight through to the pip parser
		RunE: func(cmd *cobra.Command, args []string) error {
			// Drop a leading "--" if present (the Python hook passes one
			// to make argv flow-through unambiguous).
			if len(args) > 0 && args[0] == "--" {
				args = args[1:]
			}
			// Reuse the existing pip parser by synthesising the verb.
			pipArgs := append([]string{"install"}, args...)
			result := parsers.Pip().Parse(pipArgs)
			if !result.IsInstall || len(result.Packages) == 0 {
				return nil // nothing to gate
			}

			cfg, err := config.Load()
			if err != nil {
				// Fail open: surface to stderr but exit 0 so pip proceeds.
				fmt.Fprintln(os.Stderr, "refuse: config error:", err.Error())
				return nil
			}

			eng := gate.Engine{
				Client: server.New(cfg.ServerURL, cfg.APIKey),
				Policy: cfg.Policy,
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			res := eng.Decide(ctx, result)
			if res.Decision == gate.DecisionBlock {
				if res.Message != "" {
					fmt.Fprint(os.Stderr, res.Message)
				}
				os.Exit(2)
			}
			if res.Message != "" {
				fmt.Fprintln(os.Stderr, res.Message)
			}
			return nil
		},
	}
}
