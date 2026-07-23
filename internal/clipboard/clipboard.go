// Package clipboard copies text to the system clipboard via whichever
// well-known helper binary is available.
package clipboard

import (
	"errors"
	"os/exec"
	"strings"
)

var candidates = [][]string{
	{"wl-copy"},
	{"xclip", "-selection", "clipboard"},
	{"xsel", "--clipboard", "--input"},
	{"pbcopy"},
}

// Copy puts text on the clipboard and reports the helper used.
func Copy(text string) (string, error) {
	for _, c := range candidates {
		path, err := exec.LookPath(c[0])
		if err != nil {
			continue
		}
		cmd := exec.Command(path, c[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if err := cmd.Run(); err == nil {
			return c[0], nil
		}
	}
	return "", errors.New("no working clipboard helper (wl-copy, xclip, xsel, pbcopy)")
}
