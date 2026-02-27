package scanner

import (
	"path/filepath"
	"reflect"
	"testing"

	"gorphan/internal/testutil"
)

func TestNormalizeExtensions(t *testing.T) {
	got := NormalizeExtensions("md, .markdown, MD, ,")
	want := []string{".md", ".markdown"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected normalized extensions\nwant: %#v\n got: %#v", want, got)
	}
}

func TestScan_WithIgnoreRules(t *testing.T) {
	dir := t.TempDir()
	testutil.MustWrite(t, filepath.Join(dir, "index.md"), "# index")
	testutil.MustWrite(t, filepath.Join(dir, "guide", "intro.md"), "# intro")
	testutil.MustWrite(t, filepath.Join(dir, "guide", "notes.markdown"), "# notes")
	testutil.MustWrite(t, filepath.Join(dir, "guide", "skip.md"), "# skip")
	testutil.MustWrite(t, filepath.Join(dir, "assets", "readme.md"), "# assets")
	testutil.MustWrite(t, filepath.Join(dir, "README.txt"), "not markdown")

	files, err := Scan(Options{
		Dir:        dir,
		Extensions: []string{".md", ".markdown"},
		Ignore:     []string{"guide/skip.md", "assets/*"},
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	want := []string{
		filepath.Join(dir, "guide", "intro.md"),
		filepath.Join(dir, "guide", "notes.markdown"),
		filepath.Join(dir, "index.md"),
	}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("unexpected files\nwant: %#v\n got: %#v", want, files)
	}
}

func TestScan_IgnoreDirectoryPrefix(t *testing.T) {
	dir := t.TempDir()
	testutil.MustWrite(t, filepath.Join(dir, "docs", "index.md"), "# docs")
	testutil.MustWrite(t, filepath.Join(dir, "drafts", "a.md"), "# a")

	files, err := Scan(Options{
		Dir:        dir,
		Extensions: []string{".md"},
		Ignore:     []string{"drafts"},
	})
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	want := []string{filepath.Join(dir, "docs", "index.md")}
	if !reflect.DeepEqual(files, want) {
		t.Fatalf("unexpected files\nwant: %#v\n got: %#v", want, files)
	}
}
