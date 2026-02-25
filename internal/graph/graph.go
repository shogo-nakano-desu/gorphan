package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

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

type edgeBuildResult struct {
	src      string
	targets  []string
	warnings []string
	err      error
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

	sources := make([]string, 0, len(inventory))
	for src := range inventory {
		sources = append(sources, src)
	}
	sort.Strings(sources)

	results := make(chan edgeBuildResult, len(sources))
	workerCount := runtime.GOMAXPROCS(0)
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > len(sources) && len(sources) > 0 {
		workerCount = len(sources)
	}

	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for src := range jobs {
				results <- buildEdgesForSource(src, scanDirAbs, extSet, inventory, opts.Extensions)
			}
		}()
	}

	for _, src := range sources {
		jobs <- src
	}
	close(jobs)
	wg.Wait()
	close(results)

	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		adj[res.src] = res.targets
		for _, warning := range res.warnings {
			warningSet[warning] = struct{}{}
		}
	}

	warnings := toSortedSlice(warningSet)

	return &Graph{
		Root:      rootAbs,
		Adjacency: adj,
		Warnings:  warnings,
	}, nil
}

func buildEdgesForSource(src, scanDir string, extSet map[string]struct{}, inventory map[string]struct{}, extensions []string) edgeBuildResult {
	content, err := os.ReadFile(src)
	if err != nil {
		return edgeBuildResult{src: src, err: fmt.Errorf("read markdown file %q: %w", src, err)}
	}

	links := parser.ExtractLocalMarkdownLinks(string(content), extensions)
	targetSet := make(map[string]struct{})
	warningSet := make(map[string]struct{})
	srcDir := filepath.Dir(src)

	for _, link := range links {
		target := filepath.Clean(filepath.Join(srcDir, filepath.FromSlash(link)))
		target, err = filepath.Abs(target)
		if err != nil {
			return edgeBuildResult{src: src, err: fmt.Errorf("resolve linked path %q in %q: %w", link, src, err)}
		}

		if !isWithinDir(scanDir, target) {
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

	targets := toSortedSlice(targetSet)
	warnings := toSortedSlice(warningSet)
	return edgeBuildResult{
		src:      src,
		targets:  targets,
		warnings: warnings,
	}
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

func ExportDOT(g *Graph, scanDir string) (string, error) {
	if g == nil {
		return "", fmt.Errorf("graph is required")
	}
	nodes := make([]string, 0, len(g.Adjacency))
	for node := range g.Adjacency {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	lines := []string{"digraph gorphan {"}
	for _, src := range nodes {
		srcLabel, err := relativeLabel(scanDir, src)
		if err != nil {
			return "", err
		}
		if len(g.Adjacency[src]) == 0 {
			lines = append(lines, fmt.Sprintf("  %q;", srcLabel))
			continue
		}
		for _, dst := range g.Adjacency[src] {
			dstLabel, err := relativeLabel(scanDir, dst)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("  %q -> %q;", srcLabel, dstLabel))
		}
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n"), nil
}

func ExportMermaid(g *Graph, scanDir string) (string, error) {
	if g == nil {
		return "", fmt.Errorf("graph is required")
	}
	nodes := make([]string, 0, len(g.Adjacency))
	for node := range g.Adjacency {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	lines := []string{"graph TD"}
	for _, src := range nodes {
		srcLabel, err := relativeLabel(scanDir, src)
		if err != nil {
			return "", err
		}
		if len(g.Adjacency[src]) == 0 {
			lines = append(lines, fmt.Sprintf("  %q", srcLabel))
			continue
		}
		for _, dst := range g.Adjacency[src] {
			dstLabel, err := relativeLabel(scanDir, dst)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("  %q --> %q", srcLabel, dstLabel))
		}
	}
	return strings.Join(lines, "\n"), nil
}

func relativeLabel(scanDir, abs string) (string, error) {
	if strings.TrimSpace(scanDir) == "" {
		return filepath.ToSlash(abs), nil
	}
	scanDirAbs, err := filepath.Abs(scanDir)
	if err != nil {
		return "", fmt.Errorf("resolve scan dir: %w", err)
	}
	rel, err := filepath.Rel(filepath.Clean(scanDirAbs), abs)
	if err != nil {
		return "", fmt.Errorf("make relative label: %w", err)
	}
	return filepath.ToSlash(rel), nil
}
