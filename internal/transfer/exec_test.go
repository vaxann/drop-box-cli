package transfer

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCopyCmdEndToEnd runs the real tar|ssh pipeline with ssh replaced by
// a stub that executes the remote command locally, proving the quoting
// survives paths with spaces, parentheses and quotes.
func TestCopyCmdEndToEnd(t *testing.T) {
	base := t.TempDir()

	binDir := filepath.Join(base, "bin")
	srcDir := filepath.Join(base, "src dir")
	dstDir := filepath.Join(base, "Viktor Litvinov - 1-PowerLink's")
	for _, d := range []string{binDir, srcDir, dstDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// Fake ssh: drop "-- host" and run the remote command via sh.
	stub := "#!/bin/sh\nshift 2\nexec sh -c \"$1\"\n"
	if err := os.WriteFile(filepath.Join(binDir, "ssh"), []byte(stub), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	// Two files from different directories, as a multi-file drag-and-drop
	// produces, plus a nested directory.
	otherDir := filepath.Join(base, "other's dir")
	subDir := filepath.Join(srcDir, "shots dir")
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(subDir, 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(srcDir, "image (2).png"):     "content-one",
		filepath.Join(otherDir, "второй файл.png"): "content-two",
		filepath.Join(subDir, "nested.png"):        "content-three",
	}
	for p, c := range files {
		if err := os.WriteFile(p, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	srcs := []string{
		filepath.Join(srcDir, "image (2).png"),
		filepath.Join(otherDir, "второй файл.png"),
		subDir, // whole directory
	}
	cmd := CopyCmd("myhost", srcs, dstDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pipeline failed: %v\n%s", err, out)
	}

	for dst, want := range map[string]string{
		filepath.Join(dstDir, "image (2).png"):           "content-one",
		filepath.Join(dstDir, "второй файл.png"):         "content-two",
		filepath.Join(dstDir, "shots dir", "nested.png"): "content-three",
	} {
		got, err := os.ReadFile(dst)
		if err != nil {
			t.Errorf("missing %s: %v", dst, err)
			continue
		}
		if string(got) != want {
			t.Errorf("%s content = %q, want %q", dst, got, want)
		}
	}
}
