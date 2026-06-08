//go:build !windows

package shim

import (
	"fmt"
	"os"
	"strings"
	"syscall"
)

// execReal replaces the current process with the real package manager via
// syscall.Exec — the parent shell sees the real exit code directly, and the
// shim drops out of the process tree entirely.
func execReal(name string, args []string) int {
	bin, err := findReal(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 127
	}

	argv := append([]string{bin}, args...)
	env := os.Environ()
	if err := syscall.Exec(bin, argv, env); err != nil {
		fmt.Fprintf(os.Stderr, "refuse: exec %s failed: %v\n", strings.Join(argv, " "), err)
		return 1
	}
	return 0 // unreachable
}
