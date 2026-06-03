package parsers

import "strings"

// NPX handles `npx`. npx fetches a package from the npm registry and runs
// its bin, installing transiently if not cached. From a gate perspective
// this is effectively a one-shot install — we treat the package being run
// as a "direct install" of the npm ecosystem.
//
//   npx <pkg>              → direct
//   npx <pkg>@1.0          → direct, pinned
//   npx -p <pkg> <cmd>     → direct (the -p package is what gets fetched)
//   npx <local-script>     → passthrough (path-like)
//
// Anything that looks like a local path is left alone; npx can run any
// command, and we don't want to gate `npx ./scripts/foo`.

var npxFlagsTakingArg = map[string]bool{
	"-p":           true,
	"--package":    true,
	"-c":           true,
	"--call":       true,
	"--userconfig": true,
	"--shell":      true,
}

// NPX returns a parser for `npx` argv.
func NPX() Parser { return npxParser{} }

type npxParser struct{}

func (npxParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	// Collect the first positional after flag-eating. npx runs ONE command;
	// later positionals are args to that command, not extra installs.
	var primary string
	var explicit []string // anything passed via -p / --package
	skipNext := false
	for i, a := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if a == "-p" || a == "--package" {
			if i+1 < len(args) {
				explicit = append(explicit, args[i+1])
				skipNext = true
			}
			continue
		}
		if strings.HasPrefix(a, "--package=") {
			explicit = append(explicit, strings.TrimPrefix(a, "--package="))
			continue
		}
		if npxFlagsTakingArg[a] {
			skipNext = true
			continue
		}
		if strings.HasPrefix(a, "-") {
			continue
		}
		// First non-flag positional is the package/command. Stop here.
		primary = a
		break
	}

	candidates := explicit
	if len(candidates) == 0 && primary != "" {
		candidates = []string{primary}
	}

	pkgs := make([]PkgRef, 0, len(candidates))
	for _, spec := range candidates {
		// Skip path-like / URL specs (`./foo`, `/abs/path`, `https://…`).
		if strings.HasPrefix(spec, ".") || strings.HasPrefix(spec, "/") ||
			strings.Contains(spec, "://") {
			continue
		}
		name, version := splitNameVersion(spec)
		if name == "" {
			continue
		}
		pkgs = append(pkgs, PkgRef{Ecosystem: "npm", Name: name, Version: version})
	}
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "npx"}
}
