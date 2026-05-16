package shim

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// SelfDir returns the directory containing the running refuse binary, which
// is also where the shims live. We need this to avoid recursing back into
// ourselves when resolving the real package manager.
func SelfDir() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		// Best-effort — fall back to the raw path.
		resolved = exe
	}
	return filepath.Dir(resolved), nil
}

// findReal walks PATH (minus the shim dir) and returns the first executable
// named `name`. Returns an error if nothing matches.
func findReal(name string) (string, error) {
	self, err := SelfDir()
	if err != nil {
		return "", err
	}
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == "" || dir == self {
			continue
		}
		cand := filepath.Join(dir, name)
		if info, err := os.Stat(cand); err == nil && !info.IsDir() && info.Mode()&0o111 != 0 {
			return cand, nil
		}
	}
	return "", fmt.Errorf("refuse: real %q not found on PATH (self dir %s excluded)", name, self)
}

// execReal replaces the current process with the real package manager. On
// Unix this is a syscall.Exec — the parent shell sees the real exit code
// directly. On Windows we fall back to os/exec since true exec doesn't
// exist there.
func execReal(name string, args []string) int {
	bin, err := findReal(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 127
	}

	if isWindows() {
		cmd := exec.Command(bin, args...)
		cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				return exitErr.ExitCode()
			}
			fmt.Fprintln(os.Stderr, "refuse:", err.Error())
			return 1
		}
		return 0
	}

	argv := append([]string{bin}, args...)
	env := os.Environ()
	if err := syscall.Exec(bin, argv, env); err != nil {
		fmt.Fprintf(os.Stderr, "refuse: exec %s failed: %v\n", strings.Join(argv, " "), err)
		return 1
	}
	return 0 // unreachable
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}
