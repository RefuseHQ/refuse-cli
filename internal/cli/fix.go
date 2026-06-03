package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/parsers"
	"github.com/RefuseHQ/refuse-cli/internal/server"
)

func realFixCmd() *cobra.Command {
	var fixType string
	cmd := &cobra.Command{
		Use:   "fix <ecosystem> <package>[@<version>]",
		Short: "Print the suggested safe version for a vulnerable package",
		Long: `Looks up the package in the refuse server and prints the suggested
safe version to stdout. Pipe into your package manager to auto-bump:

  npm install lodash@$(refuse fix npm lodash@4.17.10)
  pip install pyyaml==$(refuse fix PyPI pyyaml@5.3)

Exit codes:
  0  fix found — version on stdout
  0  not vulnerable — empty stdout
  1  server / lookup error
  2  vulnerable but no fix available`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ecosystem := args[0]
			name, version := parsers.SplitNameVersionExt(args[1])
			if name == "" {
				return fmt.Errorf("missing package name")
			}
			if version == "" {
				version = "latest"
			}

			cfg, err := config.Load()
			if err != nil {
				return err
			}
			c := server.New(cfg.ServerURL, cfg.APIKey)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			resp, err := c.CheckPackage(ctx, server.CheckPackageRequest{
				Ecosystem: ecosystem,
				Name:      name,
				Version:   version,
			})
			if err != nil {
				return err
			}

			if !resp.Vulnerable || len(resp.SuggestedFixes) == 0 {
				if !resp.Vulnerable {
					// Not vulnerable — print nothing, exit 0.
					return nil
				}
				return fmt.Errorf("%s@%s is vulnerable but no fix version was suggested", name, version)
			}

			// Pick by type. minimum_safe is the default — smallest version
			// that escapes every matching advisory; almost always non-
			// breaking.
			fixType = strings.ToLower(fixType)
			for _, f := range resp.SuggestedFixes {
				if strings.EqualFold(f.Type, fixType) {
					fmt.Fprintln(cmd.OutOrStdout(), f.Version)
					return nil
				}
			}
			// Requested type wasn't returned; fall back to the first.
			fmt.Fprintln(cmd.OutOrStdout(), resp.SuggestedFixes[0].Version)
			return nil
		},
	}
	cmd.Flags().StringVar(&fixType, "type", "minimum_safe",
		"Which suggestion to print: minimum_safe | latest_stable | latest")
	return cmd
}
