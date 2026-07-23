// Package ui implements the interactive terminal flow: wait for a dropped
// file, pick a server, pick a directory, transfer, report the remote path.
package ui

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/vaxann/drop-box-cli/internal/clipboard"
	"github.com/vaxann/drop-box-cli/internal/droppath"
	"github.com/vaxann/drop-box-cli/internal/sshconf"
	"github.com/vaxann/drop-box-cli/internal/store"
	"github.com/vaxann/drop-box-cli/internal/transfer"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	cursorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	selectedStyle = lipgloss.NewStyle().Bold(true)
	dimStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	filterStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	okStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
)

type state int

const (
	stateWaiting state = iota
	statePickServer
	statePickDir
	stateNewDir
	stateTransfer
)

const newDirLabel = "✎  enter a new path…"

type resolveDoneMsg struct {
	dir string
	err error
}

type scpDoneMsg struct {
	absDir string
	err    error
}

// Model is the root Bubble Tea model.
type Model struct {
	hosts   []sshconf.Host
	st      *store.Store
	oneShot bool

	state    state
	drop     textinput.Model
	dirInput textinput.Model
	pick     picker

	files     []string // local files being sent
	host      string   // chosen host alias
	chosenDir string   // dir as typed/configured (may contain ~)

	errMsg string
}

// New builds the model. If files is non-empty the model starts at server
// selection and quits after the transfer (one-shot mode).
func New(hosts []sshconf.Host, st *store.Store, files []string) Model {
	drop := textinput.New()
	drop.Placeholder = "drop a file here, then press Enter"
	drop.Prompt = "→ "
	drop.Focus()

	dirInput := textinput.New()
	dirInput.Placeholder = "/remote/target/dir or ~/dir"
	dirInput.Prompt = "→ "

	m := Model{hosts: hosts, st: st, drop: drop, dirInput: dirInput}
	if len(files) > 0 {
		m.oneShot = true
		m.files = files
		m.state = statePickServer
		m.pick = m.serverPicker()
	}
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) serverPicker() picker {
	items := make([]pickItem, len(m.hosts))
	for i, h := range m.hosts {
		desc := h.HostName
		if h.User != "" && desc != "" {
			desc = h.User + "@" + desc
		}
		items[i] = pickItem{Label: h.Alias, Desc: desc}
	}
	title := "Send to which server?"
	if len(m.files) == 1 {
		title = fmt.Sprintf("Send %q to which server?", path.Base(m.files[0]))
	} else {
		title = fmt.Sprintf("Send %d files to which server?", len(m.files))
	}
	return newPicker(title, items)
}

func (m Model) dirPicker() picker {
	dirs := m.st.Dirs(m.host)
	items := make([]pickItem, 0, len(dirs)+1)
	for _, d := range dirs {
		items = append(items, pickItem{Label: d})
	}
	items = append(items, pickItem{Label: newDirLabel})
	return newPicker(fmt.Sprintf("Target directory on %s:", m.host), items)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m.updateKey(msg)
	case resolveDoneMsg:
		if msg.err != nil || msg.dir == "" {
			m.errMsg = fmt.Sprintf("resolving directory on %s failed: %v", m.host, msg.err)
			m.state = statePickDir
			m.pick = m.dirPicker()
			return m, nil
		}
		abs := msg.dir
		return m, tea.ExecProcess(
			transfer.CopyCmd(m.host, m.files, abs),
			func(err error) tea.Msg { return scpDoneMsg{absDir: abs, err: err} },
		)
	case scpDoneMsg:
		return m.finishTransfer(msg.absDir, msg.err)
	}

	var cmd tea.Cmd
	switch m.state {
	case stateWaiting:
		m.drop, cmd = m.drop.Update(msg)
	case stateNewDir:
		m.dirInput, cmd = m.dirInput.Update(msg)
	}
	return m, cmd
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state {
	case stateWaiting:
		if msg.String() == "esc" {
			return m, tea.Quit
		}
		if msg.String() == "enter" {
			line := strings.TrimSpace(m.drop.Value())
			if line == "" {
				return m, nil
			}
			files := droppath.Parse(line)
			var bad []string
			for _, f := range files {
				if _, err := os.Stat(f); err != nil {
					bad = append(bad, f)
				}
			}
			if len(files) == 0 || len(bad) > 0 {
				m.errMsg = "not a file: " + strings.Join(bad, ", ")
				return m, nil
			}
			m.errMsg = ""
			m.files = files
			m.drop.SetValue("")
			m.state = statePickServer
			m.pick = m.serverPicker()
			return m, nil
		}
		var cmd tea.Cmd
		m.drop, cmd = m.drop.Update(msg)
		return m, cmd

	case statePickServer:
		switch m.pick.Update(msg) {
		case pickerChosen:
			m.host = m.hosts[m.pick.Selected()].Alias
			m.errMsg = ""
			dirs := m.st.Dirs(m.host)
			if len(dirs) == 0 {
				m.state = stateNewDir
				m.dirInput.SetValue("")
				m.dirInput.Focus()
				return m, textinput.Blink
			}
			m.state = statePickDir
			m.pick = m.dirPicker()
		case pickerCancel:
			if m.oneShot {
				return m, tea.Quit
			}
			m.state = stateWaiting
		}
		return m, nil

	case statePickDir:
		switch m.pick.Update(msg) {
		case pickerChosen:
			label := m.pick.items[m.pick.Selected()].Label
			if label == newDirLabel {
				m.state = stateNewDir
				m.dirInput.SetValue("")
				m.dirInput.Focus()
				return m, textinput.Blink
			}
			return m.startTransfer(label)
		case pickerCancel:
			m.state = statePickServer
			m.pick = m.serverPicker()
		}
		return m, nil

	case stateNewDir:
		switch msg.String() {
		case "esc":
			m.state = statePickDir
			m.pick = m.dirPicker()
			return m, nil
		case "enter":
			dir := strings.TrimSpace(m.dirInput.Value())
			if dir == "" {
				return m, nil
			}
			return m.startTransfer(dir)
		}
		var cmd tea.Cmd
		m.dirInput, cmd = m.dirInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m Model) startTransfer(dir string) (tea.Model, tea.Cmd) {
	m.chosenDir = dir
	m.errMsg = ""
	m.state = stateTransfer
	buf := &bytes.Buffer{}
	cmd := transfer.ResolveCmd(m.host, dir)
	cmd.Stdout = buf // tea.ExecProcess only fills stdio it finds unset
	return m, tea.ExecProcess(cmd, func(err error) tea.Msg {
		return resolveDoneMsg{dir: strings.TrimSpace(buf.String()), err: err}
	})
}

func (m Model) finishTransfer(absDir string, err error) (tea.Model, tea.Cmd) {
	if err != nil {
		m.errMsg = fmt.Sprintf("scp failed: %v", err)
		m.state = statePickDir
		m.pick = m.dirPicker()
		if m.oneShot {
			return m, tea.Sequence(tea.Println(errStyle.Render("✗ "+m.errMsg)), tea.Quit)
		}
		return m, nil
	}
	_ = m.st.Touch(m.host, m.chosenDir)

	remotePaths := make([]string, len(m.files))
	for i, f := range m.files {
		remotePaths[i] = absDir + "/" + path.Base(strings.TrimRight(f, "/"))
	}
	clipNote := "clipboard unavailable"
	if tool, cerr := clipboard.Copy(strings.Join(remotePaths, "\n")); cerr == nil {
		clipNote = "copied to clipboard via " + tool
	}

	var lines []string
	for _, rp := range remotePaths {
		lines = append(lines, okStyle.Render("✓ ")+m.host+":"+rp)
	}
	lines = append(lines, dimStyle.Render("  "+clipNote))
	printed := strings.Join(lines, "\n")

	m.files = nil
	m.state = stateWaiting
	m.drop.Focus()
	if m.oneShot {
		return m, tea.Sequence(tea.Println(printed), tea.Quit)
	}
	return m, tea.Sequence(tea.Println(printed), textinput.Blink)
}

func (m Model) View() string {
	var b strings.Builder
	switch m.state {
	case stateWaiting:
		b.WriteString(titleStyle.Render("drop-box-cli") + dimStyle.Render("  — drag a file into this window") + "\n")
		b.WriteString(m.drop.View() + "\n")
		b.WriteString(dimStyle.Render("enter: send · esc: quit") + "\n")
	case statePickServer, statePickDir:
		b.WriteString(m.pick.View())
		b.WriteString(dimStyle.Render("type to filter · 1-9/enter: choose · esc: back") + "\n")
	case stateNewDir:
		b.WriteString(titleStyle.Render(fmt.Sprintf("New directory on %s:", m.host)) + "\n")
		b.WriteString(m.dirInput.View() + "\n")
		b.WriteString(dimStyle.Render("enter: send · esc: back") + "\n")
	case stateTransfer:
		b.WriteString(dimStyle.Render("transferring…") + "\n")
	}
	if m.errMsg != "" {
		b.WriteString(errStyle.Render("✗ "+m.errMsg) + "\n")
	}
	return b.String()
}
