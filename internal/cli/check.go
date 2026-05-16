package cli

import (
	"context"
	"encoding/json"
	"os"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/parsers"
	"github.com/RefuseHQ/refuse-cli/internal/server"
	"github.com/spf13/cobra"
)

func realCheckCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "check <ecosystem> <package>[@version]",
		Short: "Check a single package against the refuse server",
		Long: `Examples:
  refuse check npm lodash@4.17.10
  refuse check PyPI requests@2.19.0
  refuse check --json npm lodash@4.17.10`,
		Args: cobra.RangeArgs(2, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			ecosystem := args[0]
			name, version := parsers.SplitNameVersionExt(args[1])
			if version == "" {
				version = "latest"
			}

			ctx, cancel := context.WithTimeout(context.Background(), defaultCheckTimeout)
			defer cancel()
			c := server.New(cfg.ServerURL, cfg.APIKey)
			res, err := c.CheckPackage(ctx, server.CheckPackageRequest{Ecosystem: ecosystem, Name: name, Version: version})
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(res)
			}
			renderPackageResult(cmd.OutOrStdout(), res)
			if res.Vulnerable {
				os.Exit(2)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print the raw JSON response instead of a summary")
	return cmd
}

func realCheckLockfileCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "check-lockfile <path>",
		Short: "Parse and check a lockfile",
		Long: `Supports the lockfile formats the server knows: package-lock.json,
yarn.lock, pnpm-lock.yaml, requirements.txt, Cargo.lock, Gemfile.lock,
go.sum, composer.lock, pom.xml, *.csproj, mix.lock, pubspec.lock.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return err
			}
			content, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(context.Background(), defaultLockfileTimeout)
			defer cancel()
			c := server.New(cfg.ServerURL, cfg.APIKey)
			res, err := c.CheckLockfile(ctx, server.CheckLockfileRequest{
				Filename: filepathBase(args[0]),
				Content:  string(content),
			})
			if err != nil {
				return err
			}
			if jsonOut {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(res)
			}
			renderBatchResult(cmd.OutOrStdout(), res)
			if res.Summary.Vulnerable > 0 {
				os.Exit(2)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Print the raw JSON response instead of a summary")
	return cmd
}
