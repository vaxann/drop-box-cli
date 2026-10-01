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

func TestCopyCmd(t *testing.T) {
	cmd := CopyCmd("web", []string{"/tmp/a.png", "/data/my file (2).png"}, "/srv/in box")
	want := []string{"sh", "-c",
		`tar -cf - -C '/tmp/' './a.png' -C '/data/' './my file (2).png'` +
			` | ssh -- 'web' 'cd -- '\''/srv/in box'\'' && tar -xf -'`}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("args = %#v, want %#v", cmd.Args, want)
	}
}

func TestCopyCmdTrailingSlashDir(t *testing.T) {
	cmd := CopyCmd("web", []string{"/tmp/some dir/"}, "/dst")
	want := []string{"sh", "-c",
		`tar -cf - -C '/tmp/' './some dir' | ssh -- 'web' 'cd -- '\''/dst'\'' && tar -xf -'`}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("args = %#v, want %#v", cmd.Args, want)
	}
}

func TestResolveCmdArgs(t *testing.T) {
	cmd := ResolveCmd("web", "~/in")
	want := []string{"ssh", "--", "web", `mkdir -p -- "$HOME"/'in' && cd -- "$HOME"/'in' && pwd`}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("args = %v, want %v", cmd.Args, want)
	}
}

func TestBatchOptions(t *testing.T) {
	cmd := ResolveCmd("web", "/in", BatchOptions...)
	want := []string{"ssh", "-o", "BatchMode=yes", "-o", "ConnectTimeout=10", "--", "web",
		`mkdir -p -- '/in' && cd -- '/in' && pwd`}
	if !reflect.DeepEqual(cmd.Args, want) {
		t.Errorf("resolve args = %v, want %v", cmd.Args, want)
	}

	cmd = CopyCmd("web", []string{"/tmp/a.png"}, "/in", BatchOptions...)
	wantSh := `tar -cf - -C '/tmp/' './a.png'` +
		` | ssh '-o' 'BatchMode=yes' '-o' 'ConnectTimeout=10' -- 'web' 'cd -- '\''/in'\'' && tar -xf -'`
	if cmd.Args[2] != wantSh {
		t.Errorf("copy script = %s, want %s", cmd.Args[2], wantSh)
	}
}
