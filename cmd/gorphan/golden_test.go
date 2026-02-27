package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorphan/internal/testutil"
)

func TestGolden_TextOutput(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	orphan := filepath.Join(docs, "orphan.md")
	testutil.MustWrite(t, root, "# root")
	testutil.MustWrite(t, orphan, "# orphan")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d; stderr=%s", code, stderr.String())
	}

	expected := mustReadGolden(t, "text_orphans.txt")
	got := strings.TrimSpace(stdout.String())
	if got != expected {
		t.Fatalf("golden mismatch\n--- expected ---\n%s\n--- got ---\n%s", expected, got)
	}
}

func TestGolden_JSONOutput(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	orphan := filepath.Join(docs, "orphan.md")
	testutil.MustWrite(t, root, "# root")
	testutil.MustWrite(t, orphan, "# orphan")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs, "--format", "json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d; stderr=%s", code, stderr.String())
	}

	got := strings.TrimSpace(stdout.String())
	got = strings.ReplaceAll(got, filepath.ToSlash(root), "<ROOT>")
	got = strings.ReplaceAll(got, filepath.ToSlash(docs), "<DIR>")
	got = strings.ReplaceAll(got, root, "<ROOT>")
	got = strings.ReplaceAll(got, docs, "<DIR>")

	expected := mustReadGolden(t, "json_orphans.json")
	if got != expected {
		t.Fatalf("golden mismatch\n--- expected ---\n%s\n--- got ---\n%s", expected, got)
	}
}

func mustReadGolden(t *testing.T, name string) string {
	t.Helper()
	path := filepath.Join("testdata", "golden", name)
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden file failed: %v", err)
	}
	return strings.TrimSpace(string(b))
}
