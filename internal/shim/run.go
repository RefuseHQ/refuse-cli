package shim

import (
	"context"
	"fmt"
	"os"

	"github.com/RefuseHQ/refuse-cli/internal/config"
	"github.com/RefuseHQ/refuse-cli/internal/gate"
	"github.com/RefuseHQ/refuse-cli/internal/parsers"
	"github.com/RefuseHQ/refuse-cli/internal/server"
)

// runFromArgv is the real implementation behind Run(): parse argv, vet via
// the gate, exec the real package manager. Set REFUSE_NO_GATE=1 to skip the
// gate entirely (useful for triaging "is refuse breaking my install?").
func runFromArgv(name string, args []string) int {
	if os.Getenv("REFUSE_NO_GATE") == "1" {
		return execReal(name, args)
	}

	parser := parsers.ForName(name)
	if parser == nil {
		// Unknown manager — shouldn't reach here, but exec anyway.
		return execReal(name, args)
	}

	result := parser.Parse(args)
	if !result.IsInstall {
		return execReal(name, args)
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintln(os.Stderr, "refuse: config error:", err.Error())
		return execReal(name, args)
	}

	eng := gate.Engine{
		Client: server.New(cfg.ServerURL, cfg.APIKey),
		Policy: cfg.Policy,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	res := eng.Decide(ctx, result)
	if res.Decision == gate.DecisionBlock {
		if res.Message != "" {
			fmt.Fprint(os.Stderr, res.Message)
		}
		// 2 is the exit code agents recognise as "tool refused" in Claude
		// Code's hook protocol. For shim mode any non-zero would do, but
		// staying consistent makes downstream tooling simpler.
		return 2
	}
	if res.Message != "" {
		fmt.Fprintln(os.Stderr, res.Message)
	}
	return execReal(name, args)
}
