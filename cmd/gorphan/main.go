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
	ConfigPath       string
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	writef := func(w io.Writer, format string, a ...any) error {
		_, err := fmt.Fprintf(w, format, a...)
		return err
	}

	cfg, err := parseArgs(args, stderr)
	if err != nil {
		return 2
	}
	extensions := scanner.NormalizeExtensions(cfg.Ext)

	files, err := scanner.Scan(scanner.Options{
		Dir:        cfg.Dir,
		Extensions: extensions,
		Ignore:     cfg.Ignore,
	})
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
			return 2
		}
		return 2
	}

	linkGraph, err := graph.Build(graph.Options{
		Root:       cfg.Root,
		ScanDir:    cfg.Dir,
		Files:      files,
		Extensions: extensions,
	})
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
			return 2
		}
		return 2
	}
	analysis, err := graph.Analyze(linkGraph, cfg.Dir, files)
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
			return 2
		}
		return 2
	}
	analysis.Orphans, err = filterIgnoredCheckFiles(cfg.Dir, analysis.Orphans, cfg.IgnoreCheckFiles)
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
			return 2
		}
		return 2
	}
	sort.Strings(analysis.Orphans)
	analysis.OrphansRelative, err = toRelativeSlash(cfg.Dir, analysis.Orphans)
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
			return 2
		}
		return 2
	}
	graphText := ""
	switch cfg.GraphFormat {
	case "dot":
		graphText, err = graph.ExportDOT(linkGraph, cfg.Dir)
	case "mermaid":
		graphText, err = graph.ExportMermaid(linkGraph, cfg.Dir)
	}
	if err != nil {
		if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
			return 2
		}
		return 2
	}

	if cfg.Verbose {
		totalEdges := 0
		for _, targets := range linkGraph.Adjacency {
			totalEdges += len(targets)
		}
		if writeErr := writef(stdout, "Validated inputs:\n"); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- root: %s\n", cfg.Root); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- dir: %s\n", cfg.Dir); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- ext: %s\n", cfg.Ext); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- ignore: %v\n", cfg.Ignore); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- ignore-check-files: %v\n", cfg.IgnoreCheckFiles); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- format: %s\n", cfg.Format); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- unresolved: %s\n", cfg.Unresolved); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- graph: %s\n", cfg.GraphFormat); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- scanned markdown files: %d\n", len(files)); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- graph nodes: %d\n", len(linkGraph.Adjacency)); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- graph edges: %d\n", totalEdges); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- reachable files: %d\n", len(analysis.Reachable)); writeErr != nil {
			return 2
		}
		if writeErr := writef(stdout, "- orphan files: %d\n", len(analysis.Orphans)); writeErr != nil {
			return 2
		}
		if _, writeErr := fmt.Fprintln(stdout); writeErr != nil {
			return 2
		}
	}

	warnings := append([]string(nil), linkGraph.Warnings...)
	unresolvedFailed := false
	switch cfg.Unresolved {
	case "none":
		warnings = nil
	case "warn":
		for _, warning := range warnings {
			if _, writeErr := fmt.Fprintf(stderr, "warning: %s\n", warning); writeErr != nil {
				return 2
			}
		}
	case "fail":
		if len(warnings) > 0 {
			unresolvedFailed = true
		}
		for _, warning := range warnings {
			if _, writeErr := fmt.Fprintf(stderr, "error: %s\n", warning); writeErr != nil {
				return 2
			}
		}
	case "report":
		// warnings are emitted in standard report output.
	}

	rep := report.Result{
		Root:     cfg.Root,
		Dir:      cfg.Dir,
		Orphans:  analysis.OrphansRelative,
		Warnings: warnings,
		Graph:    graphText,
		Summary: report.Summary{
			Scanned:   len(files),
			Reachable: len(analysis.Reachable),
			Orphans:   len(analysis.Orphans),
		},
	}

	switch cfg.Format {
	case "json":
		rendered, err := report.RenderJSON(rep)
		if err != nil {
			if _, writeErr := fmt.Fprintf(stderr, "error: %v\n", err); writeErr != nil {
				return 2
			}
			return 2
		}
		if _, writeErr := fmt.Fprintln(stdout, rendered); writeErr != nil {
			return 2
		}
	default:
		if _, writeErr := fmt.Fprintln(stdout, report.RenderText(rep, cfg.Verbose, cfg.Unresolved == "report", cfg.GraphFormat != "none")); writeErr != nil {
			return 2
		}
	}

	if len(analysis.Orphans) > 0 || unresolvedFailed {
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

	dirAbs, err := filepath.Abs(cfg.Dir)
	if err != nil {
		return fmt.Errorf("resolve --dir: %w", err)
	}
	dirAbs = filepath.Clean(dirAbs)

	dirInfo, err := os.Stat(dirAbs)
	if err != nil {
		return fmt.Errorf("scan directory does not exist: %s", dirAbs)
	}
	if !dirInfo.IsDir() {
		return fmt.Errorf("--dir must be a directory: %s", dirAbs)
	}

	rootAbs, err := filepath.Abs(cfg.Root)
	if err != nil {
		return fmt.Errorf("resolve --root: %w", err)
	}
	rootAbs = filepath.Clean(rootAbs)

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

	scanDirAbs, err := filepath.Abs(scanDir)
	if err != nil {
		return nil, fmt.Errorf("resolve scan dir: %w", err)
	}
	scanDirAbs = filepath.Clean(scanDirAbs)

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
	fileAbs, err := filepath.Abs(file)
	if err != nil {
		return false, fmt.Errorf("resolve orphan file path %q: %w", file, err)
	}
	fileAbs = filepath.Clean(fileAbs)
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
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("resolve base directory: %w", err)
	}
	baseAbs = filepath.Clean(baseAbs)

	relative := make([]string, 0, len(files))
	for _, file := range files {
		abs, err := filepath.Abs(file)
		if err != nil {
			return nil, fmt.Errorf("resolve file path %q: %w", file, err)
		}
		abs = filepath.Clean(abs)
		rel, err := filepath.Rel(baseAbs, abs)
		if err != nil {
			return nil, fmt.Errorf("convert orphan path to relative: %w", err)
		}
		relative = append(relative, filepath.ToSlash(rel))
	}
	return relative, nil
}
