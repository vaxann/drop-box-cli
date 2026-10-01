// Package store keeps per-host target directories and usage statistics:
// preconfigured directories from config.yaml plus use counts and last-use
// times in history.json, so frequently used servers and directories can be
// offered first.
package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	appDir         = "drop-box-cli"
	maxHistPerHost = 50
)

// Config is the user-editable configuration file.
type Config struct {
	Servers map[string]ServerConfig `yaml:"servers"`
}

// ServerConfig holds the preconfigured directories for one host alias.
type ServerConfig struct {
	Dirs []string `yaml:"dirs"`
}

// HistEntry is one remembered target directory.
type HistEntry struct {
	Path     string    `json:"path"`
	Count    int       `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

// HostUse is the usage record of one server.
type HostUse struct {
	Count    int       `json:"count"`
	LastUsed time.Time `json:"last_used"`
}

// histFile is the on-disk layout of history.json.
type histFile struct {
	Hosts map[string]HostUse     `json:"hosts"`
	Dirs  map[string][]HistEntry `json:"dirs"`
}

// Store combines config and history.
type Store struct {
	dir   string
	cfg   Config
	hosts map[string]HostUse
	dirs  map[string][]HistEntry
}

// Open loads (or lazily creates) the store under the user config dir.
func Open() (*Store, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	dir := filepath.Join(base, appDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	s := &Store{
		dir:   dir,
		hosts: map[string]HostUse{},
		dirs:  map[string][]HistEntry{},
	}

	if data, err := os.ReadFile(filepath.Join(dir, "config.yaml")); err == nil {
		if err := yaml.Unmarshal(data, &s.cfg); err != nil {
			return nil, err
		}
	}
	if data, err := os.ReadFile(filepath.Join(dir, "history.json")); err == nil {
		s.loadHistory(data)
	}
	return s, nil
}

// loadHistory accepts both the current {hosts, dirs} layout and the legacy
// flat map of host -> entries. A corrupt file must not brick the tool.
func (s *Store) loadHistory(data []byte) {
	var f histFile
	if err := json.Unmarshal(data, &f); err == nil && (f.Hosts != nil || f.Dirs != nil) {
		if f.Hosts != nil {
			s.hosts = f.Hosts
		}
		if f.Dirs != nil {
			s.dirs = f.Dirs
		}
		return
	}
	var legacy map[string][]HistEntry
	if err := json.Unmarshal(data, &legacy); err == nil {
		for host, entries := range legacy {
			for i := range entries {
				if entries[i].Count == 0 {
					entries[i].Count = 1
				}
			}
			s.dirs[host] = entries
		}
	}
}

// ConfigPath returns the path of the user-editable config file.
func (s *Store) ConfigPath() string {
	return filepath.Join(s.dir, "config.yaml")
}

// byUsage orders higher use counts first, breaking ties by recency.
func byUsage(ci, cj int, ti, tj time.Time) bool {
	if ci != cj {
		return ci > cj
	}
	return ti.After(tj)
}

// Dirs returns target directories for a host: most frequently used first
// (ties broken by recency), then preconfigured dirs that were never used.
func (s *Store) Dirs(host string) []string {
	seen := map[string]bool{}
	var out []string

	entries := append([]HistEntry(nil), s.dirs[host]...)
	sort.SliceStable(entries, func(i, j int) bool {
		return byUsage(entries[i].Count, entries[j].Count, entries[i].LastUsed, entries[j].LastUsed)
	})
	for _, e := range entries {
		if !seen[e.Path] {
			seen[e.Path] = true
			out = append(out, e.Path)
		}
	}
	for _, d := range s.cfg.Servers[host].Dirs {
		if !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

// HostUse reports the usage record for a host alias.
func (s *Store) HostUse(alias string) HostUse {
	return s.hosts[alias]
}

// Touch records that dir on host was just used and persists the history.
func (s *Store) Touch(host, dir string) error {
	now := time.Now()

	use := s.hosts[host]
	use.Count++
	use.LastUsed = now
	s.hosts[host] = use

	entries := s.dirs[host]
	found := false
	for i := range entries {
		if entries[i].Path == dir {
			entries[i].Count++
			entries[i].LastUsed = now
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, HistEntry{Path: dir, Count: 1, LastUsed: now})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return byUsage(entries[i].Count, entries[j].Count, entries[i].LastUsed, entries[j].LastUsed)
	})
	if len(entries) > maxHistPerHost {
		entries = entries[:maxHistPerHost]
	}
	s.dirs[host] = entries

	data, err := json.MarshalIndent(histFile{Hosts: s.hosts, Dirs: s.dirs}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "history.json"), data, 0o600)
}

// Last reports the most recently used host and, on it, the most recently
// used directory. ok is false when nothing has been sent yet.
func (s *Store) Last() (host, dir string, ok bool) {
	var hostTime time.Time
	for h, use := range s.hosts {
		if len(s.dirs[h]) > 0 && use.LastUsed.After(hostTime) {
			host, hostTime = h, use.LastUsed
		}
	}
	if host == "" {
		return "", "", false
	}
	var dirTime time.Time
	for _, e := range s.dirs[host] {
		if dir == "" || e.LastUsed.After(dirTime) {
			dir, dirTime = e.Path, e.LastUsed
		}
	}
	return host, dir, true
}
