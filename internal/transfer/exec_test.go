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

	src := filepath.Join(srcDir, "image (2).png")
	if err := os.WriteFile(src, []byte("test-content"), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := CopyCmd("myhost", []string{src}, dstDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("pipeline failed: %v\n%s", err, out)
	}

	got, err := os.ReadFile(filepath.Join(dstDir, "image (2).png"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "test-content" {
		t.Errorf("content = %q", got)
	}
}
