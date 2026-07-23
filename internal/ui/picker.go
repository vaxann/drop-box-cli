package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// pickItem is one selectable row.
type pickItem struct {
	Label string // primary text, matched by the filter
	Desc  string // dimmed suffix, also matched by the filter
}

// picker is a minimal filterable list: type to filter, arrows to move,
// digits 1-9 to quick-select while the filter is empty.
type picker struct {
	title    string
	items    []pickItem
	filter   string
	filtered []int // indexes into items
	cursor   int   // index into filtered
}

func newPicker(title string, items []pickItem) picker {
	p := picker{title: title, items: items}
	p.refilter()
	return p
}

func (p *picker) refilter() {
	p.filtered = p.filtered[:0]
	q := strings.ToLower(p.filter)
	for i, it := range p.items {
		if q == "" || strings.Contains(strings.ToLower(it.Label+" "+it.Desc), q) {
			p.filtered = append(p.filtered, i)
		}
	}
	if p.cursor >= len(p.filtered) {
		p.cursor = len(p.filtered) - 1
	}
	if p.cursor < 0 {
		p.cursor = 0
	}
}

// pickerResult reports what a key press did.
type pickerResult int

const (
	pickerNone pickerResult = iota
	pickerChosen
	pickerCancel
)

// Update handles a key press; on pickerChosen, Selected() is valid.
func (p *picker) Update(msg tea.KeyMsg) pickerResult {
	switch msg.String() {
	case "esc":
		if p.filter != "" {
			p.filter = ""
			p.refilter()
			return pickerNone
		}
		return pickerCancel
	case "enter":
		if len(p.filtered) > 0 {
			return pickerChosen
		}
		return pickerNone
	case "up", "ctrl+p":
		if p.cursor > 0 {
			p.cursor--
		}
		return pickerNone
	case "down", "ctrl+n":
		if p.cursor < len(p.filtered)-1 {
			p.cursor++
		}
		return pickerNone
	case "backspace":
		if p.filter != "" {
			p.filter = p.filter[:len(p.filter)-1]
			p.refilter()
		}
		return pickerNone
	}
	if msg.Type == tea.KeyRunes {
		s := string(msg.Runes)
		// Quick select by number while not filtering.
		if p.filter == "" && len(s) == 1 && s >= "1" && s <= "9" {
			n := int(s[0] - '0')
			if n <= len(p.filtered) {
				p.cursor = n - 1
				return pickerChosen
			}
			return pickerNone
		}
		p.filter += s
		p.refilter()
	}
	return pickerNone
}

// Selected returns the index (into the original items) under the cursor,
// or -1 when nothing matches the filter.
func (p *picker) Selected() int {
	if len(p.filtered) == 0 {
		return -1
	}
	return p.filtered[p.cursor]
}

const maxVisible = 12

func (p *picker) View() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(p.title) + "\n")
	if p.filter != "" {
		b.WriteString(filterStyle.Render("filter: "+p.filter) + "\n")
	}
	if len(p.filtered) == 0 {
		b.WriteString(dimStyle.Render("  nothing matches") + "\n")
		return b.String()
	}
	start := 0
	if p.cursor >= maxVisible {
		start = p.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(p.filtered) {
		end = len(p.filtered)
	}
	for row := start; row < end; row++ {
		it := p.items[p.filtered[row]]
		num := fmt.Sprintf("%d.", row+1)
		if row >= 9 {
			num = "  "
		}
		line := fmt.Sprintf("%-3s %s", num, it.Label)
		if it.Desc != "" {
			line += " " + dimStyle.Render(it.Desc)
		}
		if row == p.cursor {
			b.WriteString(cursorStyle.Render("▸ ") + selectedStyle.Render(line) + "\n")
		} else {
			b.WriteString("  " + line + "\n")
		}
	}
	if len(p.filtered) > maxVisible {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  … %d/%d", p.cursor+1, len(p.filtered))) + "\n")
	}
	return b.String()
}
