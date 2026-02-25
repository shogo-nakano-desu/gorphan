package parser

import (
	"html"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	inlineLinkRe = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)
	refDefRe     = regexp.MustCompile(`(?m)^\s{0,3}\[([^\]]+)\]:\s*(\S+)`)
	refLinkRe    = regexp.MustCompile(`\[[^\]]+\]\[([^\]]*)\]`)
)

func ExtractLocalMarkdownLinks(content string, extensions []string) []string {
	extSet := buildExtSet(extensions)
	refDefs := parseReferenceDefinitions(content)
	seen := make(map[string]struct{})
	links := make([]string, 0)

	for _, match := range inlineLinkRe.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		target, ok := normalizeMarkdownTarget(match[1], extSet)
		if !ok {
			continue
		}
		if _, exists := seen[target]; exists {
			continue
		}
		seen[target] = struct{}{}
		links = append(links, target)
	}

	for _, match := range refLinkRe.FindAllStringSubmatch(content, -1) {
		if len(match) < 2 {
			continue
		}
		label := strings.TrimSpace(match[1])
		if label == "" {
			continue
		}

		raw, ok := refDefs[strings.ToLower(label)]
		if !ok {
			continue
		}
		target, ok := normalizeMarkdownTarget(raw, extSet)
		if !ok {
			continue
		}
		if _, exists := seen[target]; exists {
			continue
		}
		seen[target] = struct{}{}
		links = append(links, target)
	}

	return links
}

func parseReferenceDefinitions(content string) map[string]string {
	matches := refDefRe.FindAllStringSubmatch(content, -1)
	defs := make(map[string]string, len(matches))
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		label := strings.ToLower(strings.TrimSpace(match[1]))
		dest := strings.TrimSpace(match[2])
		if label == "" || dest == "" {
			continue
		}
		defs[label] = dest
	}
	return defs
}

func normalizeMarkdownTarget(raw string, extSet map[string]struct{}) (string, bool) {
	target := parseDestination(raw)
	if target == "" {
		return "", false
	}
	target = decodeBasicEscapes(target)
	target = stripQueryAndFragment(target)
	target = strings.TrimSpace(target)
	if target == "" {
		return "", false
	}

	if isExternalTarget(target) {
		return "", false
	}
	if strings.HasPrefix(target, "/") {
		return "", false
	}

	target = filepath.ToSlash(filepath.Clean(target))
	ext := strings.ToLower(filepath.Ext(target))
	if ext == "" {
		return "", false
	}
	if _, ok := extSet[ext]; !ok {
		return "", false
	}
	return target, true
}

func parseDestination(raw string) string {
	dest := strings.TrimSpace(raw)
	if dest == "" {
		return ""
	}

	if strings.HasPrefix(dest, "<") {
		if end := strings.Index(dest, ">"); end > 1 {
			dest = strings.TrimSpace(dest[1:end])
		}
	}

	for i := 0; i < len(dest); i++ {
		ch := dest[i]
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			continue
		}

		if i > 0 && dest[i-1] == '\\' {
			continue
		}
		dest = dest[:i]
		break
	}

	return strings.TrimSpace(dest)
}

func decodeBasicEscapes(value string) string {
	decoded := strings.ReplaceAll(value, `\ `, " ")
	decoded = strings.ReplaceAll(decoded, `\(`, "(")
	decoded = strings.ReplaceAll(decoded, `\)`, ")")
	decoded = strings.ReplaceAll(decoded, `\[`, "[")
	decoded = strings.ReplaceAll(decoded, `\]`, "]")
	decoded = strings.ReplaceAll(decoded, `\#`, "#")
	decoded = html.UnescapeString(decoded)

	if unescaped, err := url.PathUnescape(decoded); err == nil {
		return unescaped
	}
	return decoded
}

func stripQueryAndFragment(value string) string {
	cut := len(value)
	if i := strings.Index(value, "#"); i >= 0 && i < cut {
		cut = i
	}
	if i := strings.Index(value, "?"); i >= 0 && i < cut {
		cut = i
	}
	return value[:cut]
}

func isExternalTarget(value string) bool {
	lower := strings.ToLower(value)
	return strings.HasPrefix(lower, "http://") ||
		strings.HasPrefix(lower, "https://") ||
		strings.HasPrefix(lower, "mailto:")
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
