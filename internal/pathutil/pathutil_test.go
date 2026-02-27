package pathutil

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestNormalizeExtensions(t *testing.T) {
	got := NormalizeExtensions("md, .markdown, MD, ,")
	want := []string{".md", ".markdown"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized extensions\nwant: %#v\n got: %#v", want, got)
	}
}

func TestExtensionSet_Defaults(t *testing.T) {
	set := ExtensionSet(nil)
	if _, ok := set[".md"]; !ok {
		t.Fatal("expected .md in default extension set")
	}
	if _, ok := set[".markdown"]; !ok {
		t.Fatal("expected .markdown in default extension set")
	}
}

func TestRelativeSlashMany(t *testing.T) {
	base := t.TempDir()
	paths := []string{
		filepath.Join(base, "a.md"),
		filepath.Join(base, "nested", "b.md"),
	}
	got, err := RelativeSlashMany(base, paths)
	if err != nil {
		t.Fatalf("relative conversion failed: %v", err)
	}
	want := []string{"a.md", "nested/b.md"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected relative paths\nwant: %#v\n got: %#v", want, got)
	}
}

func TestIsWithinDir(t *testing.T) {
	base := t.TempDir()
	inside := filepath.Join(base, "a.md")
	outside := filepath.Join(filepath.Dir(base), "a.md")
	if !IsWithinDir(base, inside) {
		t.Fatal("expected inside path to be within dir")
	}
	if IsWithinDir(base, outside) {
		t.Fatal("expected outside path to not be within dir")
	}
}
