package droppath

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"plain", "/tmp/cat.png", []string{"/tmp/cat.png"}},
		{"trailing space", "/tmp/cat.png ", []string{"/tmp/cat.png"}},
		{"single quoted", "'/tmp/my cat.png'", []string{"/tmp/my cat.png"}},
		{"double quoted", `"/tmp/my cat.png"`, []string{"/tmp/my cat.png"}},
		{"backslash escaped", `/tmp/my\ cat.png`, []string{"/tmp/my cat.png"}},
		{"multiple", "/tmp/a.png /tmp/b.png", []string{"/tmp/a.png", "/tmp/b.png"}},
		{"multiple quoted", "'/tmp/a 1.png' '/tmp/b 2.png'", []string{"/tmp/a 1.png", "/tmp/b 2.png"}},
		{"file uri", "file:///tmp/my%20cat.png", []string{"/tmp/my cat.png"}},
		{"file uri with host", "file://localhost/tmp/cat.png", []string{"/tmp/cat.png"}},
		{"quote inside name", `'/tmp/it'\''s.png'`, []string{"/tmp/it's.png"}},
		{"empty", "   ", nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Parse(c.in)
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Parse(%q) = %#v, want %#v", c.in, got, c.want)
			}
		})
	}
}
