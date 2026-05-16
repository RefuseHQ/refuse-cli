package cli

import (
	"fmt"

	"github.com/RefuseHQ/refuse-cli/internal/shim"
	"github.com/spf13/cobra"
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

By default we install shims for: npm, pnpm, yarn, bun, pip, pip3.
Pass --shims=npm,pip to scope.`,
		RunE: func(cmd *cobra.Command, args []string) error {
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
			fmt.Fprintln(cmd.OutOrStdout(), "\nOpen a new shell (or run `exec $SHELL`) to pick up the PATH update.")
			return nil
		},
	}
	cmd.Flags().StringSliceVar(&shims, "shims", nil, "Comma-separated list of managers (default: npm,pnpm,yarn,bun,pip,pip3)")
	return cmd
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
