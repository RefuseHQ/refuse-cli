//go:build windows

package shim

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
)

// execReal runs the real package manager as a child process and waits for it
// to exit. Unlike POSIX where we syscall.Exec into the child, Windows has no
// equivalent — refuse.exe stays alive as the parent, the child inherits our
// console, and both processes share the same SetConsoleCtrlHandler chain.
//
// libuv (the I/O loop inside every modern Node.js install, including npm's
// own bootstrap) keeps a per-loop signal slot it asserts on at signal.c:277
// — handle->pending_signum == 0. When refuse.exe also registers a Go signal
// handler, Ctrl-C delivery and the parent's normal-exit signal teardown
// race the child's libuv dispatch, violating that invariant. The child
// abort()s with the assertion, the console is left in a broken state, and
// the user's PowerShell session wedges.
//
// The fix is to make refuse.exe an inert signal-passer: install a
// signal.Notify that swallows os.Interrupt in the parent (without acting on
// it) so the OS still delivers CTRL_C_EVENT to the child via the shared
// console, and the child owns its own signal lifecycle uncontested. We use
// Start()/Wait() instead of Run() so we can install the handler between
// process start and process exit.
func execReal(name string, args []string) int {
	bin, err := findReal(name)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 127
	}

	cmd := exec.Command(bin, args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	// Swallow Ctrl-C in the parent so only the child reacts. Buffer of 1
	// is enough — we never read from the channel, but the runtime needs
	// somewhere to deliver pending signals.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	defer signal.Stop(sigCh)

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "refuse: exec %s failed: %v\n", strings.Join(append([]string{bin}, args...), " "), err)
		return 1
	}

	if err := cmd.Wait(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "refuse:", err.Error())
		return 1
	}
	return 0
}
