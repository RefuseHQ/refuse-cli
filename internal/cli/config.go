package cli

import (
	"encoding/json"
	"fmt"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/spf13/cobra"
)

func realConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Show or update the resolved CLI config (~/.refuse/config.yaml)",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "Print the resolved config",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			// Mask the API key — config show is fine for screenshots only when
			// secrets stay hidden by default.
			if cfg.APIKey != "" {
				cfg.APIKey = maskKey(cfg.APIKey)
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(cfg)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config key (server_url, api_key, policy.severity_threshold, policy.fail_closed)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			switch args[0] {
			case "server_url":
				cfg.ServerURL = args[1]
			case "api_key":
				cfg.APIKey = args[1]
			case "policy.severity_threshold":
				cfg.Policy.SeverityThreshold = args[1]
			case "policy.fail_closed":
				cfg.Policy.FailClosed = args[1] == "true" || args[1] == "1"
			default:
				return fmt.Errorf("unknown key %q", args[0])
			}
			if err := config.Save(cfg); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "saved %s\n", args[0])
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "get <key>",
		Short: "Print a single config value",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			switch args[0] {
			case "server_url":
				fmt.Fprintln(cmd.OutOrStdout(), cfg.ServerURL)
			case "api_key":
				if cfg.APIKey == "" {
					fmt.Fprintln(cmd.OutOrStdout(), "(unset)")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), maskKey(cfg.APIKey))
				}
			case "policy.severity_threshold":
				fmt.Fprintln(cmd.OutOrStdout(), cfg.Policy.SeverityThreshold)
			case "policy.fail_closed":
				fmt.Fprintln(cmd.OutOrStdout(), cfg.Policy.FailClosed)
			default:
				return fmt.Errorf("unknown key %q", args[0])
			}
			return nil
		},
	})

	return cmd
}

func maskKey(k string) string {
	if len(k) <= 10 {
		return "***"
	}
	return k[:4] + "…" + k[len(k)-2:]
}
