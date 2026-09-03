package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Inicializa um novo projeto HotStack",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInit()
	},
}

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

	config := `[project]
name = "my-project"
description = "Projeto exemplo"

[agents]
cursor = true
copilot = true
claude = true
opencode = true
junie = false
windsurf = false
cline = false
`

	if err := os.WriteFile(filepath.Join(".hot", "config.toml"), []byte(config), 0644); err != nil {
		return fmt.Errorf("erro ao criar config: %w", err)
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
	if skillsCopied > 0 {
		fmt.Printf("  Skills: %d copiada(s) para .hot/skills/\n", skillsCopied)
	}
	return nil
}
