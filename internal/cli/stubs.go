// Subcommand stubs — placeholder implementations so the binary compiles and
// `refuse --help` works. Each subcommand gets a dedicated file as it's
// implemented for real.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show CLI install state, server URL, and recent activity",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "status: not implemented yet")
			return nil
		},
	}
}

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <ecosystem> <package>[@version]",
		Short: "Check a single package against the refuse server",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "check: not implemented yet")
			return nil
		},
	}
}

func newCheckLockfileCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check-lockfile <path>",
		Short: "Parse and check a lockfile (package-lock.json, requirements.txt, ...)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "check-lockfile: not implemented yet")
			return nil
		},
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or update the resolved CLI config (~/.refuse/config.yaml)",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the resolved config",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "config show: not implemented yet")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config key",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "config set: not implemented yet")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "get <key>",
		Short: "Print a single config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "config get: not implemented yet")
			return nil
		},
	})
	return cmd
}

func newInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install shims into ~/.refuse/bin and add it to PATH",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "install: not implemented yet")
			return nil
		},
	}
}

func newUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove shims and revert shell-rc edits",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "uninstall: not implemented yet")
			return nil
		},
	}
}

func newHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hook",
		Short: "Manage agent pre-tool-use hooks",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "install <agent>",
		Short: "Install a pre-tool-use hook for an agent (claude-code, …)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "hook install: not implemented yet")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "remove <agent>",
		Short: "Remove a previously-installed hook",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "hook remove: not implemented yet")
			return nil
		},
	})
	cmd.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "Show installed hooks across known agents",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "hook list: not implemented yet")
			return nil
		},
	})
	return cmd
}

func newGateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "gate",
		Short: "Decision engine called by shims and agent hooks (stdin protocol)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "gate: not implemented yet")
			return nil
		},
	}
}

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "First-time setup wizard (server URL, API key, shims, hooks)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "init: not implemented yet")
			return nil
		},
	}
}

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose PATH, hook files, and server reachability",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(), "doctor: not implemented yet")
			return nil
		},
	}
}
