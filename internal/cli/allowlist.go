package cli

import (
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/RefuseHQ/refuse-cli/internal/allowlist"
)

func realAllowlistCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allowlist",
		Short: "Manage the project's accepted-risk allowlist (`.refuse.yaml`)",
		Long: `Per-project allowlist for CVEs your team has reviewed and accepted.
Stored in ` + "`.refuse.yaml`" + ` at the project root. Entries hide the matched
advisory from the gate's block decision so day-to-day installs don't
churn on the same alert.`,
	}
	cmd.AddCommand(allowlistAddCmd())
	cmd.AddCommand(allowlistRemoveCmd())
	cmd.AddCommand(allowlistListCmd())
	return cmd
}

func allowlistAddCmd() *cobra.Command {
	var pkg, ecosystem, reason, expires, path string
	cmd := &cobra.Command{
		Use:   "add <CVE-ID>",
		Short: "Add or update an entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := resolveAllowlistPath(path)
			f, err := allowlist.Load(fp)
			if err != nil {
				return err
			}
			entry := allowlist.Entry{
				CVE:       args[0],
				Package:   pkg,
				Ecosystem: ecosystem,
				Reason:    reason,
				Expires:   expires,
			}
			f.Add(entry)
			if err := allowlist.Save(fp, f); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "refuse: %s added to %s\n", args[0], fp)
			return nil
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Only suppress this CVE for the named package (optional)")
	cmd.Flags().StringVar(&ecosystem, "ecosystem", "", "Only suppress for this ecosystem — npm, PyPI, crates.io, etc. (optional)")
	cmd.Flags().StringVar(&reason, "reason", "", "Why the risk is accepted (recommended)")
	cmd.Flags().StringVar(&expires, "expires", "", "ISO date YYYY-MM-DD after which the entry stops suppressing (optional)")
	cmd.Flags().StringVar(&path, "file", "", "Path to .refuse.yaml (default: nearest ancestor, or ./.refuse.yaml if none)")
	return cmd
}

func allowlistRemoveCmd() *cobra.Command {
	var pkg, ecosystem, path string
	cmd := &cobra.Command{
		Use:   "remove <CVE-ID>",
		Short: "Remove an entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := resolveAllowlistPath(path)
			f, err := allowlist.Load(fp)
			if err != nil {
				return err
			}
			n := f.Remove(args[0], ecosystem, pkg)
			if n == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "refuse: no matching entry in %s\n", fp)
				return nil
			}
			if err := allowlist.Save(fp, f); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "refuse: removed %d entry(s) from %s\n", n, fp)
			return nil
		},
	}
	cmd.Flags().StringVar(&pkg, "package", "", "Match this package (must match exactly what was added)")
	cmd.Flags().StringVar(&ecosystem, "ecosystem", "", "Match this ecosystem")
	cmd.Flags().StringVar(&path, "file", "", "Path to .refuse.yaml")
	return cmd
}

func allowlistListCmd() *cobra.Command {
	var path string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Print every entry, marking expired ones",
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := resolveAllowlistPath(path)
			f, err := allowlist.Load(fp)
			if err != nil {
				return err
			}
			if len(f.Allowlist) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "refuse: %s is empty\n", fp)
				return nil
			}
			entries := append([]allowlist.Entry(nil), f.Allowlist...)
			sort.Slice(entries, func(i, j int) bool { return entries[i].CVE < entries[j].CVE })
			now := time.Now()
			fmt.Fprintf(cmd.OutOrStdout(), "%s — %d entry(s)\n", fp, len(entries))
			for _, e := range entries {
				expiredNote := ""
				if e.Expires != "" {
					t, err := time.Parse("2006-01-02", e.Expires)
					if err == nil && t.Before(now) {
						expiredNote = "  (EXPIRED — no longer suppressing)"
					}
				}
				scope := "all packages"
				if e.Package != "" {
					scope = e.Package
					if e.Ecosystem != "" {
						scope = e.Ecosystem + ":" + e.Package
					}
				}
				fmt.Fprintf(cmd.OutOrStdout(), "  %-16s  %s%s\n", e.CVE, scope, expiredNote)
				if e.Reason != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "    reason:  %s\n", e.Reason)
				}
				if e.Expires != "" {
					fmt.Fprintf(cmd.OutOrStdout(), "    expires: %s\n", e.Expires)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&path, "file", "", "Path to .refuse.yaml")
	return cmd
}

// resolveAllowlistPath honours an explicit --file flag; otherwise walks
// up from cwd for the nearest .refuse.yaml; if none, defaults to
// ./.refuse.yaml so `add` can create it.
func resolveAllowlistPath(explicit string) string {
	if explicit != "" {
		return explicit
	}
	cwd, _ := os.Getwd()
	if found := allowlist.Find(cwd); found != "" {
		return found
	}
	return allowlist.FileName
}
