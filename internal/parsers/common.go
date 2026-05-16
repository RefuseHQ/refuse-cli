package parsers

import "strings"

// flagsThatTakeAnArgument is the global set of "this flag consumes the next
// token" patterns used by tokenize(). Each manager extends it via its own
// table; we keep the universal ones here.
var universalFlagsTakingArg = map[string]bool{
	"--prefix":     true,
	"--registry":   true,
	"--cwd":        true,
	"--config":     true,
	"-C":           true,
	"--directory":  true,
}

// splitPositionals walks argv left-to-right and returns the positional
// tokens (skipping flags and their arguments). `extraTakeArg` is a per-
// manager map of flags that consume the next token.
func splitPositionals(args []string, extraTakeArg map[string]bool) []string {
	out := make([]string, 0, len(args))
	skipNext := false
	for _, a := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if !strings.HasPrefix(a, "-") {
			out = append(out, a)
			continue
		}
		// Flag. Decide whether to also skip the next token.
		base := a
		if eq := strings.IndexByte(a, '='); eq != -1 {
			base = a[:eq] // --prefix=/opt → "--prefix", no next-token skip
			continue
		}
		if universalFlagsTakingArg[base] || (extraTakeArg != nil && extraTakeArg[base]) {
			skipNext = true
		}
	}
	return out
}

// splitNameVersion takes an npm-style specifier and returns (name, version).
// Handles scoped packages: `@scope/name@^1.0.0` → ("@scope/name", "^1.0.0").
// Returns ("", "") for empty input.
func splitNameVersion(spec string) (name, version string) {
	if spec == "" {
		return "", ""
	}
	// Trim a leading '@' from scoped names before searching for the version
	// separator, then put it back.
	scoped := strings.HasPrefix(spec, "@")
	rest := spec
	if scoped {
		rest = spec[1:]
	}
	at := strings.IndexByte(rest, '@')
	if at == -1 {
		return spec, ""
	}
	prefix := rest[:at]
	suffix := rest[at+1:]
	if scoped {
		prefix = "@" + prefix
	}
	return prefix, suffix
}

// splitPipSpec parses a single pip install specifier. Handles `==`, `~=`,
// `>=`, `<=`, `<`, `>`, `!=`. Returns ("", "") for empty input.
//
// Note we only extract the *first* version constraint and treat it as the
// requested version; pip's spec syntax is richer (`requests>=2,<3`) but
// we don't try to satisfy a range from the CLI — server side handles that
// for lockfile mode.
func splitPipSpec(spec string) (name, version string) {
	if spec == "" {
		return "", ""
	}
	// Find the first non-name character.
	for i, r := range spec {
		switch r {
		case '=', '<', '>', '~', '!':
			version = strings.TrimLeft(spec[i:], "=<>~!")
			return strings.TrimSpace(spec[:i]), strings.TrimSpace(version)
		}
	}
	return spec, ""
}
