package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/RefuseHQ/refuse-cli/internal/pyhook"
)

func realPythonHookCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "python-hook",
		Short: "Manage the `python -m pip` interception hook",
		Long: `Installs a .pth + module into a Python env's site-packages so that
` + "`python -m pip install ...`" + ` (which bypasses the PATH shim) is also
routed through the refuse gate.

The hook patches pip's InstallCommand at interpreter startup to shell
out to ` + "`refuse pip-gate <args>`" + ` before running the real install.
Fails open on any error so unrelated Python tooling isn't broken if
refuse is unreachable.`,
	}
	cmd.AddCommand(pythonHookInstallCmd())
	cmd.AddCommand(pythonHookUninstallCmd())
	cmd.AddCommand(pythonHookStatusCmd())
	return cmd
}

func pythonHookInstallCmd() *cobra.Command {
	var target string
	var pythonBin string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Drop the hook into a Python env's site-packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			site, err := resolveSiteDir(target, pythonBin)
			if err != nil {
				return err
			}
			modulePath, pthPath, err := pyhook.Install(site)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "refuse: installed pip hook in %s\n  %s\n  %s\n", site, modulePath, pthPath)
			fmt.Fprintln(cmd.OutOrStdout(), "Open a new Python interpreter or restart your tooling to pick it up.")
			return nil
		},
	}
	cmd.Flags().StringVar(&target, "target", "", "site-packages directory to install into (default: auto-detect via --python)")
	cmd.Flags().StringVar(&pythonBin, "python", "", "python interpreter to query for site-packages (default: python3 on PATH)")
	return cmd
}

func pythonHookUninstallCmd() *cobra.Command {
	var target string
	var pythonBin string
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the hook from a Python env's site-packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			site, err := resolveSiteDir(target, pythonBin)
			if err != nil {
				return err
			}
			removed, err := pyhook.Uninstall(site)
			if err != nil {
				return err
			}
			if len(removed) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "refuse: no hook files found in %s\n", site)
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "refuse: removed from %s:\n", site)
			for _, p := range removed {
				fmt.Fprintln(cmd.OutOrStdout(), "  "+p)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&target, "target", "", "site-packages directory to uninstall from")
	cmd.Flags().StringVar(&pythonBin, "python", "", "python interpreter to query (default: python3 on PATH)")
	return cmd
}

func pythonHookStatusCmd() *cobra.Command {
	var target string
	var pythonBin string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show whether the hook is installed in a Python env",
		RunE: func(cmd *cobra.Command, args []string) error {
			site, err := resolveSiteDir(target, pythonBin)
			if err != nil {
				return err
			}
			moduleOK, pthOK := pyhook.Status(site)
			fmt.Fprintf(cmd.OutOrStdout(), "site-packages:  %s\n", site)
			fmt.Fprintf(cmd.OutOrStdout(), "  %-24s %s\n", pyhook.ModuleName, presenceLabel(moduleOK))
			fmt.Fprintf(cmd.OutOrStdout(), "  %-24s %s\n", pyhook.PthName, presenceLabel(pthOK))
			if moduleOK && pthOK {
				fmt.Fprintln(cmd.OutOrStdout(), "hook is active for `python -m pip install`")
			} else {
				fmt.Fprintln(cmd.OutOrStdout(), "hook NOT active — run `refuse python-hook install`")
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&target, "target", "", "site-packages directory to check")
	cmd.Flags().StringVar(&pythonBin, "python", "", "python interpreter to query (default: python3 on PATH)")
	return cmd
}

func presenceLabel(ok bool) string {
	if ok {
		return "installed"
	}
	return "missing"
}

// resolveSiteDir picks the site-packages directory to act on. Explicit
// --target wins; otherwise we ask the python interpreter via --python (or
// `python3` by default) where its site-packages lives.
func resolveSiteDir(target, pythonBin string) (string, error) {
	if target != "" {
		return target, nil
	}
	site, err := pyhook.SitePackages(pythonBin)
	if err != nil {
		return "", fmt.Errorf("locate site-packages: %w (pass --target=<dir> to override)", err)
	}
	return site, nil
}

