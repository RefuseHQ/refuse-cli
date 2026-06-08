// Package main is the entry point for the `refuse` binary.
//
// `refuse` is a multicall binary: it inspects argv[0] on launch and either
// behaves as the regular CLI (`refuse check`, `refuse install`, etc.) or as
// a transparent shim for a package manager (when invoked through a symlink
// in ~/.refuse/bin/{npm,pnpm,yarn,pip,...}).
package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/RefuseHQ/refuse-cli/internal/cli"
	"github.com/RefuseHQ/refuse-cli/internal/shim"
)

func main() {
	name := strings.ToLower(filepath.Base(os.Args[0]))
	name = strings.TrimSuffix(name, ".exe")

	// Shim mode — argv[0] looks like a package manager. Vet the call, then
	// exec the real binary.
	if shim.IsKnownManager(name) {
		os.Exit(shim.Run(name, os.Args[1:]))
	}

	// Normal CLI mode.
	if err := cli.NewRoot().Execute(); err != nil {
		// Cobra prints the error itself; just set the exit code.
		os.Exit(1)
	}
}
