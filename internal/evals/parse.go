package evals

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Base    string   `toml:"base"`
	BaseDir string   `toml:"base_dir"`
	Intent  string   `toml:"intent"`
	Graders []Grader `toml:"graders"`
}

func ParseTest(raw string) (Config, string, error) {
	cfg, body, err := parseFrontmatter(raw)
	if err != nil {
		return Config{}, "", err
	}
	return cfg, body, nil
}

func parseFrontmatter(raw string) (Config, string, error) {
	var cfg Config

	trimmed := strings.TrimLeft(raw, "\ufeff \t\r\n")
	if !strings.HasPrefix(trimmed, "+++") {
		return cfg, trimmed, fmt.Errorf("test.md sem frontmatter TOML (deve começar com +++)")
	}

	newline := strings.IndexByte(trimmed, '\n')
	if newline == -1 {
		return cfg, "", fmt.Errorf("frontmatter não terminado")
	}

	rest := trimmed[newline+1:]
	end := strings.Index(rest, "\n+++")
	if end == -1 {
		return cfg, "", fmt.Errorf("frontmatter não fechado (falta +++)")
	}

	front := rest[:end]
	body := strings.TrimSpace(rest[end+4:])

	if _, err := toml.Decode(front, &cfg); err != nil {
		return cfg, "", fmt.Errorf("erro ao parsear TOML: %w", err)
	}

	if cfg.Base == "" && cfg.BaseDir == "" {
		return cfg, "", fmt.Errorf("test.md sem base ou base_dir definida")
	}

	return cfg, body, nil
}

func LoadCase(dir, name string) (string, *Case, error) {
	testFile := filepath.Join(dir, name, "test.md")
	raw, err := os.ReadFile(testFile)
	if err != nil {
		return "", nil, fmt.Errorf("erro ao ler test.md: %w", err)
	}

	cfg, body, err := ParseTest(string(raw))
	if err != nil {
		return name, nil, err
	}

	return name, &Case{
		Base:    cfg.Base,
		BaseDir: cfg.BaseDir,
		Intent:  cfg.Intent,
		Graders: cfg.Graders,
		Context: body,
	}, nil
}
