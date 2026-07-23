package store

import (
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

func TestDirsMergeAndOrder(t *testing.T) {
	s := openTemp(t)
	s.cfg.Servers = map[string]ServerConfig{
		"web": {Dirs: []string{"/var/www/uploads", "~/incoming"}},
	}
	s.hist["web"] = []HistEntry{
		{Path: "/tmp/old", LastUsed: time.Now().Add(-time.Hour)},
		{Path: "~/incoming", LastUsed: time.Now()},
	}

	got := s.Dirs("web")
	want := []string{"~/incoming", "/tmp/old", "/var/www/uploads"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Dirs = %v, want %v", got, want)
	}
}

func TestTouchPersists(t *testing.T) {
	base := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", base)

	s, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Touch("web", "/data"); err != nil {
		t.Fatal(err)
	}
	if err := s.Touch("web", "/other"); err != nil {
		t.Fatal(err)
	}
	if err := s.Touch("web", "/data"); err != nil { // /data becomes most recent again
		t.Fatal(err)
	}

	if _, err := os.Stat(filepath.Join(base, "drop-box-cli", "history.json")); err != nil {
		t.Fatalf("history.json not written: %v", err)
	}

	s2, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	got := s2.Dirs("web")
	want := []string{"/data", "/other"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("after reopen Dirs = %v, want %v", got, want)
	}
}

func TestDirsUnknownHost(t *testing.T) {
	s := openTemp(t)
	if got := s.Dirs("nope"); got != nil {
		t.Errorf("Dirs for unknown host = %v, want nil", got)
	}
}
