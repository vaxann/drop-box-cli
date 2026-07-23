// Package transfer builds the ssh/scp commands used to resolve the remote
// target directory and copy files. The system ssh/scp binaries are used so
// keys, ssh-agent, ProxyJump and the rest of ~/.ssh/config work as usual.
package transfer

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

// ResolveCmd returns a command that creates dir on host (if needed) and
// prints its absolute path to stdout.
func ResolveCmd(host, dir string) *exec.Cmd {
	arg := RemoteArg(dir)
	remote := fmt.Sprintf("mkdir -p -- %s && cd -- %s && pwd", arg, arg)
	return exec.Command("ssh", "--", host, remote)
}

// CopyCmd returns a command that streams files (or directories) into
// absDir on host via tar over ssh. absDir must already be absolute (see
// ResolveCmd).
//
// scp is deliberately avoided: how it treats the remote path depends on
// the local OpenSSH version — the classic protocol runs it through the
// remote shell while the sftp default (OpenSSH >= 9.0) takes it near
// literally — so no single escaping is correct for both. A tar pipe keeps
// all quoting on our side of a plain ssh command.
func CopyCmd(host string, files []string, absDir string) *exec.Cmd {
	var b strings.Builder
	b.WriteString("tar -cf -")
	for _, f := range files {
		dir, name := filepath.Split(filepath.Clean(f))
		if dir == "" {
			dir = "."
		}
		fmt.Fprintf(&b, " -C %s %s", singleQuote(dir), singleQuote("./"+name))
	}
	remote := fmt.Sprintf("cd -- %s && tar -xf -", singleQuote(absDir))
	fmt.Fprintf(&b, " | ssh -- %s %s", singleQuote(host), singleQuote(remote))
	return exec.Command("sh", "-c", b.String())
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

func singleQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
