package report

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Summary struct {
	Scanned   int `json:"scanned"`
	Reachable int `json:"reachable"`
	Orphans   int `json:"orphans"`
}

type Result struct {
	Root     string   `json:"root"`
	Dir      string   `json:"dir"`
	Orphans  []string `json:"orphans"`
	Warnings []string `json:"warnings,omitempty"`
	Graph    string   `json:"graph,omitempty"`
	Summary  Summary  `json:"summary"`
}

func RenderText(r Result, verbose bool, showWarnings bool, showGraph bool) string {
	lines := make([]string, 0)
	if len(r.Orphans) == 0 {
		lines = append(lines, "No orphan markdown files found.")
	} else {
		lines = append(lines, fmt.Sprintf("Orphan markdown files (%d):", len(r.Orphans)))
		for _, orphan := range r.Orphans {
			lines = append(lines, "- "+orphan)
		}
	}

	if verbose {
		lines = append(lines, "")
		lines = append(lines, "Summary:")
		lines = append(lines, fmt.Sprintf("- root: %s", r.Root))
		lines = append(lines, fmt.Sprintf("- dir: %s", r.Dir))
		lines = append(lines, fmt.Sprintf("- scanned: %d", r.Summary.Scanned))
		lines = append(lines, fmt.Sprintf("- reachable: %d", r.Summary.Reachable))
		lines = append(lines, fmt.Sprintf("- orphans: %d", r.Summary.Orphans))
		lines = append(lines, fmt.Sprintf("- warnings: %d", len(r.Warnings)))
	}

	if showWarnings && len(r.Warnings) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Unresolved local links (%d):", len(r.Warnings)))
		for _, warning := range r.Warnings {
			lines = append(lines, "- "+warning)
		}
	}

	if showGraph && strings.TrimSpace(r.Graph) != "" {
		lines = append(lines, "")
		lines = append(lines, "Graph:")
		lines = append(lines, r.Graph)
	}

	return strings.Join(lines, "\n")
}

func RenderJSON(r Result) (string, error) {
	out, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal report json: %w", err)
	}
	return string(out), nil
}
