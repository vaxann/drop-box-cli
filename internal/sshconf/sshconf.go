// Package sshconf lists concrete host aliases declared in ~/.ssh/config,
// following Include directives and skipping wildcard patterns.
package sshconf

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// Host is one selectable SSH destination.
type Host struct {
	Alias    string
	HostName string
	User     string
}

const maxIncludeDepth = 8

// serviceHosts are well-known code-hosting endpoints that are never file
// transfer destinations, so they are hidden from the server list.
var serviceHosts = map[string]bool{
	"github.com":              true,
	"ssh.github.com":          true,
	"gist.github.com":         true,
	"gitlab.com":              true,
	"bitbucket.org":           true,
	"codeberg.org":            true,
	"git.sr.ht":               true,
	"ssh.dev.azure.com":       true,
	"vs-ssh.visualstudio.com": true,
}

// Load reads hosts from ~/.ssh/config.
func Load() ([]Host, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return LoadFile(filepath.Join(home, ".ssh", "config"))
}

// LoadFile reads hosts from the given ssh config file.
func LoadFile(path string) ([]Host, error) {
	p := &parser{
		seen:    map[string]bool{},
		byAlias: map[string]*Host{},
	}
	if err := p.parseFile(path, 0); err != nil {
		return nil, err
	}
	hosts := make([]Host, 0, len(p.order))
	for _, alias := range p.order {
		h := *p.byAlias[alias]
		if serviceHosts[strings.ToLower(h.Alias)] || serviceHosts[strings.ToLower(h.HostName)] {
			continue
		}
		hosts = append(hosts, h)
	}
	return hosts, nil
}

type parser struct {
	seen    map[string]bool
	byAlias map[string]*Host
	order   []string
	current []*Host
}

func (p *parser) parseFile(path string, depth int) error {
	if depth > maxIncludeDepth || p.seen[path] {
		return nil
	}
	p.seen[path] = true
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && depth > 0 {
			return nil
		}
		return err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := splitLine(line)
		if len(fields) == 0 {
			continue
		}
		key := strings.ToLower(fields[0])
		switch key {
		case "host":
			p.current = nil
			for _, pat := range fields[1:] {
				if strings.ContainsAny(pat, "*?!") || pat == "" {
					continue
				}
				h, ok := p.byAlias[pat]
				if !ok {
					h = &Host{Alias: pat}
					p.byAlias[pat] = h
					p.order = append(p.order, pat)
				}
				p.current = append(p.current, h)
			}
		case "match":
			p.current = nil
		case "include":
			for _, pat := range fields[1:] {
				p.include(pat, depth)
			}
		case "hostname":
			if len(fields) > 1 {
				for _, h := range p.current {
					if h.HostName == "" {
						h.HostName = fields[1]
					}
				}
			}
		case "user":
			if len(fields) > 1 {
				for _, h := range p.current {
					if h.User == "" {
						h.User = fields[1]
					}
				}
			}
		}
	}
	return sc.Err()
}

func (p *parser) include(pattern string, depth int) {
	if strings.HasPrefix(pattern, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			pattern = filepath.Join(home, pattern[2:])
		}
	}
	if !filepath.IsAbs(pattern) {
		if home, err := os.UserHomeDir(); err == nil {
			pattern = filepath.Join(home, ".ssh", pattern)
		}
	}
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return
	}
	for _, m := range matches {
		_ = p.parseFile(m, depth+1)
	}
}

// splitLine tokenizes a config line ssh-style: whitespace-separated, with
// double quotes allowing embedded spaces.
func splitLine(line string) []string {
	var (
		out     []string
		cur     strings.Builder
		inQuote bool
		has     bool
	)
	flush := func() {
		if has {
			out = append(out, cur.String())
		}
		cur.Reset()
		has = false
	}
	for _, r := range line {
		switch {
		case r == '"':
			inQuote = !inQuote
			has = true
		case (r == ' ' || r == '\t') && !inQuote:
			flush()
		default:
			cur.WriteRune(r)
			has = true
		}
	}
	flush()
	return out
}
