package parsers

// ForName returns the parser bound to a given package-manager binary name.
// Returns nil for unknown names (the shim layer treats nil → passthrough).
func ForName(name string) Parser {
	switch name {
	case "npm":
		return NPM()
	case "pnpm":
		return PNPM()
	case "yarn":
		return Yarn()
	case "bun":
		return Bun()
	case "pip", "pip3":
		return Pip()
	case "cargo":
		return Cargo()
	case "gem":
		return Gem()
	case "go":
		return Go()
	}
	return nil
}
