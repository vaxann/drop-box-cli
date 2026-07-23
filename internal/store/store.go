// Package store keeps per-host target directories: preconfigured ones from
// config.yaml and a most-recently-used history in history.json.
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
	LastUsed time.Time `json:"last_used"`
}

// Store combines config and history.
type Store struct {
	dir  string
	cfg  Config
	hist map[string][]HistEntry
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
	s := &Store{dir: dir, hist: map[string][]HistEntry{}}

	if data, err := os.ReadFile(filepath.Join(dir, "config.yaml")); err == nil {
		if err := yaml.Unmarshal(data, &s.cfg); err != nil {
			return nil, err
		}
	}
	if data, err := os.ReadFile(filepath.Join(dir, "history.json")); err == nil {
		if err := json.Unmarshal(data, &s.hist); err != nil {
			// A corrupt history should not brick the tool.
			s.hist = map[string][]HistEntry{}
		}
	}
	return s, nil
}

// ConfigPath returns the path of the user-editable config file.
func (s *Store) ConfigPath() string {
	return filepath.Join(s.dir, "config.yaml")
}

// Dirs returns target directories for a host: history first (most recently
// used on top), then preconfigured dirs that were never used.
func (s *Store) Dirs(host string) []string {
	seen := map[string]bool{}
	var out []string

	entries := append([]HistEntry(nil), s.hist[host]...)
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].LastUsed.After(entries[j].LastUsed)
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

// Touch records that dir was just used for host and persists the history.
func (s *Store) Touch(host, dir string) error {
	now := time.Now()
	entries := s.hist[host]
	found := false
	for i := range entries {
		if entries[i].Path == dir {
			entries[i].LastUsed = now
			found = true
			break
		}
	}
	if !found {
		entries = append(entries, HistEntry{Path: dir, LastUsed: now})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].LastUsed.After(entries[j].LastUsed)
	})
	if len(entries) > maxHistPerHost {
		entries = entries[:maxHistPerHost]
	}
	s.hist[host] = entries

	data, err := json.MarshalIndent(s.hist, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, "history.json"), data, 0o600)
}
