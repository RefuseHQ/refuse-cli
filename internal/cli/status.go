package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/hook"
	"github.com/RefuseHQ/refuse-cli/internal/shim"
)

func realStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show CLI install state, server URL, and recent activity",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load()
			home, _ := os.UserHomeDir()
			binDir := filepath.Join(home, ".refuse", "bin")

			fmt.Fprintf(cmd.OutOrStdout(), "server:  %s\n", cfg.ServerURL)
			fmt.Fprintf(cmd.OutOrStdout(), "api_key: %s\n", presence(cfg.APIKey))
			fmt.Fprintf(cmd.OutOrStdout(), "shims:   %s\n", binDir)

			// Per-manager shim state
			selfExe, _ := os.Executable()
			if resolved, err := filepath.EvalSymlinks(selfExe); err == nil {
				selfExe = resolved
			}
			for _, name := range shim.DefaultShims {
				target := filepath.Join(binDir, name)
				link, err := os.Readlink(target)
				switch {
				case os.IsNotExist(err):
					fmt.Fprintf(cmd.OutOrStdout(), "  %-6s  not installed\n", name)
				case err != nil:
					fmt.Fprintf(cmd.OutOrStdout(), "  %-6s  %v\n", name, err)
				case link != selfExe:
					fmt.Fprintf(cmd.OutOrStdout(), "  %-6s  installed (foreign symlink: %s)\n", name, link)
				default:
					fmt.Fprintf(cmd.OutOrStdout(), "  %-6s  installed\n", name)
				}
			}

			fmt.Fprintln(cmd.OutOrStdout(), "hooks:")
			for _, a := range hook.All() {
				st, _ := a.Status()
				state := "not installed"
				if st.Installed {
					state = "installed"
				} else if !st.Found {
					state = "agent config not found"
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %-12s  %s\n", a.Name(), state)
			}

			// Server reachability (best-effort).
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", cfg.ServerURL+"/healthz", nil)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "reach:   unreachable (%s)\n", err.Error())
			} else {
				res.Body.Close()
				fmt.Fprintf(cmd.OutOrStdout(), "reach:   HTTP %d from %s/healthz\n", res.StatusCode, cfg.ServerURL)
			}
			return nil
		},
	}
}

func presence(s string) string {
	if s == "" {
		return "(unset)"
	}
	return "set"
}
