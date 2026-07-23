package transfer

import (
	"reflect"
	"testing"
)

func TestRemoteArg(t *testing.T) {
	cases := map[string]string{
		"~":            `"$HOME"`,
		"~/in box":     `"$HOME"/'in box'`,
		"/var/www":     `'/var/www'`,
		"/it's":        `'/it'\''s'`,
		"relative/dir": `'relative/dir'`,
	}
	for in, want := range cases {
		if got := RemoteArg(in); got != want {
			t.Errorf("RemoteArg(%q) = %s, want %s", in, got, want)
		}
	}
}

func TestEscapeRemote(t *testing.T) {
	if got := EscapeRemote("/srv/my files/in(1)"); got != `/srv/my\ files/in\(1\)` {
		t.Errorf("EscapeRemote = %s", got)
	}
}

func TestScpCmdArgs(t *testing.T) {
	cmd := ScpCmd("web", []string{"/tmp/a.png", "b:tricky.png"}, "/srv/in box")
	want := []string{"scp", "-r", "--", "/tmp/a.png", "./b:tricky.png", `web:/srv/in\ box/`}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("args = %v, want %v", cmd.Args, want)
	}
}

func TestResolveCmdArgs(t *testing.T) {
	cmd := ResolveCmd("web", "~/in")
	want := []string{"ssh", "--", "web", `mkdir -p -- "$HOME"/'in' && cd -- "$HOME"/'in' && pwd`}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("args = %v, want %v", cmd.Args, want)
	}
}
