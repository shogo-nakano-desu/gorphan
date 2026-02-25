package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseArgsAndValidate_Success(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "docs", "index.md")
	mustWrite(t, root, "# root")

	cfg, err := parseArgs([]string{"--root", root, "--dir", filepath.Join(dir, "docs"), "--ignore", "drafts", "--format", "json"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}

	if cfg.Format != "json" {
		t.Fatalf("unexpected format: %s", cfg.Format)
	}
	if len(cfg.Ignore) != 1 || cfg.Ignore[0] != "drafts" {
		t.Fatalf("unexpected ignores: %#v", cfg.Ignore)
	}
	if !filepath.IsAbs(cfg.Root) || !filepath.IsAbs(cfg.Dir) {
		t.Fatalf("expected normalized absolute paths: root=%s dir=%s", cfg.Root, cfg.Dir)
	}
}

func TestParseArgsAndValidate_RootOutsideDir(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	scanDir := filepath.Join(dir, "docs")
	mustWrite(t, root, "# root")
	if err := os.MkdirAll(scanDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	_, err := parseArgs([]string{"--root", root, "--dir", scanDir}, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), "--root must be within --dir") {
		t.Fatalf("expected root-within-dir error, got: %v", err)
	}
}

func TestRun_ValidInput(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	child := filepath.Join(docs, "child.md")
	mustWrite(t, root, "[child](./child.md)")
	mustWrite(t, child, "# child")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs, "--verbose"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "graph nodes: 2") {
		t.Fatalf("expected verbose graph output, got: %s", stdout.String())
	}
}

func TestRun_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	mustWrite(t, root, "# root")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", dir, "--format", "xml"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "--format must be one of") {
		t.Fatalf("expected format error message, got: %s", stderr.String())
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write failed: %v", err)
	}
}
