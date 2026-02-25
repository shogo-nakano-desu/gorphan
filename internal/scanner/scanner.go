package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

type Options struct {
	Dir        string
	Extensions []string
	Ignore     []string
}

func NormalizeExtensions(raw string) []string {
	parts := strings.Split(raw, ",")
	seen := make(map[string]struct{}, len(parts))
	exts := make([]string, 0, len(parts))

	for _, part := range parts {
		ext := strings.ToLower(strings.TrimSpace(part))
		if ext == "" {
			continue
		}
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		if _, ok := seen[ext]; ok {
			continue
		}
		seen[ext] = struct{}{}
		exts = append(exts, ext)
	}

	if len(exts) == 0 {
		return []string{".md", ".markdown"}
	}
	return exts
}

func Scan(opts Options) ([]string, error) {
	if strings.TrimSpace(opts.Dir) == "" {
		return nil, fmt.Errorf("scan dir is required")
	}

	absDir, err := filepath.Abs(opts.Dir)
	if err != nil {
		return nil, fmt.Errorf("resolve scan dir: %w", err)
	}
	absDir = filepath.Clean(absDir)

	extSet := make(map[string]struct{}, len(opts.Extensions))
	for _, ext := range opts.Extensions {
		extSet[strings.ToLower(ext)] = struct{}{}
	}
	if len(extSet) == 0 {
		for _, ext := range NormalizeExtensions("") {
			extSet[ext] = struct{}{}
		}
	}

	files := make([]string, 0)
	err = filepath.WalkDir(absDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(absDir, path)
		if err != nil {
			return err
		}
		rel = filepath.Clean(rel)

		if rel != "." && isIgnored(rel, opts.Ignore) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := extSet[ext]; !ok {
			return nil
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		files = append(files, filepath.Clean(absPath))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan markdown files: %w", err)
	}

	sort.Strings(files)
	return files, nil
}

func isIgnored(rel string, rules []string) bool {
	if rel == "." || len(rules) == 0 {
		return false
	}

	relSlash := filepath.ToSlash(rel)
	for _, raw := range rules {
		rule := strings.TrimSpace(raw)
		if rule == "" {
			continue
		}
		rule = filepath.ToSlash(filepath.Clean(rule))

		if hasGlob(rule) {
			matched, err := filepath.Match(rule, relSlash)
			if err == nil && matched {
				return true
			}
			continue
		}

		rule = strings.TrimSuffix(rule, "/")
		if relSlash == rule || strings.HasPrefix(relSlash, rule+"/") {
			return true
		}
	}

	return false
}

func hasGlob(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}
