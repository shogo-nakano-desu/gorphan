package e2e

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"gorphan/internal/testutil"
)

func TestCLIEndToEnd_NoOrphans(t *testing.T) {
	repoRoot := mustRepoRoot(t)
	docs := filepath.Join(t.TempDir(), "docs")
	root := filepath.Join(docs, "index.md")
	child := filepath.Join(docs, "child.md")
	testutil.MustWrite(t, root, "[child](./child.md)")
	testutil.MustWrite(t, child, "# child")

	stdout, stderr, code := runCLI(t, repoRoot, "--root", root, "--dir", docs)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "No orphan markdown files found.") {
		t.Fatalf("unexpected output:\n%s", stdout)
	}
}

func TestCLIEndToEnd_WithOrphansJSON(t *testing.T) {
	repoRoot := mustRepoRoot(t)
	docs := filepath.Join(t.TempDir(), "docs")
	root := filepath.Join(docs, "index.md")
	orphan := filepath.Join(docs, "orphan.md")
	testutil.MustWrite(t, root, "# root")
	testutil.MustWrite(t, orphan, "# orphan")

	stdout, stderr, code := runCLI(t, repoRoot, "--root", root, "--dir", docs, "--format", "json")
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d\nstdout:\n%s\nstderr:\n%s", code, stdout, stderr)
	}

	var payload struct {
		Orphans []string `json:"orphans"`
	}
	if err := json.Unmarshal([]byte(stdout), &payload); err != nil {
		t.Fatalf("failed to parse json output: %v\noutput:\n%s", err, stdout)
	}
	if len(payload.Orphans) != 1 || payload.Orphans[0] != "orphan.md" {
		t.Fatalf("unexpected orphan payload: %#v", payload.Orphans)
	}
}

func runCLI(t *testing.T, repoRoot string, args ...string) (stdout string, stderr string, exitCode int) {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "gorphan")

	build := exec.Command("go", "build", "-o", bin, "./cmd/gorphan")
	build.Dir = repoRoot
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("failed to build cli: %v\noutput:\n%s", err, string(out))
	}

	cmd := exec.Command(bin, args...)
	cmd.Dir = repoRoot
	var outBuf bytes.Buffer
	var errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err == nil {
		return outBuf.String(), errBuf.String(), 0
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("failed to run cli: %v\nstdout:\n%s\nstderr:\n%s", err, outBuf.String(), errBuf.String())
	}
	return outBuf.String(), errBuf.String(), exitErr.ExitCode()
}

func mustRepoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("resolve repo root failed: %v", err)
	}
	return root
}
