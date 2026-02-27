package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestFindConfigArg(t *testing.T) {
	path, explicit, err := FindConfigArg([]string{"--config", "custom.yaml"})
	if err != nil {
		t.Fatalf("find config arg failed: %v", err)
	}
	if !explicit || path != "custom.yaml" {
		t.Fatalf("unexpected config arg result: path=%s explicit=%v", path, explicit)
	}
}

func TestLoadAndParseYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gorphan.yaml")
	content := `
root: docs/index.md
dir: docs
ext: .md,.markdown
ignore:
  - drafts
  - archive/*
ignore-check-files:
  - docs/private.md
  - notes.md
format: json
verbose: true
unresolved: report
graph: mermaid
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	cfg, found, err := Load(path, true)
	if err != nil {
		t.Fatalf("load config failed: %v", err)
	}
	if !found {
		t.Fatalf("expected config to be found")
	}
	if cfg.Root != "docs/index.md" || cfg.Dir != "docs" || cfg.Format != "json" {
		t.Fatalf("unexpected scalar values: %#v", cfg)
	}
	if !reflect.DeepEqual(cfg.Ignore, []string{"drafts", "archive/*"}) {
		t.Fatalf("unexpected ignore list: %#v", cfg.Ignore)
	}
	if !reflect.DeepEqual(cfg.IgnoreCheckFiles, []string{"docs/private.md", "notes.md"}) {
		t.Fatalf("unexpected ignore-check-files list: %#v", cfg.IgnoreCheckFiles)
	}
	if cfg.Verbose == nil || *cfg.Verbose != true {
		t.Fatalf("expected verbose=true, got %#v", cfg.Verbose)
	}
}

func TestLoadMissingOptional(t *testing.T) {
	_, found, err := Load(filepath.Join(t.TempDir(), ".gorphan.yaml"), false)
	if err != nil {
		t.Fatalf("expected no error for missing optional config, got: %v", err)
	}
	if found {
		t.Fatalf("expected no config found")
	}
}
