package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorphan/internal/testutil"
)

func TestIntegration_NoOrphans(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	a := filepath.Join(docs, "a.md")
	testutil.MustWrite(t, root, "[a](./a.md)")
	testutil.MustWrite(t, a, "# a")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No orphan markdown files found.") {
		t.Fatalf("unexpected output: %s", stdout.String())
	}
}

func TestIntegration_MultipleOrphans(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	o1 := filepath.Join(docs, "o1.md")
	o2 := filepath.Join(docs, "o2.md")
	testutil.MustWrite(t, root, "# root")
	testutil.MustWrite(t, o1, "# orphan1")
	testutil.MustWrite(t, o2, "# orphan2")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "- o1.md") || !strings.Contains(stdout.String(), "- o2.md") {
		t.Fatalf("expected both orphan files, got: %s", stdout.String())
	}
}

func TestIntegration_CyclicLinksWithDisconnectedComponent(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	a := filepath.Join(docs, "a.md")
	b := filepath.Join(docs, "b.md")
	c := filepath.Join(docs, "c.md")
	d := filepath.Join(docs, "d.md")
	testutil.MustWrite(t, root, "[a](./a.md)")
	testutil.MustWrite(t, a, "[b](./b.md)")
	testutil.MustWrite(t, b, "[a](./a.md)")
	testutil.MustWrite(t, c, "[d](./d.md)")
	testutil.MustWrite(t, d, "# d")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit 1, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "- c.md") || !strings.Contains(stdout.String(), "- d.md") {
		t.Fatalf("expected disconnected component files as orphans, got: %s", stdout.String())
	}
}

func TestIntegration_InvalidRootValidationFailure(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	if err := os.MkdirAll(docs, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", filepath.Join(docs, "missing.md"), "--dir", docs}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "root file does not exist") {
		t.Fatalf("expected root validation error, got: %s", stderr.String())
	}
}
