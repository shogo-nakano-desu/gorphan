package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestIntegration_NoOrphans(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	a := filepath.Join(docs, "a.md")
	mustWrite(t, root, "[a](./a.md)")
	mustWrite(t, a, "# a")

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
	mustWrite(t, root, "# root")
	mustWrite(t, o1, "# orphan1")
	mustWrite(t, o2, "# orphan2")

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
	mustWrite(t, root, "[a](./a.md)")
	mustWrite(t, a, "[b](./b.md)")
	mustWrite(t, b, "[a](./a.md)")
	mustWrite(t, c, "[d](./d.md)")
	mustWrite(t, d, "# d")

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
