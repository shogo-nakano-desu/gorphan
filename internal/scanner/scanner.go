package scanner

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"gorphan/internal/pathutil"
)

type Options struct {
	Dir        string
	Extensions []string
	Ignore     []string
}

type ignoreMatcher struct {
	globs    []string
	prefixes []string
}

func NormalizeExtensions(raw string) []string {
	return pathutil.NormalizeExtensions(raw)
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

	extSet := pathutil.ExtensionSet(opts.Extensions)
	ignore := compileIgnoreRules(opts.Ignore)

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

		if rel != "." && ignore.matches(rel) {
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

		files = append(files, filepath.Clean(path))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("scan markdown files: %w", err)
	}

	sort.Strings(files)
	return files, nil
}

func compileIgnoreRules(rules []string) ignoreMatcher {
	matcher := ignoreMatcher{}
	if len(rules) == 0 {
		return matcher
	}

	matcher.globs = make([]string, 0, len(rules))
	matcher.prefixes = make([]string, 0, len(rules))
	for _, raw := range rules {
		rule := strings.TrimSpace(raw)
		if rule == "" {
			continue
		}
		rule = filepath.ToSlash(filepath.Clean(rule))
		if rule == "." {
			continue
		}
		if hasGlob(rule) {
			matcher.globs = append(matcher.globs, rule)
			continue
		}
		matcher.prefixes = append(matcher.prefixes, strings.TrimSuffix(rule, "/"))
	}
	return matcher
}

func (m ignoreMatcher) matches(rel string) bool {
	if rel == "." || (len(m.globs) == 0 && len(m.prefixes) == 0) {
		return false
	}

	relSlash := filepath.ToSlash(rel)
	for _, rule := range m.globs {
		matched, err := filepath.Match(rule, relSlash)
		if err == nil && matched {
			return true
		}
	}
	for _, rule := range m.prefixes {
		if relSlash == rule || strings.HasPrefix(relSlash, rule+"/") {
			return true
		}
	}

	return false
}

func hasGlob(pattern string) bool {
	return strings.ContainsAny(pattern, "*?[")
}
