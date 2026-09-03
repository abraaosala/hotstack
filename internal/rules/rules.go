package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Rule struct {
	Path     string
	Filename string
	Order    int
	Content  string
}

type Config struct {
	Project ProjectConfig
	Agents  AgentsConfig
}

type ProjectConfig struct {
	Name        string
	Description string
}

type AgentsConfig struct {
	Cursor   bool
	Copilot  bool
	Claude   bool
	OpenCode bool
	Junie    bool
	Windsurf bool
	Cline    bool
}

var supportedAgents = map[string]bool{
	"cursor":   true,
	"copilot":  true,
	"claude":   true,
	"opencode": true,
	"junie":    true,
	"windsurf": true,
	"cline":    true,
}

func Parse(rulesDir string) ([]Rule, error) {
	entries, err := os.ReadDir(rulesDir)
	if err != nil {
		return nil, fmt.Errorf("erro ao ler diretório de regras: %w", err)
	}

	var rules []Rule
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") && !strings.HasSuffix(name, ".mdc") {
			continue
		}

		content, err := os.ReadFile(filepath.Join(rulesDir, name))
		if err != nil {
			return nil, fmt.Errorf("erro ao ler regra %s: %w", name, err)
		}

		rules = append(rules, Rule{
			Path:     filepath.Join(rulesDir, name),
			Filename: name,
			Order:    parseOrder(name),
			Content:  string(content),
		})
	}

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Order < rules[j].Order
	})

	return rules, nil
}

func (c Config) EnabledAgents() []string {
	var agents []string
	for name, enabled := range map[string]bool{
		"cursor":   c.Agents.Cursor,
		"copilot":  c.Agents.Copilot,
		"claude":   c.Agents.Claude,
		"opencode": c.Agents.OpenCode,
		"junie":    c.Agents.Junie,
		"windsurf": c.Agents.Windsurf,
		"cline":    c.Agents.Cline,
	} {
		if enabled {
			agents = append(agents, name)
		}
	}
	sort.Strings(agents)
	return agents
}

func parseOrder(filename string) int {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return 999
	}

	var order int
	for _, c := range parts[0] {
		if c < '0' || c > '9' {
			return 999
		}
		order = order*10 + int(c-'0')
	}
	return order
}

func (r Rule) Title() string {
	base := strings.TrimSuffix(strings.TrimSuffix(r.Filename, ".mdc"), ".md")
	idx := strings.Index(base, "_")
	if idx != -1 {
		base = base[idx+1:]
	}
	return strings.ReplaceAll(base, "_", " ")
}
