package pathutil

import (
	"fmt"
	"path/filepath"
	"strings"
)

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

func ExtensionSet(extensions []string) map[string]struct{} {
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
		for _, ext := range NormalizeExtensions("") {
			set[ext] = struct{}{}
		}
	}
	return set
}

func NormalizeAbs(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

func IsWithinDir(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return false
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return false
	}
	return true
}

func RelativeSlash(baseDir, path string) (string, error) {
	rel, err := filepath.Rel(baseDir, path)
	if err != nil {
		return "", err
	}
	return filepath.ToSlash(rel), nil
}

func RelativeSlashMany(baseDir string, paths []string) ([]string, error) {
	out := make([]string, 0, len(paths))
	for _, p := range paths {
		rel, err := RelativeSlash(baseDir, p)
		if err != nil {
			return nil, fmt.Errorf("convert path to relative: %w", err)
		}
		out = append(out, rel)
	}
	return out, nil
}
