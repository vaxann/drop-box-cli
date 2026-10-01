// Package core holds the logic shared by the interactive TUI and the
// non-interactive subcommands: ordering servers by usage, building remote
// paths and running a transfer without a terminal.
package core

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/vaxann/drop-box-cli/internal/sshconf"
	"github.com/vaxann/drop-box-cli/internal/store"
	"github.com/vaxann/drop-box-cli/internal/transfer"
)

// SortHosts returns hosts most frequently used first (ties broken by
// recency, then config order), so the usual destination comes first.
func SortHosts(hosts []sshconf.Host, st *store.Store) []sshconf.Host {
	sorted := append([]sshconf.Host(nil), hosts...)
	sort.SliceStable(sorted, func(i, j int) bool {
		ui, uj := st.HostUse(sorted[i].Alias), st.HostUse(sorted[j].Alias)
		if ui.Count != uj.Count {
			return ui.Count > uj.Count
		}
		return ui.LastUsed.After(uj.LastUsed)
	})
	return sorted
}

// Describe renders a host's connection target as user@hostname.
func Describe(h sshconf.Host) string {
	if h.User != "" && h.HostName != "" {
		return h.User + "@" + h.HostName
	}
	return h.HostName
}

// RemotePaths returns where files land once copied into absDir.
func RemotePaths(absDir string, files []string) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = absDir + "/" + path.Base(strings.TrimRight(f, "/"))
	}
	return out
}

// Send copies files into dir on host without a terminal: ssh runs in batch
// mode, so a missing key fails fast instead of waiting for a password.
// It returns the absolute remote directory and the remote file paths.
func Send(host, dir string, files []string) (string, []string, error) {
	var stdout, stderr bytes.Buffer
	cmd := transfer.ResolveCmd(host, dir, transfer.BatchOptions...)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", nil, cmdError(fmt.Sprintf("resolving %s on %s", dir, host), err, &stderr)
	}
	absDir := strings.TrimSpace(stdout.String())
	if absDir == "" {
		return "", nil, fmt.Errorf("resolving %s on %s: empty path", dir, host)
	}

	stderr.Reset()
	cmd = transfer.CopyCmd(host, files, absDir, transfer.BatchOptions...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", nil, cmdError(fmt.Sprintf("copying to %s:%s", host, absDir), err, &stderr)
	}
	return absDir, RemotePaths(absDir, files), nil
}

// cmdError prefers the command's own stderr (e.g. "Permission denied
// (publickey).") over the bare exit status.
func cmdError(what string, err error, stderr *bytes.Buffer) error {
	if msg := strings.TrimSpace(stderr.String()); msg != "" {
		return fmt.Errorf("%s: %w", what, errors.New(msg))
	}
	return fmt.Errorf("%s: %w", what, err)
}
