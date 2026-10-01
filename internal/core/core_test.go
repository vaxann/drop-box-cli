package core

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/vaxann/drop-box-cli/internal/sshconf"
	"github.com/vaxann/drop-box-cli/internal/store"
)

// stubSSH puts a fake ssh first on PATH: it skips options up to "--" and
// the host, then runs the remote command locally.
func stubSSH(t *testing.T, script string) {
	t.Helper()
	bin := t.TempDir()
	if err := os.WriteFile(filepath.Join(bin, "ssh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
}

func TestSendEndToEnd(t *testing.T) {
	stubSSH(t, "#!/bin/sh\nwhile [ \"$1\" != \"--\" ]; do shift; done\nshift 2\nexec sh -c \"$1\"\n")

	src := filepath.Join(t.TempDir(), "my shot.png")
	if err := os.WriteFile(src, []byte("png"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(t.TempDir(), "new dir")

	absDir, paths, err := Send("box", dst, []string{src})
	if err != nil {
		t.Fatal(err)
	}
	if absDir != dst {
		t.Errorf("absDir = %q, want %q", absDir, dst)
	}
	want := []string{dst + "/my shot.png"}
	if !reflect.DeepEqual(paths, want) {
		t.Errorf("paths = %v, want %v", paths, want)
	}
	if got, err := os.ReadFile(want[0]); err != nil || string(got) != "png" {
		t.Errorf("remote file = %q, %v", got, err)
	}
}

func TestSendReportsSSHError(t *testing.T) {
	stubSSH(t, "#!/bin/sh\necho 'box: Permission denied (publickey).' >&2\nexit 255\n")

	_, _, err := Send("box", "/in", []string{"/etc/hostname"})
	if err == nil || !strings.Contains(err.Error(), "Permission denied (publickey).") {
		t.Errorf("err = %v, want the ssh stderr in it", err)
	}
}

func TestSortHosts(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	st, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range []string{"b", "c", "b"} {
		if err := st.Touch(h, "/d"); err != nil {
			t.Fatal(err)
		}
	}
	hosts := []sshconf.Host{{Alias: "a"}, {Alias: "b"}, {Alias: "c"}}
	var got []string
	for _, h := range SortHosts(hosts, st) {
		got = append(got, h.Alias)
	}
	if want := []string{"b", "c", "a"}; !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}

func TestDescribe(t *testing.T) {
	for h, want := range map[sshconf.Host]string{
		{HostName: "x.org", User: "me"}: "me@x.org",
		{HostName: "x.org"}:             "x.org",
		{User: "me"}:                    "",
	} {
		if got := Describe(h); got != want {
			t.Errorf("Describe(%+v) = %q, want %q", h, got, want)
		}
	}
}
