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
	Warnings  []string
}

type Analysis struct {
	Reachable       []string
	Orphans         []string
	OrphansRelative []string
	ReachableSet    map[string]struct{}
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
	warningSet := make(map[string]struct{})

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
				warningSet[fmt.Sprintf("unresolved local markdown link: %s -> %s", src, target)] = struct{}{}
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

	warnings := toSortedSlice(warningSet)

	return &Graph{
		Root:      rootAbs,
		Adjacency: adj,
		Warnings:  warnings,
	}, nil
}

func Analyze(g *Graph, scanDir string, allFiles []string) (*Analysis, error) {
	if g == nil {
		return nil, fmt.Errorf("graph is required")
	}
	if strings.TrimSpace(scanDir) == "" {
		return nil, fmt.Errorf("scan dir is required")
	}

	scanDirAbs, err := filepath.Abs(scanDir)
	if err != nil {
		return nil, fmt.Errorf("resolve scan dir: %w", err)
	}
	scanDirAbs = filepath.Clean(scanDirAbs)

	inventory := make(map[string]struct{}, len(allFiles))
	for _, file := range allFiles {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, fmt.Errorf("resolve file path %q: %w", file, err)
		}
		inventory[filepath.Clean(abs)] = struct{}{}
	}
	if _, ok := inventory[g.Root]; !ok {
		return nil, fmt.Errorf("root markdown file is not in scan result: %s", g.Root)
	}

	reachableSet := traverseReachable(g.Root, g.Adjacency)
	reachable := toSortedSlice(reachableSet)

	orphans := make([]string, 0)
	for _, file := range allFiles {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, fmt.Errorf("resolve file path %q: %w", file, err)
		}
		abs = filepath.Clean(abs)
		if _, ok := reachableSet[abs]; ok {
			continue
		}
		orphans = append(orphans, abs)
	}
	sort.Strings(orphans)

	orphansRelative, err := toRelativeSlash(scanDirAbs, orphans)
	if err != nil {
		return nil, err
	}

	return &Analysis{
		Reachable:       reachable,
		Orphans:         orphans,
		OrphansRelative: orphansRelative,
		ReachableSet:    reachableSet,
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

func traverseReachable(root string, adjacency map[string][]string) map[string]struct{} {
	visited := make(map[string]struct{})
	stack := []string{root}

	for len(stack) > 0 {
		n := len(stack) - 1
		node := stack[n]
		stack = stack[:n]
		if _, seen := visited[node]; seen {
			continue
		}
		visited[node] = struct{}{}

		neighbors := adjacency[node]
		for i := len(neighbors) - 1; i >= 0; i-- {
			next := neighbors[i]
			if _, seen := visited[next]; seen {
				continue
			}
			stack = append(stack, next)
		}
	}
	return visited
}

func toSortedSlice(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func toRelativeSlash(scanDir string, absPaths []string) ([]string, error) {
	out := make([]string, 0, len(absPaths))
	for _, p := range absPaths {
		rel, err := filepath.Rel(scanDir, p)
		if err != nil {
			return nil, fmt.Errorf("convert orphan path to relative: %w", err)
		}
		out = append(out, filepath.ToSlash(rel))
	}
	sort.Strings(out)
	return out, nil
}
