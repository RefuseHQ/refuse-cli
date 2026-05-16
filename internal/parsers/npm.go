package parsers

// NPM handles `npm`, `pnpm`, and `bun` — they all accept the same install/add
// verb shape. `yarn` has its own file because the verb mapping diverges.

var npmInstallVerbs = map[string]bool{
	"install": true,
	"i":       true,
	"add":     true,
	"isntall": true, // npm accepts this canonical typo
}

// npm flags that *consume the next token*.
var npmFlagsTakingArg = map[string]bool{
	"--workspace":  true,
	"-w":           true,
	"--save-exact": false, // boolean
	"--save-dev":   false,
	"-D":           false,
	"-E":           false,
	"-g":           false,
	"--global":     false,
	"--no-save":    false,
}

// NPM returns a parser for npm/pnpm/bun argv.
func NPM() Parser { return npmLikeParser{ecosystem: "npm", lockfile: "package-lock.json"} }

// PNPM returns the same parser bound to pnpm-lock.yaml for `pnpm install`
// (no args) → lockfile mode.
func PNPM() Parser { return npmLikeParser{ecosystem: "npm", lockfile: "pnpm-lock.yaml"} }

// Bun is the bun equivalent. Uses bun.lock(b).
func Bun() Parser { return npmLikeParser{ecosystem: "npm", lockfile: "bun.lock"} }

type npmLikeParser struct {
	ecosystem string
	lockfile  string
}

func (p npmLikeParser) Parse(args []string) ParseResult {
	if len(args) == 0 {
		return ParseResult{}
	}
	verb := args[0]

	// `ci` is always lockfile mode for npm/pnpm; bun doesn't ship a `ci`
	// verb but accepting it as lockfile-mode is a safe default.
	if verb == "ci" {
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: p.lockfile,
			Reason:       p.ecosystem + " ci → lockfile mode",
		}
	}

	if !npmInstallVerbs[verb] {
		return ParseResult{}
	}

	positionals := splitPositionals(args[1:], npmFlagsTakingArg)

	if len(positionals) == 0 {
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: p.lockfile,
			Reason:       p.ecosystem + " " + verb + " (no args) → lockfile mode",
		}
	}

	pkgs := make([]PkgRef, 0, len(positionals))
	for _, spec := range positionals {
		name, version := splitNameVersion(spec)
		if name == "" {
			continue
		}
		pkgs = append(pkgs, PkgRef{Ecosystem: p.ecosystem, Name: name, Version: version})
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: p.ecosystem + " direct"}
}
