package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gorphan/internal/graph"
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
	Root    string
	Dir     string
	Ext     string
	Ignore  []string
	Format  string
	Verbose bool
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
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
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}

	linkGraph, err := graph.Build(graph.Options{
		Root:       cfg.Root,
		ScanDir:    cfg.Dir,
		Files:      files,
		Extensions: extensions,
	})
	if err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return 2
	}

	if cfg.Verbose {
		totalEdges := 0
		for _, targets := range linkGraph.Adjacency {
			totalEdges += len(targets)
		}
		fmt.Fprintf(stdout, "Validated inputs:\n")
		fmt.Fprintf(stdout, "- root: %s\n", cfg.Root)
		fmt.Fprintf(stdout, "- dir: %s\n", cfg.Dir)
		fmt.Fprintf(stdout, "- ext: %s\n", cfg.Ext)
		fmt.Fprintf(stdout, "- ignore: %v\n", cfg.Ignore)
		fmt.Fprintf(stdout, "- format: %s\n", cfg.Format)
		fmt.Fprintf(stdout, "- scanned markdown files: %d\n", len(files))
		fmt.Fprintf(stdout, "- graph nodes: %d\n", len(linkGraph.Adjacency))
		fmt.Fprintf(stdout, "- graph edges: %d\n", totalEdges)
	}

	// Phase 4 only: graph construction is complete. Reachability and orphan reporting are next.
	return 0
}

func parseArgs(args []string, stderr io.Writer) (config, error) {
	var cfg config
	var ignores multiFlag

	fs := flag.NewFlagSet("gorphan", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.StringVar(&cfg.Root, "root", "", "root markdown file (required)")
	fs.StringVar(&cfg.Dir, "dir", "", "directory to scan recursively (required)")
	fs.StringVar(&cfg.Ext, "ext", ".md,.markdown", "comma-separated markdown extensions")
	fs.Var(&ignores, "ignore", "ignore path prefix or glob (repeatable)")
	fs.StringVar(&cfg.Format, "format", "text", "output format: text or json")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "print validation diagnostics")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: gorphan --root <file.md> --dir <directory> [options]")
		fmt.Fprintln(stderr)
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	if fs.NArg() > 0 {
		return config{}, fmt.Errorf("unexpected arguments: %s", strings.Join(fs.Args(), " "))
	}

	cfg.Ignore = []string(ignores)
	if err := validateAndNormalize(&cfg); err != nil {
		fmt.Fprintf(stderr, "error: %v\n", err)
		return config{}, err
	}

	return cfg, nil
}

func validateAndNormalize(cfg *config) error {
	if strings.TrimSpace(cfg.Root) == "" {
		return fmt.Errorf("--root is required")
	}
	if strings.TrimSpace(cfg.Dir) == "" {
		return fmt.Errorf("--dir is required")
	}

	cfg.Format = strings.ToLower(strings.TrimSpace(cfg.Format))
	if cfg.Format != "text" && cfg.Format != "json" {
		return fmt.Errorf("--format must be one of: text, json")
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
