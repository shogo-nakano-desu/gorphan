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
	var cfg FileConfig
	scanner := bufio.NewScanner(strings.NewReader(content))
	var currentList string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "- ") {
			if currentList == "ignore" {
				cfg.Ignore = append(cfg.Ignore, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			} else if currentList == "ignore-check-files" {
				cfg.IgnoreCheckFiles = append(cfg.IgnoreCheckFiles, strings.TrimSpace(strings.TrimPrefix(line, "- ")))
			}
			continue
		}

		currentList = ""
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		switch key {
		case "root":
			cfg.Root = value
		case "dir":
			cfg.Dir = value
		case "ext":
			cfg.Ext = value
		case "ignore":
			currentList = "ignore"
			if value != "" {
				cfg.Ignore = append(cfg.Ignore, value)
			}
		case "ignore-check-files":
			currentList = "ignore-check-files"
			if value != "" {
				cfg.IgnoreCheckFiles = append(cfg.IgnoreCheckFiles, value)
			}
		case "format":
			cfg.Format = value
		case "verbose":
			if value == "" {
				continue
			}
			b, err := strconv.ParseBool(value)
			if err != nil {
				return FileConfig{}, fmt.Errorf("invalid verbose value: %s", value)
			}
			cfg.Verbose = &b
		case "unresolved":
			cfg.Unresolved = value
		case "graph":
			cfg.Graph = value
		}
	}
	if err := scanner.Err(); err != nil {
		return FileConfig{}, err
	}
	return cfg, nil
}
