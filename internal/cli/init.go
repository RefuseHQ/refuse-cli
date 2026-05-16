package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/spf13/cobra"
)

func realInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "First-time setup wizard (server URL + API key)",
		Long: `Interactive — asks for the refuse server URL and (optionally) an API key,
then persists them to ~/.refuse/config.yaml. Safe to re-run: existing values
become the default for each prompt.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, _ := config.Load()
			r := bufio.NewReader(os.Stdin)

			defServer := cfg.ServerURL
			if defServer == "" {
				defServer = config.DefaultServerURL
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Server URL [%s]: ", defServer)
			line, _ := r.ReadString('\n')
			if v := strings.TrimSpace(line); v != "" {
				cfg.ServerURL = v
			} else {
				cfg.ServerURL = defServer
			}

			maskedExisting := ""
			if cfg.APIKey != "" {
				maskedExisting = maskKey(cfg.APIKey)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "API key (rfs_…) [%s]: ", presenceForInit(maskedExisting))
			line, _ = r.ReadString('\n')
			if v := strings.TrimSpace(line); v != "" {
				cfg.APIKey = v
			}

			if err := config.Save(cfg); err != nil {
				return err
			}
			p, _ := config.UserConfigPath()
			fmt.Fprintf(cmd.OutOrStdout(), "\nsaved %s\n", p)
			fmt.Fprintln(cmd.OutOrStdout(), "Next: `refuse install` to drop PATH shims, then `refuse hook install claude-code`.")
			return nil
		},
	}
}

func presenceForInit(s string) string {
	if s == "" {
		return "leave empty to skip"
	}
	return s
}
