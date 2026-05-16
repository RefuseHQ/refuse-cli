package parsers

// Yarn handles both Yarn classic v1 (`yarn add`, `yarn install`) and Berry.
// The verb semantics:
//   yarn add foo bar           → direct
//   yarn install / yarn        → lockfile
//   yarn run / yarn test ...   → passthrough

var yarnFlagsTakingArg = map[string]bool{
	"--cwd": true,
	"-D":    false,
	"--dev": false,
}

// Yarn returns a parser for `yarn` argv.
func Yarn() Parser { return yarnParser{} }

type yarnParser struct{}

func (yarnParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		// Bare `yarn` is shorthand for `yarn install`.
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "yarn.lock",
			Reason:       "yarn (no args) → install lockfile",
		}
	}
	verb := args[0]
	switch verb {
	case "add":
		positionals := splitPositionals(args[1:], yarnFlagsTakingArg)
		pkgs := make([]PkgRef, 0, len(positionals))
		for _, spec := range positionals {
			name, version := splitNameVersion(spec)
			if name == "" {
				continue
			}
			pkgs = append(pkgs, PkgRef{Ecosystem: "npm", Name: name, Version: version})
		}
		return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "yarn add"}
	case "install":
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "yarn.lock",
			Reason:       "yarn install → lockfile",
		}
	}
	return ParseResult{}
}
