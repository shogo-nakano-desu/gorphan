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
	"gorphan/internal/pathutil"
)

type Options struct {
	Root       string
	ScanDir    string
	Files      []string
	Extensions []string
	MaxWorkers int
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

type pathIndex struct {
	idByPath map[string]int
	paths    []string
}

type edgeBuildResult struct {
	src      string
	targets  []string
	warnings []string
	err      error
}

type buildState struct {
	rootAbs    string
	scanDirAbs string
	extSet     map[string]struct{}
	inventory  map[string]struct{}
	sources    []string
	adj        map[string][]string
}

func Build(opts Options) (*Graph, error) {
	state, err := prepareBuildState(opts)
	if err != nil {
		return nil, err
	}

	results := runEdgeWorkers(state.sources, state.scanDirAbs, state.extSet, state.inventory, opts.Extensions, opts.MaxWorkers)
	warnings, err := applyEdgeResults(state.adj, results)
	if err != nil {
		return nil, err
	}

	return &Graph{
		Root:      state.rootAbs,
		Adjacency: state.adj,
		Warnings:  warnings,
	}, nil
}

func prepareBuildState(opts Options) (buildState, error) {
	if strings.TrimSpace(opts.Root) == "" {
		return buildState{}, fmt.Errorf("graph root is required")
	}
	if strings.TrimSpace(opts.ScanDir) == "" {
		return buildState{}, fmt.Errorf("graph scan dir is required")
	}

	rootAbs, err := pathutil.NormalizeAbs(opts.Root)
	if err != nil {
		return buildState{}, fmt.Errorf("resolve root: %w", err)
	}
	scanDirAbs, err := pathutil.NormalizeAbs(opts.ScanDir)
	if err != nil {
		return buildState{}, fmt.Errorf("resolve scan dir: %w", err)
	}

	extSet := pathutil.ExtensionSet(opts.Extensions)
	inventory, adj, err := buildInventory(opts.Files)
	if err != nil {
		return buildState{}, err
	}
	sources := sortedKeys(inventory)

	return buildState{
		rootAbs:    rootAbs,
		scanDirAbs: scanDirAbs,
		extSet:     extSet,
		inventory:  inventory,
		sources:    sources,
		adj:        adj,
	}, nil
}

func buildInventory(files []string) (map[string]struct{}, map[string][]string, error) {
	inventory := make(map[string]struct{}, len(files))
	adj := make(map[string][]string, len(files))
	for _, file := range files {
		abs, err := pathutil.NormalizeAbs(file)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve file path %q: %w", file, err)
		}
		inventory[abs] = struct{}{}
		adj[abs] = []string{}
	}
	return inventory, adj, nil
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func runEdgeWorkers(sources []string, scanDir string, extSet map[string]struct{}, inventory map[string]struct{}, extensions []string, maxWorkers int) <-chan edgeBuildResult {
	results := make(chan edgeBuildResult, len(sources))
	if len(sources) == 0 {
		close(results)
		return results
	}

	workerCount := resolveWorkerCount(len(sources), maxWorkers)
	jobs := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for src := range jobs {
				results <- buildEdgesForSource(src, scanDir, extSet, inventory, extensions)
			}
		}()
	}

	go func() {
		for _, src := range sources {
			jobs <- src
		}
		close(jobs)
		wg.Wait()
		close(results)
	}()

	return results
}

func resolveWorkerCount(sourceCount int, maxWorkers int) int {
	workerCount := runtime.GOMAXPROCS(0)
	if maxWorkers > 0 && workerCount > maxWorkers {
		workerCount = maxWorkers
	}
	if workerCount < 1 {
		workerCount = 1
	}
	if workerCount > sourceCount {
		workerCount = sourceCount
	}
	if workerCount < 1 {
		workerCount = 1
	}
	return workerCount
}

func applyEdgeResults(adj map[string][]string, results <-chan edgeBuildResult) ([]string, error) {
	warningSet := make(map[string]struct{})
	for res := range results {
		if res.err != nil {
			return nil, res.err
		}
		adj[res.src] = res.targets
		for _, warning := range res.warnings {
			warningSet[warning] = struct{}{}
		}
	}
	return toSortedSlice(warningSet), nil
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
		targetPath := filepath.Clean(filepath.Join(srcDir, filepath.FromSlash(link)))
		target, err := pathutil.NormalizeAbs(targetPath)
		if err != nil {
			return edgeBuildResult{src: src, err: fmt.Errorf("resolve linked path %q in %q: %w", link, src, err)}
		}

		if !pathutil.IsWithinDir(scanDir, target) {
			continue
		}
		if _, ok := extSet[strings.ToLower(filepath.Ext(target))]; !ok {
			continue
		}
		if _, ok := inventory[target]; !ok {
			warningSet[unresolvedWarning(src, target)] = struct{}{}
			continue
		}
		targetSet[target] = struct{}{}
	}

	return edgeBuildResult{
		src:      src,
		targets:  toSortedSlice(targetSet),
		warnings: toSortedSlice(warningSet),
	}
}

func unresolvedWarning(src, target string) string {
	return fmt.Sprintf("unresolved local markdown link: %s -> %s", src, target)
}

func Analyze(g *Graph, scanDir string, allFiles []string) (*Analysis, error) {
	if g == nil {
		return nil, fmt.Errorf("graph is required")
	}
	if strings.TrimSpace(scanDir) == "" {
		return nil, fmt.Errorf("scan dir is required")
	}

	scanDirAbs, err := pathutil.NormalizeAbs(scanDir)
	if err != nil {
		return nil, fmt.Errorf("resolve scan dir: %w", err)
	}

	index := pathIndex{
		idByPath: make(map[string]int, len(allFiles)+len(g.Adjacency)),
		paths:    make([]string, 0, len(allFiles)+len(g.Adjacency)),
	}
	inventory, inventoryIDs, err := indexInventoryFiles(&index, allFiles)
	if err != nil {
		return nil, err
	}
	rootID := index.intern(g.Root)
	if _, ok := inventory[rootID]; !ok {
		return nil, fmt.Errorf("root markdown file is not in scan result: %s", g.Root)
	}

	adjacencyIDs := indexAdjacency(&index, g.Adjacency)
	reachableIDs := traverseReachableIDs(rootID, adjacencyIDs)
	reachableSet, reachable := reachableFromIDs(&index, reachableIDs)
	orphans := findOrphans(&index, inventoryIDs, reachableIDs)
	orphansRelative, err := pathutil.RelativeSlashMany(scanDirAbs, orphans)
	if err != nil {
		return nil, fmt.Errorf("convert orphan path to relative: %w", err)
	}
	sort.Strings(orphansRelative)

	return &Analysis{
		Reachable:       reachable,
		Orphans:         orphans,
		OrphansRelative: orphansRelative,
		ReachableSet:    reachableSet,
	}, nil
}

func indexInventoryFiles(index *pathIndex, allFiles []string) (map[int]struct{}, []int, error) {
	inventory := make(map[int]struct{}, len(allFiles))
	inventoryIDs := make([]int, 0, len(allFiles))
	for _, file := range allFiles {
		abs, err := pathutil.NormalizeAbs(file)
		if err != nil {
			return nil, nil, fmt.Errorf("resolve file path %q: %w", file, err)
		}
		id := index.intern(abs)
		if _, ok := inventory[id]; ok {
			continue
		}
		inventory[id] = struct{}{}
		inventoryIDs = append(inventoryIDs, id)
	}
	return inventory, inventoryIDs, nil
}

func indexAdjacency(index *pathIndex, adjacency map[string][]string) map[int][]int {
	adjacencyIDs := make(map[int][]int, len(adjacency))
	for src, targets := range adjacency {
		srcID := index.intern(src)
		dstIDs := make([]int, 0, len(targets))
		for _, target := range targets {
			dstIDs = append(dstIDs, index.intern(target))
		}
		adjacencyIDs[srcID] = dstIDs
	}
	return adjacencyIDs
}

func reachableFromIDs(index *pathIndex, reachableIDs map[int]struct{}) (map[string]struct{}, []string) {
	reachableSet := make(map[string]struct{}, len(reachableIDs))
	for id := range reachableIDs {
		reachableSet[index.path(id)] = struct{}{}
	}
	reachable := toSortedSlice(reachableSet)
	return reachableSet, reachable
}

func findOrphans(index *pathIndex, inventoryIDs []int, reachableIDs map[int]struct{}) []string {
	orphans := make([]string, 0)
	for _, id := range inventoryIDs {
		if _, ok := reachableIDs[id]; ok {
			continue
		}
		orphans = append(orphans, index.path(id))
	}
	sort.Strings(orphans)
	return orphans
}

func traverseReachableIDs(root int, adjacency map[int][]int) map[int]struct{} {
	visited := make(map[int]struct{})
	stack := []int{root}

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

func (i *pathIndex) intern(path string) int {
	if id, ok := i.idByPath[path]; ok {
		return id
	}
	id := len(i.paths)
	i.idByPath[path] = id
	i.paths = append(i.paths, path)
	return id
}

func (i *pathIndex) path(id int) string {
	return i.paths[id]
}

func toSortedSlice(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for item := range set {
		out = append(out, item)
	}
	sort.Strings(out)
	return out
}

func ExportDOT(g *Graph, scanDir string) (string, error) {
	if g == nil {
		return "", fmt.Errorf("graph is required")
	}
	scanDirAbs, err := normalizedScanDir(scanDir)
	if err != nil {
		return "", err
	}
	nodes := make([]string, 0, len(g.Adjacency))
	for node := range g.Adjacency {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	lines := []string{"digraph gorphan {"}
	for _, src := range nodes {
		srcLabel, err := relativeLabel(scanDirAbs, src)
		if err != nil {
			return "", err
		}
		if len(g.Adjacency[src]) == 0 {
			lines = append(lines, fmt.Sprintf("  %q;", srcLabel))
			continue
		}
		for _, dst := range g.Adjacency[src] {
			dstLabel, err := relativeLabel(scanDirAbs, dst)
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
	scanDirAbs, err := normalizedScanDir(scanDir)
	if err != nil {
		return "", err
	}
	nodes := make([]string, 0, len(g.Adjacency))
	for node := range g.Adjacency {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	lines := []string{"graph TD"}
	for _, src := range nodes {
		srcLabel, err := relativeLabel(scanDirAbs, src)
		if err != nil {
			return "", err
		}
		if len(g.Adjacency[src]) == 0 {
			lines = append(lines, fmt.Sprintf("  %q", srcLabel))
			continue
		}
		for _, dst := range g.Adjacency[src] {
			dstLabel, err := relativeLabel(scanDirAbs, dst)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("  %q --> %q", srcLabel, dstLabel))
		}
	}
	return strings.Join(lines, "\n"), nil
}

func normalizedScanDir(scanDir string) (string, error) {
	if strings.TrimSpace(scanDir) == "" {
		return "", nil
	}
	scanDirAbs, err := pathutil.NormalizeAbs(scanDir)
	if err != nil {
		return "", fmt.Errorf("resolve scan dir: %w", err)
	}
	return scanDirAbs, nil
}

func relativeLabel(scanDirAbs, abs string) (string, error) {
	if scanDirAbs == "" {
		return filepath.ToSlash(abs), nil
	}
	rel, err := pathutil.RelativeSlash(scanDirAbs, abs)
	if err != nil {
		return "", fmt.Errorf("make relative label: %w", err)
	}
	return rel, nil
}
