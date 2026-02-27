package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkScanLargeTree(b *testing.B) {
	dir := createScanFixture(b, 120, 80)
	opts := Options{
		Dir:        dir,
		Extensions: []string{".md", ".markdown"},
		Ignore:     []string{"drafts", "assets/*"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		files, err := Scan(opts)
		if err != nil {
			b.Fatalf("scan failed: %v", err)
		}
		if len(files) == 0 {
			b.Fatal("expected scan results")
		}
	}
}

func createScanFixture(b *testing.B, dirCount, filesPerDir int) string {
	b.Helper()
	root := b.TempDir()

	for d := 0; d < dirCount; d++ {
		docDir := filepath.Join(root, fmt.Sprintf("section-%03d", d))
		if err := os.MkdirAll(docDir, 0o755); err != nil {
			b.Fatalf("mkdir failed: %v", err)
		}
		for f := 0; f < filesPerDir; f++ {
			ext := ".md"
			if f%7 == 0 {
				ext = ".markdown"
			}
			path := filepath.Join(docDir, fmt.Sprintf("doc-%03d%s", f, ext))
			if err := os.WriteFile(path, []byte("# doc\n"), 0o644); err != nil {
				b.Fatalf("write failed: %v", err)
			}
		}
		assetDir := filepath.Join(root, "assets", fmt.Sprintf("bundle-%03d", d))
		if err := os.MkdirAll(assetDir, 0o755); err != nil {
			b.Fatalf("mkdir assets failed: %v", err)
		}
		if err := os.WriteFile(filepath.Join(assetDir, "skip.md"), []byte("# skip\n"), 0o644); err != nil {
			b.Fatalf("write assets file failed: %v", err)
		}
	}

	if err := os.MkdirAll(filepath.Join(root, "drafts"), 0o755); err != nil {
		b.Fatalf("mkdir drafts failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "drafts", "ignored.md"), []byte("# ignored\n"), 0o644); err != nil {
		b.Fatalf("write drafts file failed: %v", err)
	}

	return root
}
