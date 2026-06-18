package shim

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Markers wrap the block we manage in shell-rc files. Anything between them
// (inclusive) is owned by refuse — uninstall removes the block, install
// rewrites it in place.
const (
	beginMarker = "# >>> refuse shim (managed) >>>"
	endMarker   = "# <<< refuse shim (managed) <<<"
)

// markerPair is a begin/end delimiter set we recognise in shell-rc files.
type markerPair struct{ begin, end string }

// managedMarkers is every block refuse owns. The first is canonical (what we
// write today); the rest are legacy delimiters from older installers that we
// still strip so install consolidates duplicates and uninstall cleans up fully.
// The curl|sh bootstrap (scripts/install.sh) historically wrote a
// "refuse cli (managed)" block with a different marker than the binary, so an
// uninstall used to leave it behind — this list fixes that.
var managedMarkers = []markerPair{
	{beginMarker, endMarker},
	{"# >>> refuse cli (managed) >>>", "# <<< refuse cli (managed) <<<"},
}

// updateShellRC writes (`enable`=true) or removes (`enable`=false) the PATH
// block in the user's shell-rc files. Returns the rc files actually touched.
func updateShellRC(binDir string, enable bool) ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// Sample of common rc files. We only touch a file if it already exists
	// (don't create a fresh .zshrc on a user's system just to add PATH).
	candidates := []string{
		filepath.Join(home, ".bashrc"),
		filepath.Join(home, ".bash_profile"),
		filepath.Join(home, ".zshrc"),
		filepath.Join(home, ".profile"),
	}

	var touched []string
	for _, p := range candidates {
		info, err := os.Stat(p)
		if err != nil || info.IsDir() {
			continue
		}
		changed, err := patchFile(p, binDir, enable)
		if err != nil {
			return touched, fmt.Errorf("patch %s: %w", p, err)
		}
		if changed {
			touched = append(touched, p)
		}
	}

	// Fish has its own conf.d location — additive, optional.
	fishConf := filepath.Join(home, ".config", "fish", "conf.d", "refuse.fish")
	if enable {
		if _, err := os.Stat(filepath.Dir(fishConf)); err == nil {
			if err := writeFishConf(fishConf, binDir); err != nil {
				return touched, err
			}
			touched = append(touched, fishConf)
		}
	} else {
		if _, err := os.Stat(fishConf); err == nil {
			_ = os.Remove(fishConf)
			touched = append(touched, fishConf)
		}
	}

	return touched, nil
}

func managedBlock(binDir string) string {
	var b strings.Builder
	b.WriteString(beginMarker + "\n")
	fmt.Fprintf(&b, `export PATH="%s:$PATH"`, binDir)
	b.WriteString("\n" + endMarker + "\n")
	return b.String()
}

// stripBlock removes every begin..end block (inclusive, plus a trailing
// newline) from content. Loops so duplicate blocks are all removed.
func stripBlock(content, begin, end string) string {
	for {
		start := strings.Index(content, begin)
		if start < 0 {
			return content
		}
		rel := strings.Index(content[start:], end)
		if rel < 0 {
			return content // unterminated marker — leave it rather than eat the rest
		}
		stop := start + rel + len(end)
		if stop < len(content) && content[stop] == '\n' {
			stop++
		}
		content = content[:start] + content[stop:]
	}
}

func patchFile(path, binDir string, enable bool) (bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}
	content := string(raw)

	// Strip every managed block we recognise — canonical and legacy, and any
	// duplicates of each — so install consolidates to one block and uninstall
	// removes them all (including a stray "refuse cli (managed)" block from an
	// older curl|sh install).
	withoutBlock := content
	for _, m := range managedMarkers {
		withoutBlock = stripBlock(withoutBlock, m.begin, m.end)
	}

	var next string
	if enable {
		// Append the managed block to the (now-clean) file. Ensure a trailing
		// newline before adding.
		next = strings.TrimRight(withoutBlock, "\n") + "\n\n" + managedBlock(binDir)
	} else {
		next = withoutBlock
	}

	if next == content {
		return false, nil
	}
	return true, os.WriteFile(path, []byte(strings.TrimRight(next, "\n")+"\n"), 0o644)
}

func writeFishConf(path, binDir string) error {
	body := fmt.Sprintf("# Managed by refuse — do not edit.\nset -gx PATH %s $PATH\n", binDir)
	// Ensure the parent dir exists.
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

// for tests: expose helpers via this stub so the file is referenced even
// when the bytes package isn't otherwise needed.
var _ = bytes.NewReader
