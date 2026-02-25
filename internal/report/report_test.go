package report

import (
	"strings"
	"testing"
)

func TestRenderText_NoOrphans(t *testing.T) {
	r := Result{
		Root:    "/tmp/docs/index.md",
		Dir:     "/tmp/docs",
		Orphans: nil,
		Summary: Summary{
			Scanned:   2,
			Reachable: 2,
			Orphans:   0,
		},
	}

	out := RenderText(r, true, false, false)
	if !strings.Contains(out, "No orphan markdown files found.") {
		t.Fatalf("expected no orphan message, got: %s", out)
	}
	if !strings.Contains(out, "- scanned: 2") {
		t.Fatalf("expected summary scanned count, got: %s", out)
	}
}

func TestRenderText_WithOrphans(t *testing.T) {
	r := Result{
		Orphans: []string{"a.md", "b.md"},
		Summary: Summary{Orphans: 2},
	}
	out := RenderText(r, false, false, false)
	if !strings.Contains(out, "Orphan markdown files (2):") {
		t.Fatalf("expected orphan header, got: %s", out)
	}
	if !strings.Contains(out, "- a.md") || !strings.Contains(out, "- b.md") {
		t.Fatalf("expected orphan items, got: %s", out)
	}
}

func TestRenderText_WithWarningsAndGraph(t *testing.T) {
	r := Result{
		Warnings: []string{"unresolved local markdown link: a -> b"},
		Graph:    "graph TD\n  \"a\" --> \"b\"",
	}

	out := RenderText(r, false, true, true)
	if !strings.Contains(out, "Unresolved local links (1):") {
		t.Fatalf("expected warnings section, got: %s", out)
	}
	if !strings.Contains(out, "Graph:") || !strings.Contains(out, `"a" --> "b"`) {
		t.Fatalf("expected graph section, got: %s", out)
	}
}

func TestRenderJSON(t *testing.T) {
	r := Result{
		Root:    "/tmp/docs/index.md",
		Dir:     "/tmp/docs",
		Orphans: []string{"orphan.md"},
		Summary: Summary{Scanned: 2, Reachable: 1, Orphans: 1},
	}

	out, err := RenderJSON(r)
	if err != nil {
		t.Fatalf("render json failed: %v", err)
	}
	if !strings.Contains(out, `"orphan.md"`) {
		t.Fatalf("expected orphan in json output, got: %s", out)
	}
	if !strings.Contains(out, `"scanned": 2`) {
		t.Fatalf("expected summary in json output, got: %s", out)
	}
}
