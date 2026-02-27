package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	configpkg "gorphan/internal/config"
	"gorphan/internal/graph"
	"gorphan/internal/pathutil"
	"gorphan/internal/report"
	"gorphan/internal/scanner"
)

type multiFlag []string

func (m *multiFlag) String() string {
	return strings.Join(*m, ",")
}

func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}

type config struct {
	Root             string
	Dir              string
	Ext              string
	Ignore           []string
	IgnoreCheckFiles []string
	Format           string
	Verbose          bool
	Unresolved       string
	GraphFormat      string
	Workers          int
	MaxGraphNodes    int
	ConfigPath       string
}

type runState struct {
	cfg              config
	extensions       []string
	files            []string
	linkGraph        *graph.Graph
	analysis         *graph.Analysis
	warnings         []string
	graphText        string
	unresolvedFailed bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	cfg, err := parseArgs(args, stderr)
	if err != nil {
		return 2
	}

	state := &runState{cfg: cfg, extensions: scanner.NormalizeExtensions(cfg.Ext)}
	if err := state.scanFiles(); err != nil {
		return writeRunError(stderr, err)
	}
	if err := state.buildAndAnalyzeGraph(); err != nil {
		return writeRunError(stderr, err)
	}
	if err := state.postProcessOrphans(); err != nil {
		return writeRunError(stderr, err)
	}
	if err := state.prepareGraphText(stderr); err != nil {
		return writeRunError(stderr, err)
	}
	if err := state.renderVerbose(stdout); err != nil {
		return 2
	}
	if err := state.applyWarningPolicy(stderr); err != nil {
		return 2
	}
	if err := state.renderReport(stdout); err != nil {
		return 2
	}
	return state.exitCode()
}

func writeRunError(stderr io.Writer, err error) int {
	if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
		return 2
	}
	return 2
}

func (s *runState) scanFiles() error {
	files, err := scanner.Scan(scanner.Options{
		Dir:        s.cfg.Dir,
		Extensions: s.extensions,
		Ignore:     s.cfg.Ignore,
	})
	if err != nil {
		return err
	}
	s.files = files
	return nil
}

func (s *runState) buildAndAnalyzeGraph() error {
	linkGraph, err := graph.Build(graph.Options{
		Root:       s.cfg.Root,
		ScanDir:    s.cfg.Dir,
		Files:      s.files,
		Extensions: s.extensions,
		MaxWorkers: s.cfg.Workers,
	})
	if err != nil {
		return err
	}
	analysis, err := graph.Analyze(linkGraph, s.cfg.Dir, s.files)
	if err != nil {
		return err
	}
	s.linkGraph = linkGraph
	s.analysis = analysis
	return nil
}

func (s *runState) postProcessOrphans() error {
	orphans, err := filterIgnoredCheckFiles(s.cfg.Dir, s.analysis.Orphans, s.cfg.IgnoreCheckFiles)
	if err != nil {
		return err
	}
	sort.Strings(orphans)
	relative, err := toRelativeSlash(s.cfg.Dir, orphans)
	if err != nil {
		return err
	}
	s.analysis.Orphans = orphans
	s.analysis.OrphansRelative = relative
	return nil
}

func (s *runState) prepareGraphText(stderr io.Writer) error {
	s.graphText = ""
	graphNodeCount := len(s.linkGraph.Adjacency)
	graphLimited := s.cfg.GraphFormat != "none" && s.cfg.MaxGraphNodes > 0 && graphNodeCount > s.cfg.MaxGraphNodes
	if graphLimited {
		_, err := fmt.Fprintf(stderr, "warning: graph export skipped: node count %d exceeds --max-graph-nodes=%d\n", graphNodeCount, s.cfg.MaxGraphNodes)
		return err
	}

	var err error
	switch s.cfg.GraphFormat {
	case "dot":
		s.graphText, err = graph.ExportDOT(s.linkGraph, s.cfg.Dir)
	case "mermaid":
		s.graphText, err = graph.ExportMermaid(s.linkGraph, s.cfg.Dir)
	}
	return err
}

func (s *runState) renderVerbose(stdout io.Writer) error {
	if !s.cfg.Verbose {
		return nil
	}

	totalEdges := 0
	for _, targets := range s.linkGraph.Adjacency {
		totalEdges += len(targets)
	}

	lines := []string{
		"Validated inputs:",
		fmt.Sprintf("- root: %s", s.cfg.Root),
		fmt.Sprintf("- dir: %s", s.cfg.Dir),
		fmt.Sprintf("- ext: %s", s.cfg.Ext),
		fmt.Sprintf("- ignore: %v", s.cfg.Ignore),
		fmt.Sprintf("- ignore-check-files: %v", s.cfg.IgnoreCheckFiles),
		fmt.Sprintf("- format: %s", s.cfg.Format),
		fmt.Sprintf("- unresolved: %s", s.cfg.Unresolved),
		fmt.Sprintf("- graph: %s", s.cfg.GraphFormat),
		fmt.Sprintf("- max-graph-nodes: %d", s.cfg.MaxGraphNodes),
		fmt.Sprintf("- workers: %d", s.cfg.Workers),
		fmt.Sprintf("- scanned markdown files: %d", len(s.files)),
		fmt.Sprintf("- graph nodes: %d", len(s.linkGraph.Adjacency)),
		fmt.Sprintf("- graph edges: %d", totalEdges),
		fmt.Sprintf("- reachable files: %d", len(s.analysis.Reachable)),
		fmt.Sprintf("- orphan files: %d", len(s.analysis.Orphans)),
		"",
	}

	_, err := fmt.Fprintln(stdout, strings.Join(lines, "\n"))
	return err
}

func (s *runState) applyWarningPolicy(stderr io.Writer) error {
	s.warnings = append([]string(nil), s.linkGraph.Warnings...)
	s.unresolvedFailed = false

	switch s.cfg.Unresolved {
	case "none":
		s.warnings = nil
	case "warn":
		for _, warning := range s.warnings {
			if _, err := fmt.Fprintf(stderr, "warning: %s\n", warning); err != nil {
				return err
			}
		}
	case "fail":
		if len(s.warnings) > 0 {
			s.unresolvedFailed = true
		}
		for _, warning := range s.warnings {
			if _, err := fmt.Fprintf(stderr, "error: %s\n", warning); err != nil {
				return err
			}
		}
	case "report":
		// warnings are emitted in standard report output.
	}

	return nil
}

func (s *runState) renderReport(stdout io.Writer) error {
	rep := report.Result{
		Root:     s.cfg.Root,
		Dir:      s.cfg.Dir,
		Orphans:  s.analysis.OrphansRelative,
		Warnings: s.warnings,
		Graph:    s.graphText,
		Summary: report.Summary{
			Scanned:   len(s.files),
			Reachable: len(s.analysis.Reachable),
			Orphans:   len(s.analysis.Orphans),
		},
	}

	switch s.cfg.Format {
	case "json":
		rendered, err := report.RenderJSON(rep)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(stdout, rendered)
		return err
	default:
		_, err := fmt.Fprintln(stdout, report.RenderText(rep, s.cfg.Verbose, s.cfg.Unresolved == "report", s.cfg.GraphFormat != "none"))
		return err
	}
}

func (s *runState) exitCode() int {
	if len(s.analysis.Orphans) > 0 || s.unresolvedFailed {
		return 1
	}
	return 0
}

func parseArgs(args []string, stderr io.Writer) (config, error) {
	var cfg config
	var ignores multiFlag
	var ignoreCheckFiles multiFlag
	cfgPath, cfgExplicit, err := configpkg.FindConfigArg(args)
	if err != nil {
		return config{}, err
	}
	fileCfg, _, err := configpkg.Load(cfgPath, cfgExplicit)
	if err != nil {
		return config{}, err
	}

	cfg = config{
		Root:             fileCfg.Root,
		Dir:              fileCfg.Dir,
		Ext:              fileCfg.Ext,
		Ignore:           append([]string(nil), fileCfg.Ignore...),
		IgnoreCheckFiles: append([]string(nil), fileCfg.IgnoreCheckFiles...),
		Format:           fileCfg.Format,
		Unresolved:       fileCfg.Unresolved,
		GraphFormat:      fileCfg.Graph,
		ConfigPath:       cfgPath,
	}
	if fileCfg.Verbose != nil {
		cfg.Verbose = *fileCfg.Verbose
	}
	if cfg.Ext == "" {
		cfg.Ext = ".md,.markdown"
	}
	if cfg.Format == "" {
		cfg.Format = "text"
	}
	if cfg.Unresolved == "" {
		cfg.Unresolved = "fail"
	}
	if cfg.GraphFormat == "" {
		cfg.GraphFormat = "none"
	}
	if cfg.Workers < 0 {
		cfg.Workers = 0
	}
	if cfg.MaxGraphNodes < 0 {
		cfg.MaxGraphNodes = 0
	}

	fs := flag.NewFlagSet("gorphan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&cfg.Root, "root", cfg.Root, "root markdown file (required)")
	fs.StringVar(&cfg.Dir, "dir", cfg.Dir, "directory to scan recursively (default: current directory)")
	fs.StringVar(&cfg.Ext, "ext", cfg.Ext, "comma-separated markdown extensions")
	fs.Var(&ignores, "ignore", "ignore path prefix or glob (repeatable)")
	fs.Var(&ignoreCheckFiles, "ignore-check-file", "ignore orphan check for file by relative path or basename (repeatable)")
	fs.StringVar(&cfg.Format, "format", cfg.Format, "output format: text or json")
	fs.BoolVar(&cfg.Verbose, "verbose", cfg.Verbose, "print validation diagnostics")
	fs.StringVar(&cfg.Unresolved, "unresolved", cfg.Unresolved, "unresolved-link mode: fail, warn, report, none")
	fs.StringVar(&cfg.GraphFormat, "graph", cfg.GraphFormat, "graph export mode: none, dot, mermaid")
	fs.IntVar(&cfg.Workers, "workers", cfg.Workers, "max concurrent graph build workers (0 uses GOMAXPROCS)")
	fs.IntVar(&cfg.MaxGraphNodes, "max-graph-nodes", cfg.MaxGraphNodes, "max nodes to render for graph export (0 disables limit)")
	fs.StringVar(&cfg.ConfigPath, "config", cfg.ConfigPath, "optional config file path")
	fs.Usage = func() {
		_, _ = fmt.Fprintln(stderr, "Usage: gorphan --root <file.md> [--dir <directory>] [options]")
		_, _ = fmt.Fprintln(stderr)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if fs.NArg() > 0 {
		return config{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	cfg.Ignore = append(cfg.Ignore, []string(ignores)...)
	cfg.IgnoreCheckFiles = append(cfg.IgnoreCheckFiles, []string(ignoreCheckFiles)...)
	if err := validateAndNormalize(&cfg); err != nil {
		_, _ = fmt.Fprintf(stderr, "error: %v\n", err)
		return config{}, err
	}

	return cfg, nil
}

func validateAndNormalize(cfg *config) error {
	if strings.TrimSpace(cfg.Root) == "" {
		return fmt.Errorf("--root is required")
	}
	if strings.TrimSpace(cfg.Dir) == "" {
		cfg.Dir = "."
	}

	cfg.Format = strings.ToLower(strings.TrimSpace(cfg.Format))
	if cfg.Format != "text" && cfg.Format != "json" {
		return fmt.Errorf("--format must be one of: text, json")
	}
	cfg.Unresolved = strings.ToLower(strings.TrimSpace(cfg.Unresolved))
	if cfg.Unresolved != "fail" && cfg.Unresolved != "warn" && cfg.Unresolved != "report" && cfg.Unresolved != "none" {
		return fmt.Errorf("--unresolved must be one of: fail, warn, report, none")
	}
	cfg.GraphFormat = strings.ToLower(strings.TrimSpace(cfg.GraphFormat))
	if cfg.GraphFormat != "none" && cfg.GraphFormat != "dot" && cfg.GraphFormat != "mermaid" {
		return fmt.Errorf("--graph must be one of: none, dot, mermaid")
	}
	if cfg.Workers < 0 {
		return fmt.Errorf("--workers must be >= 0")
	}
	if cfg.MaxGraphNodes < 0 {
		return fmt.Errorf("--max-graph-nodes must be >= 0")
	}

	dirAbs, err := pathutil.NormalizeAbs(cfg.Dir)
	if err != nil {
		return fmt.Errorf("resolve --dir: %w", err)
	}

	dirInfo, err := os.Stat(dirAbs)
	if err != nil {
		return fmt.Errorf("scan directory does not exist: %s", dirAbs)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("--dir must be a directory: %s", dirAbs)
	}

	rootAbs, err := pathutil.NormalizeAbs(cfg.Root)
	if err != nil {
		return fmt.Errorf("resolve --root: %w", err)
	}

	rootInfo, err := os.Stat(rootAbs)
	if err != nil {
		return fmt.Errorf("root file does not exist: %s", rootAbs)
	}
	if rootInfo.IsDir() {
		return fmt.Errorf("--root must be a file: %s", rootAbs)
	}

	rel, err := filepath.Rel(dirAbs, rootAbs)
	if err != nil {
		return fmt.Errorf("verify root location: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("--root must be within --dir: root=%s dir=%s", rootAbs, dirAbs)
	}

	cfg.Dir = dirAbs
	cfg.Root = rootAbs
	return nil
}

func filterIgnoredCheckFiles(scanDir string, orphanFiles []string, rules []string) ([]string, error) {
	if len(orphanFiles) == 0 || len(rules) == 0 {
		return orphanFiles, nil
	}

	scanDirAbs, err := pathutil.NormalizeAbs(scanDir)
	if err != nil {
		return nil, fmt.Errorf("resolve scan dir: %w", err)
	}

	filtered := make([]string, 0, len(orphanFiles))
	for _, orphan := range orphanFiles {
		ignored, err := isIgnoredCheckFile(scanDirAbs, orphan, rules)
		if err != nil {
			return nil, err
		}
		if ignored {
			continue
		}
		filtered = append(filtered, orphan)
	}
	return filtered, nil
}

func isIgnoredCheckFile(scanDir, file string, rules []string) (bool, error) {
	fileAbs, err := pathutil.NormalizeAbs(file)
	if err != nil {
		return false, fmt.Errorf("resolve orphan file path %q: %w", file, err)
	}
	rel, err := filepath.Rel(scanDir, fileAbs)
	if err != nil {
		return false, fmt.Errorf("resolve orphan file relative path %q: %w", fileAbs, err)
	}
	rel = filepath.ToSlash(filepath.Clean(rel))
	base := filepath.Base(fileAbs)

	for _, rawRule := range rules {
		ruleRaw := strings.TrimSpace(rawRule)
		if ruleRaw == "" {
			continue
		}
		rule := normalizeIgnoreCheckRule(scanDir, ruleRaw)
		if strings.Contains(rule, "/") {
			if rel == rule {
				return true, nil
			}
			continue
		}
		if base == rule {
			return true, nil
		}
	}
	return false, nil
}

func normalizeIgnoreCheckRule(scanDir, rule string) string {
	if filepath.IsAbs(rule) {
		absRule := filepath.Clean(rule)
		if rel, err := filepath.Rel(scanDir, absRule); err == nil {
			rel = filepath.Clean(rel)
			if rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
				rule = rel
			}
		}
	}
	return filepath.ToSlash(filepath.Clean(rule))
}

func toRelativeSlash(baseDir string, files []string) ([]string, error) {
	baseAbs, err := pathutil.NormalizeAbs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolve base directory: %w", err)
	}

	absolute := make([]string, 0, len(files))
	for _, file := range files {
		abs, err := pathutil.NormalizeAbs(file)
		if err != nil {
			return nil, fmt.Errorf("resolve file path %q: %w", file, err)
		}
		absolute = append(absolute, abs)
	}

	relative, err := pathutil.RelativeSlashMany(baseAbs, absolute)
	if err != nil {
		return nil, fmt.Errorf("convert orphan path to relative: %w", err)
	}
	return relative, nil
}
