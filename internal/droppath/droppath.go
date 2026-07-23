// Package droppath parses the text a terminal emulator inserts when files
// are dragged and dropped onto it.
//
// Depending on the terminal, a dropped file may arrive as a plain path,
// a single- or double-quoted path, a backslash-escaped path, or a file://
// URI. Multiple files arrive space-separated.
package droppath

import (
	"net/url"
	"strings"
	"unicode"
)

// Parse splits a dropped line into individual file paths.
func Parse(line string) []string {
	var (
		paths          []string
		cur            strings.Builder
		inSingle       bool
		inDouble       bool
		escaped        bool
		hasContent     bool
	)
	flush := func() {
		if hasContent {
			paths = append(paths, normalize(cur.String()))
		}
		cur.Reset()
		hasContent = false
	}
	for _, r := range line {
		switch {
		case escaped:
			cur.WriteRune(r)
			hasContent = true
			escaped = false
		case r == '\\' && !inSingle:
			escaped = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			hasContent = true
		case r == '"' && !inSingle:
			inDouble = !inDouble
			hasContent = true
		case unicode.IsSpace(r) && !inSingle && !inDouble:
			flush()
		default:
			cur.WriteRune(r)
			hasContent = true
		}
	}
	flush()
	return paths
}

func normalize(p string) string {
	if !strings.HasPrefix(p, "file://") {
		return p
	}
	u := strings.TrimPrefix(p, "file://")
	// Strip an optional host component (file://localhost/tmp/x).
	if i := strings.IndexByte(u, '/'); i > 0 {
		u = u[i:]
	}
	if dec, err := url.PathUnescape(u); err == nil {
		return dec
	}
	return u
}
