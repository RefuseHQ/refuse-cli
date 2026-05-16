package cli

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/hook"
	"github.com/spf13/cobra"
)

func realDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose PATH, hooks, and server reachability",
		RunE: func(cmd *cobra.Command, args []string) error {
			ok := true

			home, _ := os.UserHomeDir()
			binDir := filepath.Join(home, ".refuse", "bin")
			pathOK := false
			for _, dir := range strings.Split(os.Getenv("PATH"), ":") {
				if dir == binDir {
					pathOK = true
					break
				}
			}
			if pathOK {
				fmt.Fprintf(cmd.OutOrStdout(), "[OK ] PATH includes %s\n", binDir)
			} else {
				ok = false
				fmt.Fprintf(cmd.OutOrStdout(), "[!  ] PATH missing %s — run `refuse install` or open a new shell\n", binDir)
			}

			cfg, err := config.Load()
			if err != nil {
				ok = false
				fmt.Fprintf(cmd.OutOrStdout(), "[ERR] config load: %v\n", err)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "[OK ] config loaded (server %s)\n", cfg.ServerURL)
			}

			// Hooks
			for _, a := range hook.All() {
				st, _ := a.Status()
				if st.Installed {
					fmt.Fprintf(cmd.OutOrStdout(), "[OK ] %s hook installed\n", a.Name())
				} else if st.Found {
					fmt.Fprintf(cmd.OutOrStdout(), "[!  ] %s found but no refuse hook — `refuse hook install %s`\n", a.Name(), a.Name())
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "[--] %s not installed\n", a.Name())
				}
			}

			// Server reachability
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", cfg.ServerURL+"/healthz", nil)
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				ok = false
				fmt.Fprintf(cmd.OutOrStdout(), "[ERR] server unreachable: %v\n", err)
			} else {
				res.Body.Close()
				if res.StatusCode == 200 {
					fmt.Fprintf(cmd.OutOrStdout(), "[OK ] server %s healthy\n", cfg.ServerURL)
				} else {
					ok = false
					fmt.Fprintf(cmd.OutOrStdout(), "[!  ] server %s returned HTTP %d\n", cfg.ServerURL, res.StatusCode)
				}
			}

			if !ok {
				os.Exit(1)
			}
			return nil
		},
	}
}
