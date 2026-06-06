package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/shim"
)

func realInstallCmd() *cobra.Command {
	var shims []string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install shims into ~/.refuse/bin and add it to PATH",
		Long: `Drop symlinks named after your package managers into ~/.refuse/bin and
prepend that directory to PATH via your shell's rc file (.bashrc / .zshrc /
.profile / fish conf.d). The shims dispatch back into the refuse binary
which vets each install before exec'ing the real package manager.

If no ~/.refuse/config.yaml exists, install writes the default one (hosted
server, anonymous, fail-open) so the shims are usable immediately. Run
` + "`refuse init`" + ` afterwards to attach an API key or switch servers.

By default we install shims for the full supported set; pass --shims=npm,pip
to scope.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Materialize a default config if the user hasn't run `refuse init`
			// yet. Keeps the install snippet on refuse.dev one step shorter and
			// makes `cat ~/.refuse/config.yaml` show what server the shims hit.
			wroteDefault, defaultPath, cfgErr := writeDefaultConfigIfMissing()
			if cfgErr != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "refuse: skipped writing default config: %v\n", cfgErr)
			}

			res, err := shim.Install(shims)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "refuse: bin dir %s\n", res.BinDir)
			if len(res.Installed) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  installed shims: %v\n", res.Installed)
			}
			for _, s := range res.Skipped {
				fmt.Fprintf(cmd.OutOrStdout(), "  skipped: %s\n", s)
			}
			for _, p := range res.ShellRC {
				fmt.Fprintf(cmd.OutOrStdout(), "  updated %s\n", p)
			}
			if wroteDefault {
				fmt.Fprintf(cmd.OutOrStdout(), "  wrote default config: %s (hosted server, anonymous)\n", defaultPath)
				fmt.Fprintln(cmd.OutOrStdout(), "    → run `refuse init` to attach an API key or switch servers")
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nOpen a new shell (or run `exec $SHELL`) to pick up the PATH update.")
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&shims, "shims", nil, "Comma-separated list of managers (default: the full supported set)")
	return cmd
}

// writeDefaultConfigIfMissing stats ~/.refuse/config.yaml; if it doesn't
// exist, writes config.Defaults() there so the shims have something to read.
// Returns whether it actually wrote the file, the path on disk, and any
// non-fatal error (e.g. UserDir lookup failure).
func writeDefaultConfigIfMissing() (wrote bool, path string, err error) {
	p, perr := config.UserConfigPath()
	if perr != nil {
		return false, "", perr
	}
	if _, statErr := os.Stat(p); statErr == nil {
		// Already there — nothing to do.
		return false, p, nil
	} else if !errors.Is(statErr, os.ErrNotExist) {
		return false, p, statErr
	}
	if err := config.Save(config.Defaults()); err != nil {
		return false, p, err
	}
	return true, p, nil
}

func realUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove shims and revert shell-rc PATH edits",
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := shim.Uninstall()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "refuse: bin dir %s\n", res.BinDir)
			if len(res.Installed) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "  removed shims: %v\n", res.Installed)
			}
			for _, s := range res.Skipped {
				fmt.Fprintf(cmd.OutOrStdout(), "  skipped: %s\n", s)
			}
			for _, p := range res.ShellRC {
				fmt.Fprintf(cmd.OutOrStdout(), "  cleaned %s\n", p)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "\nOpen a new shell for the PATH change to take effect.")
			return nil
		},
	}
}
