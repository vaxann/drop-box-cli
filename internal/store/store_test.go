package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func openTemp(t *testing.T) *Store {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestDirsFrequencyFirst(t *testing.T) {
	s := openTemp(t)
	s.cfg.Servers = map[string]ServerConfig{
		"web": {Dirs: []string{"/var/www/uploads", "~/incoming"}},
	}
	now := time.Now()
	s.dirs["web"] = []HistEntry{
		{Path: "/tmp/rare-but-recent", Count: 1, LastUsed: now},
		{Path: "~/incoming", Count: 5, LastUsed: now.Add(-time.Hour)},
		{Path: "/tmp/also-freq", Count: 5, LastUsed: now.Add(-time.Minute)},
	}

	got := s.Dirs("web")
	// Highest count first; equal counts by recency; then unused config dirs.
	want := []string{"/tmp/also-freq", "~/incoming", "/tmp/rare-but-recent", "/var/www/uploads"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Dirs = %v, want %v", got, want)
	}
}

func TestTouchCountsAndPersists(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{"/data", "/other", "/data"} {
		if err := s.Touch("web", dir); err != nil {
			t.Fatal(err)
		}
	}

	if use := s.HostUse("web"); use.Count != 3 {
		t.Errorf("host count = %d, want 3", use.Count)
	}

	s2, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	got := s2.Dirs("web")
	want := []string{"/data", "/other"} // /data used twice, wins over recency
	if !reflect.DeepEqual(got, want) {
		t.Errorf("after reopen Dirs = %v, want %v", got, want)
	}
	if use := s2.HostUse("web"); use.Count != 3 {
		t.Errorf("after reopen host count = %d, want 3", use.Count)
	}

	if _, err := os.Stat(filepath.Join(base, "drop-box-cli", "history.json")); err != nil {
		t.Fatalf("history.json not written: %v", err)
	}
}

func TestLegacyHistoryMigration(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	dir := filepath.Join(base, "drop-box-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := map[string][]HistEntry{
		"web": {{Path: "/old/path", LastUsed: time.Now()}},
	}
	data, _ := json.Marshal(legacy)
	if err := os.WriteFile(filepath.Join(dir, "history.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}

	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Dirs("web"); !reflect.DeepEqual(got, []string{"/old/path"}) {
		t.Errorf("migrated Dirs = %v", got)
	}
	// Touching must rewrite in the new format and keep the old entry.
	if err := s.Touch("web", "/new/path"); err != nil {
		t.Fatal(err)
	}
	s2, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	got := s2.Dirs("web")
	if len(got) != 2 || got[0] != "/new/path" && got[0] != "/old/path" {
		t.Errorf("after migration+touch Dirs = %v", got)
	}
}

func TestCorruptHistoryIgnored(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)
	dir := filepath.Join(base, "drop-box-cli")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "history.json"), []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Dirs("web"); got != nil {
		t.Errorf("Dirs from corrupt history = %v, want nil", got)
	}
}

func TestDirsUnknownHost(t *testing.T) {
	s := openTemp(t)
	if got := s.Dirs("nope"); got != nil {
		t.Errorf("Dirs for unknown host = %v, want nil", got)
	}
}
