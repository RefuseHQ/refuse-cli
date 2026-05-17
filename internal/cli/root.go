// Package cli builds the cobra command tree for the `refuse` binary.
package cli

import (
	"github.com/RefuseHQ/refuse-cli/internal/version"
	"github.com/spf13/cobra"
)

// NewRoot returns the root cobra command with all subcommands attached.
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "refuse",
		Short:         "Gate AI coding agents from installing vulnerable packages",
		Long:          "refuse wraps your package managers (npm, pip, cargo, yarn, pnpm, gem,\nbun, go) and refuses to install packages with known CVEs. Works on a\nlaptop, in CI, and inside Docker build stages.",
		Version:       version.Version + " (commit " + version.Commit + ", " + version.Date + ")",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.SetVersionTemplate("{{.Version}}\n")

	// Subcommands — implemented in sibling files of this package as they
	// land. Each module exposes a New<Cmd>() *cobra.Command builder.
	root.AddCommand(realStatusCmd())
	root.AddCommand(realCheckCmd())
	root.AddCommand(realCheckLockfileCmd())
	root.AddCommand(realConfigCmd())
	root.AddCommand(realInstallCmd())
	root.AddCommand(realUninstallCmd())
	root.AddCommand(realHookCmd())
	root.AddCommand(realGateCmd())
	root.AddCommand(realInitCmd())
	root.AddCommand(realDoctorCmd())

	return root
}
