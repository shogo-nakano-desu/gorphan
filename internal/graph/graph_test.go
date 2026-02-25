package graph

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	if len(g.Warnings) != 1 {
		t.Fatalf("expected one unresolved warning, got: %#v", g.Warnings)
	}
}

func TestBuild_RequiresRootAndDir(t *testing.T) {
	_, err := Build(Options{})
	if err == nil {
		t.Fatalf("expected error for empty options")
	}
}

func TestAnalyze_ReachableAndOrphans(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	a := filepath.Join(dir, "a.md")
	b := filepath.Join(dir, "b.md")
	orphan := filepath.Join(dir, "orphan.md")
	mustWrite(t, root, "# root")
	mustWrite(t, a, "# a")
	mustWrite(t, b, "# b")
	mustWrite(t, orphan, "# orphan")

	g := &Graph{
		Root: root,
		Adjacency: map[string][]string{
			root:   {a},
			a:      {b},
			b:      {},
			orphan: {},
		},
	}

	analysis, err := Analyze(g, dir, []string{orphan, b, root, a})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}

	wantReachable := []string{a, b, root}
	if !reflect.DeepEqual(analysis.Reachable, wantReachable) {
		t.Fatalf("unexpected reachable files\nwant: %#v\n got: %#v", wantReachable, analysis.Reachable)
	}

	wantOrphans := []string{orphan}
	if !reflect.DeepEqual(analysis.Orphans, wantOrphans) {
		t.Fatalf("unexpected orphan files\nwant: %#v\n got: %#v", wantOrphans, analysis.Orphans)
	}

	wantRel := []string{"orphan.md"}
	if !reflect.DeepEqual(analysis.OrphansRelative, wantRel) {
		t.Fatalf("unexpected relative orphan files\nwant: %#v\n got: %#v", wantRel, analysis.OrphansRelative)
	}
}

func TestAnalyze_RootNotInInventory_ReturnsError(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	other := filepath.Join(dir, "other.md")
	mustWrite(t, root, "# root")
	mustWrite(t, other, "# other")

	g := &Graph{
		Root:      root,
		Adjacency: map[string][]string{root: {}},
	}

	_, err := Analyze(g, dir, []string{other})
	if err == nil || !strings.Contains(err.Error(), "root markdown file is not in scan result") {
		t.Fatalf("expected root inventory error, got: %v", err)
	}
}

func TestAnalyze_HandlesRootWithNoEdges(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	orphan := filepath.Join(dir, "orphan.md")
	mustWrite(t, root, "# root")
	mustWrite(t, orphan, "# orphan")

	g := &Graph{
		Root:      root,
		Adjacency: map[string][]string{root: {}},
	}

	analysis, err := Analyze(g, dir, []string{root, orphan})
	if err != nil {
		t.Fatalf("analyze failed: %v", err)
	}
	if len(analysis.Reachable) != 1 || analysis.Reachable[0] != root {
		t.Fatalf("unexpected reachable for root-only graph: %#v", analysis.Reachable)
	}
	if len(analysis.OrphansRelative) != 1 || analysis.OrphansRelative[0] != "orphan.md" {
		t.Fatalf("unexpected orphan relative list: %#v", analysis.OrphansRelative)
	}
}

func TestExportDOT(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	a := filepath.Join(dir, "a.md")
	g := &Graph{
		Root: root,
		Adjacency: map[string][]string{
			root: {a},
			a:    {},
		},
	}

	out, err := ExportDOT(g, dir)
	if err != nil {
		t.Fatalf("export dot failed: %v", err)
	}
	if !strings.Contains(out, `"index.md" -> "a.md";`) {
		t.Fatalf("unexpected dot output: %s", out)
	}
}

func TestExportMermaid(t *testing.T) {
	dir := t.TempDir()
	root := filepath.Join(dir, "index.md")
	a := filepath.Join(dir, "a.md")
	g := &Graph{
		Root: root,
		Adjacency: map[string][]string{
			root: {a},
			a:    {},
		},
	}

	out, err := ExportMermaid(g, dir)
	if err != nil {
		t.Fatalf("export mermaid failed: %v", err)
	}
	if !strings.Contains(out, `"index.md" --> "a.md"`) {
		t.Fatalf("unexpected mermaid output: %s", out)
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
