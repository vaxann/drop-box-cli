package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/vaxann/drop-box-cli/internal/clipboard"
	"github.com/vaxann/drop-box-cli/internal/core"
	"github.com/vaxann/drop-box-cli/internal/sshconf"
	"github.com/vaxann/drop-box-cli/internal/store"
)

// Non-interactive subcommands for launchers and scripts (e.g. the Raycast
// extension). They never prompt: ssh runs in batch mode, results go to
// stdout (as JSON with --json), errors to stderr with a non-zero exit.
var subcommands = map[string]func(args []string) error{
	"hosts": cmdHosts,
	"dirs":  cmdDirs,
	"send":  cmdSend,
	"last":  cmdLast,
}

type hostJSON struct {
	Alias    string `json:"alias"`
	HostName string `json:"hostname"`
	User     string `json:"user"`
	Count    int    `json:"count"`
}

type sendJSON struct {
	Host        string   `json:"host"`
	Dir         string   `json:"dir"`
	RemotePaths []string `json:"remote_paths"`
	Clipboard   string   `json:"clipboard,omitempty"`
}

type lastJSON struct {
	Host string `json:"host"`
	Dir  string `json:"dir"`
}

func cmdHosts(args []string) error {
	fs := flag.NewFlagSet("hosts", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "print JSON")
	fs.Usage = subUsage(fs, "hosts [--json]", "List servers from ~/.ssh/config, most used first.")
	_ = fs.Parse(args)

	hosts, err := sshconf.Load()
	if err != nil {
		return fmt.Errorf("reading ~/.ssh/config: %w", err)
	}
	st, err := store.Open()
	if err != nil {
		return err
	}
	sorted := core.SortHosts(hosts, st)

	if *asJSON {
		out := make([]hostJSON, len(sorted))
		for i, h := range sorted {
			out[i] = hostJSON{h.Alias, h.HostName, h.User, st.HostUse(h.Alias).Count}
		}
		return printJSON(out)
	}
	for _, h := range sorted {
		fmt.Printf("%s\t%s\n", h.Alias, core.Describe(h))
	}
	return nil
}

func cmdDirs(args []string) error {
	fs := flag.NewFlagSet("dirs", flag.ExitOnError)
	host := fs.String("host", "", "server alias (required)")
	asJSON := fs.Bool("json", false, "print JSON")
	fs.Usage = subUsage(fs, "dirs --host ALIAS [--json]", "List target directories for a server, most used first.")
	_ = fs.Parse(args)
	if *host == "" {
		return errors.New("dirs: --host is required")
	}

	st, err := store.Open()
	if err != nil {
		return err
	}
	dirs := st.Dirs(*host)
	if *asJSON {
		if dirs == nil {
			dirs = []string{}
		}
		return printJSON(dirs)
	}
	for _, d := range dirs {
		fmt.Println(d)
	}
	return nil
}

func cmdSend(args []string) error {
	fs := flag.NewFlagSet("send", flag.ExitOnError)
	host := fs.String("host", "", "server alias (required)")
	dir := fs.String("dir", "", "target directory, created if missing; ~ is expanded remotely (required)")
	asJSON := fs.Bool("json", false, "print JSON")
	noClip := fs.Bool("no-clipboard", false, "do not copy the remote paths to the clipboard")
	fs.Usage = subUsage(fs, "send --host ALIAS --dir DIR [--json] [--no-clipboard] file ...",
		"Send files without prompting and print their remote paths.")
	_ = fs.Parse(args)

	files := fs.Args()
	switch {
	case *host == "":
		return errors.New("send: --host is required")
	case *dir == "":
		return errors.New("send: --dir is required")
	case len(files) == 0:
		return errors.New("send: no files given")
	}
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			return fmt.Errorf("cannot access %s: %w", f, err)
		}
	}

	st, err := store.Open()
	if err != nil {
		return err
	}
	absDir, remotePaths, err := core.Send(*host, *dir, files)
	if err != nil {
		return err
	}
	_ = st.Touch(*host, *dir)

	var clip string
	if !*noClip {
		clip, _ = clipboard.Copy(strings.Join(remotePaths, "\n"))
	}
	if *asJSON {
		return printJSON(sendJSON{*host, absDir, remotePaths, clip})
	}
	for _, rp := range remotePaths {
		fmt.Printf("%s:%s\n", *host, rp)
	}
	return nil
}

func cmdLast(args []string) error {
	fs := flag.NewFlagSet("last", flag.ExitOnError)
	asJSON := fs.Bool("json", false, "print JSON")
	fs.Usage = subUsage(fs, "last [--json]", "Print the most recently used server and directory.")
	_ = fs.Parse(args)

	st, err := store.Open()
	if err != nil {
		return err
	}
	host, dir, ok := st.Last()
	if !ok {
		return errors.New("nothing has been sent yet")
	}
	if *asJSON {
		return printJSON(lastJSON{host, dir})
	}
	fmt.Printf("%s\t%s\n", host, dir)
	return nil
}

func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func subUsage(fs *flag.FlagSet, synopsis, about string) func() {
	return func() {
		fmt.Fprintf(fs.Output(), "Usage: drop-box-cli %s\n\n%s\n\n", synopsis, about)
		fs.PrintDefaults()
	}
}
