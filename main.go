// drop-box-cli: drag-and-drop a file into a terminal window and beam it to
// an SSH server; the remote absolute path is printed and copied to the
// clipboard.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/vaxann/drop-box-cli/internal/sshconf"
	"github.com/vaxann/drop-box-cli/internal/store"
	"github.com/vaxann/drop-box-cli/internal/ui"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: drop-box-cli [file ...]\n\n"+
				"Without arguments, waits for files dragged into the terminal (loop mode).\n"+
				"With file arguments, sends them once and exits.\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *showVersion {
		fmt.Println("drop-box-cli", version)
		return
	}

	files := flag.Args()
	for _, f := range files {
		if _, err := os.Stat(f); err != nil {
			fatal(fmt.Sprintf("cannot access %s: %v", f, err))
		}
	}

	hosts, err := sshconf.Load()
	if err != nil {
		fatal(fmt.Sprintf("reading ~/.ssh/config: %v", err))
	}
	if len(hosts) == 0 {
		fatal("no Host entries found in ~/.ssh/config — add your servers there first")
	}

	st, err := store.Open()
	if err != nil {
		fatal(fmt.Sprintf("opening config store: %v", err))
	}

	p := tea.NewProgram(ui.New(hosts, st, files))
	if _, err := p.Run(); err != nil {
		fatal(err.Error())
	}
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, "drop-box-cli:", msg)
	os.Exit(1)
}
