package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkBuildLargeGraph(b *testing.B) {
	root, files := createGraphFixture(b, 4000)
	opts := Options{
		Root:       root,
		ScanDir:    filepath.Dir(root),
		Files:      files,
		Extensions: []string{".md"},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g, err := Build(opts)
		if err != nil {
			b.Fatalf("build failed: %v", err)
		}
		if len(g.Adjacency) != len(files) {
			b.Fatalf("unexpected node count: got %d want %d", len(g.Adjacency), len(files))
		}
	}
}

func createGraphFixture(b *testing.B, count int) (string, []string) {
	b.Helper()
	dir := b.TempDir()
	files := make([]string, 0, count)

	for i := 0; i < count; i++ {
		path := filepath.Join(dir, fmt.Sprintf("doc-%05d.md", i))
		next := ""
		if i+1 < count {
			next = fmt.Sprintf("doc-%05d.md", i+1)
		}
		content := "# doc\n"
		if next != "" {
			content += fmt.Sprintf("[next](./%s)\n", next)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			b.Fatalf("write failed: %v", err)
		}
		files = append(files, path)
	}

	return files[0], files
}
