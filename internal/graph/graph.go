package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorphan/internal/parser"
)

type Options struct {
	Root       string
	ScanDir    string
	Files      []string
	Extensions []string
}

type Graph struct {
	Root      string
	Adjacency map[string][]string
}

func Build(opts Options) (*Graph, error) {
	if strings.TrimSpace(opts.Root) == "" {
		return nil, fmt.Errorf("graph root is required")
	}
	if strings.TrimSpace(opts.ScanDir) == "" {
		return nil, fmt.Errorf("graph scan dir is required")
	}

	rootAbs, err := filepath.Abs(opts.Root)
	if err != nil {
		return nil, fmt.Errorf("resolve root: %w", err)
	}
	rootAbs = filepath.Clean(rootAbs)

	scanDirAbs, err := filepath.Abs(opts.ScanDir)
	if err != nil {
		return nil, fmt.Errorf("resolve scan dir: %w", err)
	}
	scanDirAbs = filepath.Clean(scanDirAbs)

	extSet := buildExtSet(opts.Extensions)
	inventory := make(map[string]struct{}, len(opts.Files))
	adj := make(map[string][]string, len(opts.Files))

	for _, file := range opts.Files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, fmt.Errorf("resolve file path %q: %w", file, err)
		}
		abs = filepath.Clean(abs)
		inventory[abs] = struct{}{}
		adj[abs] = []string{}
	}

	for src := range inventory {
		content, err := os.ReadFile(src)
		if err != nil {
			return nil, fmt.Errorf("read markdown file %q: %w", src, err)
		}

		links := parser.ExtractLocalMarkdownLinks(string(content), opts.Extensions)
		targetSet := make(map[string]struct{})
		srcDir := filepath.Dir(src)

		for _, link := range links {
			target := filepath.Clean(filepath.Join(srcDir, filepath.FromSlash(link)))
			target, err = filepath.Abs(target)
			if err != nil {
				return nil, fmt.Errorf("resolve linked path %q in %q: %w", link, src, err)
			}

			if !isWithinDir(scanDirAbs, target) {
				continue
			}
			if _, ok := extSet[strings.ToLower(filepath.Ext(target))]; !ok {
				continue
			}
			if _, ok := inventory[target]; !ok {
				continue
			}
			targetSet[target] = struct{}{}
		}

		targets := make([]string, 0, len(targetSet))
		for t := range targetSet {
			targets = append(targets, t)
		}
		sort.Strings(targets)
		adj[src] = targets
	}

	return &Graph{
		Root:      rootAbs,
		Adjacency: adj,
	}, nil
}

func isWithinDir(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

func buildExtSet(extensions []string) map[string]struct{} {
	set := make(map[string]struct{}, len(extensions))
	for _, ext := range extensions {
		normalized := strings.ToLower(strings.TrimSpace(ext))
		if normalized == "" {
			continue
		}
		if !strings.HasPrefix(normalized, ".") {
			normalized = "." + normalized
		}
		set[normalized] = struct{}{}
	}
	if len(set) == 0 {
		set[".md"] = struct{}{}
		set[".markdown"] = struct{}{}
	}
	return set
}
