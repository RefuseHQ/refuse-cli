package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// realProxyCmd is the user-facing scaffold for the upcoming registry
// proxy feature. The architecture is documented at
// docs/design/registry-proxy.md.
//
// Subcommands are reserved so the command surface is discoverable via
// `refuse proxy --help`. Each subcommand currently prints a not-yet-
// implemented note pointing at the design doc; real behaviour lands in
// follow-up PRs per ecosystem (PyPI first, then npm).
func realProxyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proxy",
		Short: "Route package-manager fetches through refuse (preview)",
		Long: `Reserved command tree for the upcoming local registry proxy.

The proxy is the only interception layer that catches every install
regardless of how it was invoked — including ` + "`python -m pip`" + `, uv,
conda, Docker build layers, and IDE-driven installs. It works by
pointing each package manager's index URL at a refuse-managed proxy;
the proxy gates the lookup and forwards to the upstream on allow.

The full design lives in docs/design/registry-proxy.md.

Subcommands are scaffolded today; real implementations land per
ecosystem in follow-up releases (PyPI first).`,
	}
	cmd.AddCommand(proxyEnableCmd())
	cmd.AddCommand(proxyDisableCmd())
	cmd.AddCommand(proxyStatusCmd())
	return cmd
}

func proxyEnableCmd() *cobra.Command {
	var url string
	cmd := &cobra.Command{
		Use:   "enable <ecosystem>",
		Short: "Route an ecosystem's installs through the refuse proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(),
				"refuse proxy: %s routing not yet implemented (target URL %q recorded).\n"+
					"  See docs/design/registry-proxy.md for the design + ETA.\n",
				args[0], url)
			return nil
		},
	}
	cmd.Flags().StringVar(&url, "url", "", "Override the proxy URL (defaults to the hosted refuse proxy once available)")
	return cmd
}

func proxyDisableCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "disable <ecosystem>",
		Short: "Stop routing an ecosystem's installs through refuse",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintf(cmd.OutOrStdout(),
				"refuse proxy: %s routing not yet implemented — nothing to revert.\n"+
					"  See docs/design/registry-proxy.md.\n",
				args[0])
			return nil
		},
	}
}

func proxyStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show which ecosystems are routed through the proxy",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.OutOrStdout(),
				"refuse proxy: no ecosystems routed (feature in development).\n"+
					"See docs/design/registry-proxy.md for the design + per-ecosystem rollout plan.")
			return nil
		},
	}
}
