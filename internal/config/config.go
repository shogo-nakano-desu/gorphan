package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const DefaultConfigPath = ".gorphan.yaml"

type FileConfig struct {
	Root             string
	Dir              string
	Ext              string
	Ignore           []string
	IgnoreCheckFiles []string
	Format           string
	Verbose          *bool
	Unresolved       string
	Graph            string
}

type yamlToken struct {
	isList bool
	key    string
	value  string
}

type yamlParser struct {
	cfg         FileConfig
	currentList string
}

func FindConfigArg(args []string) (path string, explicit bool, err error) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--config" {
			if i+1 >= len(args) {
				return "", false, fmt.Errorf("--config requires a value")
			}
			return args[i+1], true, nil
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config="), true, nil
		}
	}
	return DefaultConfigPath, false, nil
}

func Load(path string, required bool) (FileConfig, bool, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return FileConfig{}, false, fmt.Errorf("resolve config path: %w", err)
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) && !required {
			return FileConfig{}, false, nil
		}
		if os.IsNotExist(err) && required {
			return FileConfig{}, false, fmt.Errorf("config file does not exist: %s", abs)
		}
		return FileConfig{}, false, fmt.Errorf("read config file: %w", err)
	}

	cfg, err := parseYAML(string(b))
	if err != nil {
		return FileConfig{}, false, fmt.Errorf("parse config file %s: %w", abs, err)
	}
	return cfg, true, nil
}

func parseYAML(content string) (FileConfig, error) {
	tokens, err := tokenizeYAML(content)
	if err != nil {
		return FileConfig{}, err
	}

	p := yamlParser{}
	for _, token := range tokens {
		if err := p.apply(token); err != nil {
			return FileConfig{}, err
		}
	}
	return p.cfg, nil
}

func tokenizeYAML(content string) ([]yamlToken, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	tokens := make([]yamlToken, 0)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "- ") {
			tokens = append(tokens, yamlToken{isList: true, value: strings.TrimSpace(strings.TrimPrefix(line, "- "))})
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		tokens = append(tokens, yamlToken{key: key, value: value})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (p *yamlParser) apply(token yamlToken) error {
	if token.isList {
		return p.applyListItem(token.value)
	}
	p.currentList = ""

	switch token.key {
	case "root":
		p.cfg.Root = token.value
	case "dir":
		p.cfg.Dir = token.value
	case "ext":
		p.cfg.Ext = token.value
	case "ignore":
		p.currentList = "ignore"
		if token.value != "" {
			p.cfg.Ignore = append(p.cfg.Ignore, token.value)
		}
	case "ignore-check-files":
		p.currentList = "ignore-check-files"
		if token.value != "" {
			p.cfg.IgnoreCheckFiles = append(p.cfg.IgnoreCheckFiles, token.value)
		}
	case "format":
		p.cfg.Format = token.value
	case "verbose":
		if token.value == "" {
			return nil
		}
		b, err := strconv.ParseBool(token.value)
		if err != nil {
			return fmt.Errorf("invalid verbose value: %s", token.value)
		}
		p.cfg.Verbose = &b
	case "unresolved":
		p.cfg.Unresolved = token.value
	case "graph":
		p.cfg.Graph = token.value
	}

	return nil
}

func (p *yamlParser) applyListItem(item string) error {
	switch p.currentList {
	case "ignore":
		p.cfg.Ignore = append(p.cfg.Ignore, item)
	case "ignore-check-files":
		p.cfg.IgnoreCheckFiles = append(p.cfg.IgnoreCheckFiles, item)
	default:
		// Keep backward-compatible behavior: list items outside known list contexts are ignored.
	}
	return nil
}
