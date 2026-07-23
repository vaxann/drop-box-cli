// Package transfer builds the ssh/scp commands used to resolve the remote
// target directory and copy files. The system ssh/scp binaries are used so
// keys, ssh-agent, ProxyJump and the rest of ~/.ssh/config work as usual.
package transfer

import (
	"fmt"
	"os/exec"
	"strings"
)

// ResolveCmd returns a command that creates dir on host (if needed) and
// prints its absolute path to stdout.
func ResolveCmd(host, dir string) *exec.Cmd {
	arg := RemoteArg(dir)
	remote := fmt.Sprintf("mkdir -p -- %s && cd -- %s && pwd", arg, arg)
	return exec.Command("ssh", "--", host, remote)
}

// ScpCmd returns a command copying files into absDir on host. absDir must
// already be absolute (see ResolveCmd).
func ScpCmd(host string, files []string, absDir string) *exec.Cmd {
	args := []string{"-r", "--"}
	for _, f := range files {
		args = append(args, safeLocal(f))
	}
	args = append(args, host+":"+EscapeRemote(absDir)+"/")
	return exec.Command("scp", args...)
}

// RemoteArg quotes a user-supplied directory for use inside a remote shell
// command, keeping a leading ~ expandable.
func RemoteArg(dir string) string {
	switch {
	case dir == "~":
		return `"$HOME"`
	case strings.HasPrefix(dir, "~/"):
		return `"$HOME"/` + singleQuote(dir[2:])
	default:
		return singleQuote(dir)
	}
}

// EscapeRemote backslash-escapes a remote path for use in an scp remote
// operand, which is interpreted by the remote shell (classic mode) or by
// scp's own glob parser (sftp mode).
func EscapeRemote(p string) string {
	var b strings.Builder
	for _, r := range p {
		if strings.ContainsRune(" \t'\"\\$&;()<>|*?[]{}#!`~^", r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

func singleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// safeLocal guards against a local path being mistaken for a remote spec:
// scp treats anything before the first ':' as a host name.
func safeLocal(p string) string {
	if !strings.Contains(p, "/") && strings.Contains(p, ":") {
		return "./" + p
	}
	return p
}
