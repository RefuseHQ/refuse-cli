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
		Long:          "refuse wraps your package managers and installs deterministic hooks into coding\nagents so vulnerable packages get blocked at install time. See https://refuse.dev.",
		Version:       version.Version + " (commit " + version.Commit + ", " + version.Date + ")",
		SilenceUsage:  true,
		SilenceErrors: false,
	}
	root.SetVersionTemplate("{{.Version}}\n")

	// Subcommands — implemented in sibling files of this package as they
	// land. Each module exposes a New<Cmd>() *cobra.Command builder.
	root.AddCommand(newStatusCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newCheckLockfileCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newInstallCmd())
	root.AddCommand(newUninstallCmd())
	root.AddCommand(newHookCmd())
	root.AddCommand(newGateCmd())
	root.AddCommand(newInitCmd())
	root.AddCommand(newDoctorCmd())

	return root
}
