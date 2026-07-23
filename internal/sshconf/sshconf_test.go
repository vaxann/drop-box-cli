package sshconf

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	main := filepath.Join(dir, "config")
	extra := filepath.Join(dir, "extra")

	write(t, extra, "Host included\n  HostName inc.example.com\n")
	write(t, main, `
# comment
Host *
  ServerAliveInterval 60

Host web db
  User deploy

Host web
  HostName web.example.com

Host "quoted host"
  HostName q.example.com

Host wild-*
  HostName ignored.example.com

Include `+extra+`

Match user deploy
  HostName should-not-attach.example.com
`)

	hosts, err := LoadFile(main)
	if err != nil {
		t.Fatal(err)
	}

	byAlias := map[string]Host{}
	var order []string
	for _, h := range hosts {
		byAlias[h.Alias] = h
		order = append(order, h.Alias)
	}

	wantOrder := []string{"web", "db", "quoted host", "included"}
	if len(order) != len(wantOrder) {
		t.Fatalf("got hosts %v, want %v", order, wantOrder)
	}
	for i := range wantOrder {
		if order[i] != wantOrder[i] {
			t.Fatalf("got hosts %v, want %v", order, wantOrder)
		}
	}

	if h := byAlias["web"]; h.HostName != "web.example.com" || h.User != "deploy" {
		t.Errorf("web = %+v", h)
	}
	if h := byAlias["db"]; h.User != "deploy" || h.HostName != "" {
		t.Errorf("db = %+v", h)
	}
	if h := byAlias["included"]; h.HostName != "inc.example.com" {
		t.Errorf("included = %+v", h)
	}
}

func TestLoadFileMissing(t *testing.T) {
	if _, err := LoadFile(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Error("expected error for missing top-level config")
	}
}
