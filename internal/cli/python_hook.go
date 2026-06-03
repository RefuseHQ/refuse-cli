package cli

import (
	"fmt"
	"io"

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
	cmd.AddCommand(pythonHookListCmd())
	return cmd
}

func pythonHookListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Discover every Python interpreter on this machine",
		Long: `Scans PATH, common venv homes, pyenv, and conda for Python
interpreters refuse can install the pip hook into. Useful as a dry-run
before ` + "`refuse python-hook install --all`" + `.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			pythons := pyhook.DiscoverInterpreters()
			if len(pythons) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "refuse: no python interpreters found")
				return nil
			}
			for _, p := range pythons {
				site, err := pyhook.SitePackagesOf(p)
				if err != nil {
					fmt.Fprintf(cmd.OutOrStdout(), "%s    (unable to query site-packages: %v)\n", p, err)
					continue
				}
				moduleOK, pthOK := pyhook.Status(site)
				status := "missing"
				if moduleOK && pthOK {
					status = "installed"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "%-10s  %s\n            site: %s\n", status, p, site)
			}
			return nil
		},
	}
}

func pythonHookInstallCmd() *cobra.Command {
	var target string
	var pythonBin string
	var all bool
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Drop the hook into a Python env's site-packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				return pythonHookAllOp(cmd.OutOrStdout(), true)
			}
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
	cmd.Flags().BoolVar(&all, "all", false, "Install into every discoverable Python env (PATH + venvs + pyenv + conda)")
	return cmd
}

func pythonHookUninstallCmd() *cobra.Command {
	var target string
	var pythonBin string
	var all bool
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove the hook from a Python env's site-packages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if all {
				return pythonHookAllOp(cmd.OutOrStdout(), false)
			}
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
	cmd.Flags().BoolVar(&all, "all", false, "Uninstall from every discoverable Python env")
	return cmd
}

// pythonHookAllOp runs install (install=true) or uninstall over every
// discovered interpreter, reporting per-env status. Survives errors on
// individual envs.
func pythonHookAllOp(out io.Writer, install bool) error {
	pythons := pyhook.DiscoverInterpreters()
	if len(pythons) == 0 {
		fmt.Fprintln(out, "refuse: no python interpreters discovered")
		return nil
	}
	verb := "install"
	if !install {
		verb = "uninstall"
	}
	fmt.Fprintf(out, "refuse: %s into %d discovered Python env(s)\n", verb, len(pythons))
	ok, fail := 0, 0
	for _, p := range pythons {
		site, err := pyhook.SitePackagesOf(p)
		if err != nil {
			fmt.Fprintf(out, "  %s\n    skipped — %v\n", p, err)
			fail++
			continue
		}
		if install {
			if _, _, err := pyhook.Install(site); err != nil {
				fmt.Fprintf(out, "  %s\n    failed — %v\n", p, err)
				fail++
				continue
			}
			fmt.Fprintf(out, "  installed   %s\n", p)
		} else {
			removed, err := pyhook.Uninstall(site)
			if err != nil {
				fmt.Fprintf(out, "  %s\n    failed — %v\n", p, err)
				fail++
				continue
			}
			if len(removed) == 0 {
				fmt.Fprintf(out, "  not present %s\n", p)
				continue
			}
			fmt.Fprintf(out, "  removed     %s\n", p)
		}
		ok++
	}
	fmt.Fprintf(out, "Done: %d ok, %d failed\n", ok, fail)
	if fail > 0 {
		return fmt.Errorf("%d env(s) failed", fail)
	}
	return nil
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
