package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestBuild_GraphFiltersTargets(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	a := filepath.Join(dir, "a.md")
	b := filepath.Join(dir, "b.md")
	outside := filepath.Join(filepath.Dir(dir), "outside.md")

	mustWrite(t, root, "[a](./a.md) [outside](../outside.md) [missing](./missing.md) [text](./note.txt)")
	mustWrite(t, a, "[b](./b.md)")
	mustWrite(t, b, "[a](./a.md)")
	mustWrite(t, outside, "# outside")

	g, err := Build(Options{
		Root:       root,
		ScanDir:    dir,
		Files:      []string{root, a, b},
		Extensions: []string{".md"},
	})
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	want := map[string][]string{
		a:    {b},
		b:    {a},
		root: {a},
	}
	if !reflect.DeepEqual(g.Adjacency, want) {
		t.Fatalf("unexpected adjacency\nwant: %#v\n got: %#v", want, g.Adjacency)
	}

	if g.Root != root {
		t.Fatalf("unexpected root: %s", g.Root)
	}
}

func TestBuild_RequiresRootAndDir(t *testing.T) {
	_, err := Build(Options{})
	if err == nil {
		t.Fatalf("expected error for empty options")
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
