package parsers

// Pipenv handles `pipenv install`.
//   pipenv install <pkg>          → direct
//   pipenv install <pkg>==1.0     → direct, pinned
//   pipenv install                → lockfile (Pipfile.lock)
//   pipenv update / lock / run    → passthrough

var pipenvFlagsTakingArg = map[string]bool{
	"--python":      true,
	"--three":       false,
	"--two":         false,
	"--dev":         false,
	"-d":            false,
	"--categories":  true,
	"--index":       true,
	"-i":            true,
	"--extra-index": true,
}

// Pipenv returns a parser for `pipenv` argv.
func Pipenv() Parser { return pipenvParser{} }

type pipenvParser struct{}

func (pipenvParser) Parse(args []string) ParseResult {
	if len(args) == 0 || args[0] != "install" {
		return ParseResult{}
	}
	positionals := splitPositionals(args[1:], pipenvFlagsTakingArg)
	if len(positionals) == 0 {
		// Bare `pipenv install` → resolves Pipfile.lock.
		return ParseResult{
			IsInstall:    true,
			Mode:         ModeLockfile,
			LockfileHint: "Pipfile.lock",
			Reason:       "pipenv install → lockfile",
		}
	}
	pkgs := pyPositionalsToPkgs(positionals)
	if len(pkgs) == 0 {
		return ParseResult{}
	}
	return ParseResult{IsInstall: true, Mode: ModeDirect, Packages: pkgs, Reason: "pipenv install"}
}
