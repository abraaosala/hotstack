package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/abraaosala/hotstack/internal/rules"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

type syncTarget struct {
	filename string
	header   string
	format   string
}

var syncCmds = &cobra.Command{
	Use:   "sync",
	Short: "Sincroniza regras AI para os editores",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSync()
	},
}

func runSync() error {
	rulesDir := ".hot/rules"
	configFile := ".hot/config.toml"

	if _, err := os.Stat(rulesDir); os.IsNotExist(err) {
		return fmt.Errorf("diretório de regras não encontrado: %s", rulesDir)
	}

	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return fmt.Errorf("config não encontrado: %s", configFile)
	}

	parsed, err := rules.Parse(rulesDir)
	if err != nil {
		return err
	}

	if len(parsed) == 0 {
		color.Yellow("Nenhuma regra encontrada em %s", rulesDir)
	}

	cfg, err := rules.LoadConfig(configFile)
	if err != nil {
		return fmt.Errorf("erro ao carregar config: %w", err)
	}

	targets := targetsFor(cfg)

	projectCtx := loadProjectContext(".hot/PROJECT.md")
	if projectCtx == "" {
		color.Yellow("Aviso: .hot/PROJECT.md vazio. Preenche-o para o agent não alucinar sobre o projeto.")
	}

	generated := 0

	for _, target := range targets {
		content := renderTarget(parsed, cfg, projectCtx, target)

		dir := dirOf(target.filename)
		if dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				color.Red("✗ Erro ao criar diretório %s: %v", dir, err)
				continue
			}
		}

		if err := os.WriteFile(target.filename, []byte(content), 0644); err != nil {
			color.Red("✗ Erro ao gerar %s: %v", target.filename, err)
			continue
		}
		color.Green("✓ %-16s gerado", target.filename)
		generated++
	}

	fmt.Println(strings.Repeat("-", 40))
	fmt.Printf("Sincronizado para %d alvo(s) de %d editores configurados\n", generated, len(targets))

	syncSkills(cfg)

	return nil
}

func targetsFor(cfg rules.Config) []syncTarget {
	var targets []syncTarget
	agents := cfg.EnabledAgents()

	appendIf := func(enabled bool, t syncTarget) {
		if !enabled {
			return
		}
		targets = append(targets, t)
	}

	_ = agents

	appendIf(cfg.Agents.Claude, syncTarget{"AGENTS.md", "# Agent Rules", "markdown"})
	appendIf(cfg.Agents.Claude, syncTarget{"CLAUDE.md", "# Claude Rules", "markdown"})
	appendIf(cfg.Agents.OpenCode, syncTarget{"AGENTS.md", "# Agent Rules", "markdown"})
	appendIf(cfg.Agents.Cursor, syncTarget{".cursor/rules/general.mdc", "", "mdc"})
	appendIf(cfg.Agents.Cursor, syncTarget{".cursorrules", "", "markdown"})
	appendIf(cfg.Agents.Copilot, syncTarget{".github/copilot-instructions.md", "# Copilot Instructions", "markdown"})
	appendIf(cfg.Agents.Junie, syncTarget{".junie/rules.md", "# Junie Rules", "markdown"})
	appendIf(cfg.Agents.Windsurf, syncTarget{".windsurfrules", "", "markdown"})
	appendIf(cfg.Agents.Cline, syncTarget{".clinerules", "", "markdown"})

	return targets
}

// skillTargets maps each enabled agent to the directory where it discovers skills.
func skillTargets(cfg rules.Config) []string {
	var targets []string
	add := func(enabled bool, dir string) {
		if enabled {
			targets = append(targets, dir)
		}
	}
	add(cfg.Agents.OpenCode, ".opencode/skills")
	add(cfg.Agents.Claude, ".claude/skills")
	add(cfg.Agents.Claude, ".agents/skills")
	return targets
}

func syncSkills(cfg rules.Config) {
	for _, dir := range skillTargets(cfg) {
		n, err := copyLocalSkills(".hot/skills", dir)
		if err != nil {
			color.Red("✗ Erro ao copiar skills para %s: %v", dir, err)
			continue
		}
		if n > 0 {
			color.Green("✓ %-20s %d skill(s) sincronizada(s)", dir, n)
		}
	}
}

func dirOf(path string) string {
	idx := strings.LastIndex(path, "/")
	if strings.Contains(path, "\\") {
		back := strings.LastIndex(path, "\\")
		if back > idx {
			idx = back
		}
	}
	if idx == -1 {
		return ""
	}
	return path[:idx]
}

func loadProjectContext(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	trimmed := strings.TrimSpace(string(content))
	if strings.HasPrefix(trimmed, "# PROJECT.md") || strings.HasPrefix(trimmed, "> Responde") {
		return ""
	}
	return trimmed
}

func renderTarget(parsed []rules.Rule, cfg rules.Config, projectCtx string, target syncTarget) string {
	var b strings.Builder

	if target.header != "" {
		b.WriteString(target.header)
		b.WriteString("\n\n")
	}

	if cfg.Project.Name != "" {
		b.WriteString("> Projeto: ")
		b.WriteString(cfg.Project.Name)
		b.WriteString("\n")
	}
	if cfg.Project.Description != "" {
		b.WriteString("> ")
		b.WriteString(cfg.Project.Description)
		b.WriteString("\n")
	}

	if projectCtx != "" {
		b.WriteString("\n## Contexto do Projeto\n\n")
		b.WriteString(projectCtx)
		b.WriteString("\n\n")
	}

	for _, rule := range parsed {
		if target.format == "mdc" {
			b.WriteString("---\n")
			b.WriteString("description: ")
			b.WriteString(rule.Title())
			b.WriteString("\n---\n\n")
		} else {
			b.WriteString("## ")
			b.WriteString(rule.Title())
			b.WriteString("\n\n")
		}
		b.WriteString(rule.Content)
		b.WriteString("\n\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}
