package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
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
	if !strings.Contains(stdout.String(), "reachable files: 2") {
		t.Fatalf("expected reachable files count, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "orphan files: 0") {
		t.Fatalf("expected orphan files count, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "No orphan markdown files found.") {
		t.Fatalf("expected no-orphan message, got: %s", stdout.String())
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

func TestRun_RootExcludedByIgnore_Fails(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	mustWrite(t, root, "# root")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs, "--ignore", "index.md"}, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("expected exit code 2, got %d", code)
	}
	if !strings.Contains(stderr.String(), "root markdown file is not in scan result") {
		t.Fatalf("expected root inventory error message, got: %s", stderr.String())
	}
}

func TestRun_OrphansReturnExitCode1(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	orphan := filepath.Join(docs, "orphan.md")
	mustWrite(t, root, "# root")
	mustWrite(t, orphan, "# orphan")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1, got %d", code)
	}
	if !strings.Contains(stdout.String(), "Orphan markdown files (1):") {
		t.Fatalf("expected orphan output, got: %s", stdout.String())
	}
	if !strings.Contains(stdout.String(), "- orphan.md") {
		t.Fatalf("expected orphan relative path, got: %s", stdout.String())
	}
}

func TestRun_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	orphan := filepath.Join(docs, "orphan.md")
	mustWrite(t, root, "# root")
	mustWrite(t, orphan, "# orphan")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs, "--format", "json"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("expected exit code 1 for orphan case, got %d", code)
	}
	out := stdout.String()
	if !strings.Contains(out, `"orphans": [`) {
		t.Fatalf("expected json orphans array, got: %s", out)
	}
	if !strings.Contains(out, `"orphan.md"`) {
		t.Fatalf("expected orphan path in json, got: %s", out)
	}
}

func TestRun_WarningIsNonFatal(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	mustWrite(t, root, "[missing](./missing.md)")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0 (no orphan files), got %d", code)
	}
	if !strings.Contains(stderr.String(), "warning: unresolved local markdown link:") {
		t.Fatalf("expected unresolved warning, got: %s", stderr.String())
	}
}

func TestRun_UnresolvedReportMode(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	mustWrite(t, root, "[missing](./missing.md)")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs, "--unresolved", "report"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if strings.Contains(stderr.String(), "warning:") {
		t.Fatalf("did not expect stderr warning in report mode, got: %s", stderr.String())
	}
	if !strings.Contains(stdout.String(), "Unresolved local links (1):") {
		t.Fatalf("expected unresolved report section, got: %s", stdout.String())
	}
}

func TestRun_GraphDotMode(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	child := filepath.Join(docs, "child.md")
	mustWrite(t, root, "[child](./child.md)")
	mustWrite(t, child, "# child")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	code := run([]string{"--root", root, "--dir", docs, "--graph", "dot"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected exit code 0, got %d; stderr=%s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "digraph gorphan {") {
		t.Fatalf("expected dot graph output, got: %s", stdout.String())
	}
}

func TestParseArgs_ConfigAndFlagOverride(t *testing.T) {
	dir := t.TempDir()
	docs := filepath.Join(dir, "docs")
	root := filepath.Join(docs, "index.md")
	mustWrite(t, root, "# root")
	cfgPath := filepath.Join(dir, ".gorphan.yaml")
	cfgContent := "root: docs/index.md\ndir: docs\nignore:\n  - drafts\nformat: json\nunresolved: report\ngraph: mermaid\n"
	if err := os.WriteFile(cfgPath, []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() { _ = os.Chdir(cwd) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	cfg, err := parseArgs([]string{"--config", cfgPath, "--ignore", "archive/*", "--format", "text", "--graph", "dot"}, &bytes.Buffer{})
	if err != nil {
		t.Fatalf("parseArgs failed: %v", err)
	}
	if cfg.Format != "text" {
		t.Fatalf("expected flag override for format, got: %s", cfg.Format)
	}
	if cfg.GraphFormat != "dot" {
		t.Fatalf("expected flag override for graph format, got: %s", cfg.GraphFormat)
	}
	if cfg.Unresolved != "report" {
		t.Fatalf("expected unresolved from config, got: %s", cfg.Unresolved)
	}
	if !reflect.DeepEqual(cfg.Ignore, []string{"drafts", "archive/*"}) {
		t.Fatalf("unexpected merged ignore list: %#v", cfg.Ignore)
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
