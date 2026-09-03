package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa um novo projeto HotStack",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

func init() {
	initCmd.Flags().StringVar(&agentFlag, "agent", "", "agente que vai trabalhar no projeto (opencode, claude, cursor, copilot, junie, windsurf, cline)")
}

var agentNames = []string{
	"opencode",
	"claude",
	"cursor",
	"copilot",
	"junie",
	"windsurf",
	"cline",
}

var agentFlag string

func runInit() error {
	dirs := []string{
		".hot",
		".hot/rules",
		".hot/skills",
		".hot/evals",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("erro ao criar diretório %s: %w", dir, err)
		}
	}

	agent, err := resolveAgent()
	if err != nil {
		return err
	}

	if err := writeConfig(".hot", agent); err != nil {
		return err
	}

	project := `# PROJECT.md

> Responde aqui às perguntas abaixo. O HotStack embute este ficheiro no contexto dos agents (AGENTS.md, CLAUDE.md, etc.) para que eles não alucinem sobre o teu projeto.

## O que é este projeto?

_Descreve em 2-3 frases o propósito e o valor do projeto._

## Stack / Tecnologias

- Linguagem:
- Framework:
- Banco de dados:
- Infraestrutura/deploy:
- Outros:

## Estrutura do projeto

_Descreve os diretórios principais e o que cada um faz._

## Comandos

- Instalar dependências:
- Rodar dev:
- Rodar testes:
- Build:
- Lint:

## Convenções

- Nomes de variáveis/funções:
- Estrutura de pastas:
- Tratamento de erros:
- Padrões de commit:

## Regras de negócio

_Lista as regras de negócio críticas que o agent deve respeitar._

## Dependências externas

_Serviços, APIs, credenciais (sem secrets) que o projeto usa._

## Gotchas / Armadilhas conhecidas

_O que costuma apanhar quem trabalha neste projeto._
`

	if err := os.WriteFile(filepath.Join(".hot", "PROJECT.md"), []byte(project), 0644); err != nil {
		return fmt.Errorf("erro ao criar PROJECT.md: %w", err)
	}

	skillsCopied, err := copyEmbeddedSkills(filepath.Join(".hot", "skills"))
	if err != nil {
		return fmt.Errorf("erro ao copiar skills: %w", err)
	}

	fmt.Println("✓ Projeto HotStack inicializado!")
	fmt.Println("  Diretório: .hot/")
	fmt.Println("  Config: .hot/config.toml")
	fmt.Println("  Contexto: .hot/PROJECT.md (preenche para o agent não alucinar)")
	fmt.Printf("  Agente:   %s\n", agent)
	if skillsCopied > 0 {
		fmt.Printf("  Skills: %d copiada(s) para .hot/skills/\n", skillsCopied)
	}
	return nil
}

// resolveAgent determina o agente: usa a flag --agent se fornecida, caso
// contrário pergunta ao utilizador interativamente.
func resolveAgent() (string, error) {
	if agentFlag != "" {
		if !isValidAgent(agentFlag) {
			return "", fmt.Errorf("agente inválido: %s (válidos: %s)", agentFlag, strings.Join(agentNames, ", "))
		}
		return agentFlag, nil
	}
	return promptAgent()
}

func isValidAgent(name string) bool {
	for _, a := range agentNames {
		if a == name {
			return true
		}
	}
	return false
}

// promptAgent pergunta ao utilizador qual agente vai trabalhar no projeto.
// Retorna o nome do agente escolhido.
func promptAgent() (string, error) {
	var agent string
	prompt := &survey.Select{
		Message: "Qual agente vais usar neste projeto?",
		Options: agentNames,
	}
	if err := survey.AskOne(prompt, &agent); err != nil {
		return "", fmt.Errorf("erro ao escolher agente: %w", err)
	}
	return agent, nil
}

// writeConfig gera .hot/config.toml ativando apenas o agente escolhido.
func writeConfig(hotDir, agent string) error {
	config := "[project]\n"
	config += "name = \"my-project\"\n"
	config += "description = \"Projeto exemplo\"\n"
	config += "\n"
	config += "[agents]\n"
	for _, name := range agentNames {
		enabled := "false"
		if name == agent {
			enabled = "true"
		}
		config += fmt.Sprintf("%s = %s\n", name, enabled)
	}

	return os.WriteFile(filepath.Join(hotDir, "config.toml"), []byte(config), 0644)
}
