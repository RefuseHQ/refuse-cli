package parsers

import "strings"

// PDM handles `pdm add` and `pdm install`.
//   pdm add <pkg>              → direct
//   pdm add <pkg>@1.0          → direct, pinned
//   pdm add <pkg>==1.0         → direct, pinned (pip form)
//   pdm install / sync         → lockfile (pdm.lock)
//   pdm update / run / build   → passthrough

var pdmFlagsTakingArg = map[string]bool{
	"--group":      true,
	"-G":           true,
	"--python":     true,
	"--editable":   false,
	"-e":           false,
	"--prerelease": false,
}

// PDM returns a parser for `pdm` argv.
func PDM() Parser { return pdmParser{} }

type pdmParser struct{}

func (pdmParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	switch args[0] {
	case "add":
		positionals := splitPositionals(args[1:], pdmFlagsTakingArg)
		pkgs := make([]PkgRef, 0, len(positionals))
		for _, spec := range positionals {
			if strings.Contains(spec, "/") || strings.HasPrefix(spec, ".") ||
				strings.HasPrefix(spec, "git+") {
				continue
			}
			name, version := splitNameVersion(spec)
			if version == "" {
				name, version = splitPipSpec(spec)
			}
			if name == "" {
				continue
			}
			pkgs = append(pkgs, PkgRef{Ecosystem: "PyPI", Name: name, Version: version})
		}
		if len(pkgs) == 0 {
			return ParseResult{}
		}
		return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "pdm add"}
	case "install", "sync":
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "pdm.lock",
			Reason:       "pdm " + args[0] + " → lockfile",
		}
	}
	return ParseResult{}
}
