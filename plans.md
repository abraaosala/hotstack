# Plano: HotStack Framework de Skills

## Contexto

Baseado em:
- **Pair** (Nuno Maduro) - Protocol for AI Rules para sincronizar regras entre editores
- **Hodstack** - CLI que torna coding agents mais produtivos com skills
- **Agent Skills Specification** - formato aberto para skills de agentes AI

## Objetivo

Criar uma framework de skills que:
1. Unifica regras AI em um diretório central (como Pair)
2. Permite criar/executar skills via CLI (como Hodstack)
3. Segue o padrão Agent Skills para compatibilidade

## Estrutura Proposta

```
myskills/
├── .ai/                          # Regras AI centralizadas (Pair-style)
│   ├── rules/
│   │   ├── 01_general.mdc
│   │   ├── 02_code_style.mdc
│   │   └── 03_project_context.mdc
│   └── PROJECT.md                # Contexto do projeto
├── skills/                       # Skills (Agent Skills format)
│   ├── <skill-name>/
│   │   ├── SKILL.md             # Metadata + instruções
│   │   ├── scripts/             # Código executável (opcional)
│   │   └── references/          # Documentação (opcional)
│   └── ...
├── evals/                        # Evals para validar skills
│   ├── bases/                   # Projetos fixture
│   ├── <skill-name>/            # Testes por skill
│   │   └── <test-case>/
│   │       ├── test.md          # TOML + descrição
│   │       └── fixtures/        # Ficheiros extras
│   └── src/                     # Runner (Rust)
├── bin/
│   └── mycli                     # CLI principal
├── AGENTS.md                     # Gerado automaticamente
└── CLAUDE.md                     # Gerado automaticamente
```

## Fases de Implementação

### Fase 1: CLI Core (Go + Cobra) ✓
- [x] `go mod init` + estrutura de pastas
- [x] Comando `init` - scaffolding do projeto
- [x] Comando `sync` - sincronizar regras AI para editores
- [x] Comando `list` - listar skills disponíveis
- [x] Comando `run <skill>` - executar uma skill
- [x] Comando `eval` - rodar evals de uma skill

### Fase 2: Sistema de Regras ✓
- [x] Parser de regras `.mdc` / `.md`
- [x] Suporte a Cursor, Copilot, Claude Code, OpenCode, Junie, Windsurf, Cline
- [x] Auto-geração de AGENTS.md, CLAUDE.md
- [x] Geração de .cursor/rules, .github/copilot-instructions, .cursorrules
- [x] Loader config.toml com [project] e [agents]
- [x] Testes unitários (internal/rules)
- [x] Contexto de projeto: `.hot/PROJECT.md` com template (ComoPegarContexto)
- [x] `sync` embute PROJECT.md em AGENTS.md/CLAUDE.md para evitar alucinações

### Fase 3: Sistema de Skills ✓
- [x] Loader de SKILL.md com metadata (internal/skills)
- [x] Parser de frontmatter (name, description, disable-model-invocation)
- [x] Execução de scripts bundlados (--script, rodam na raiz do projeto)
- [x] Suporte .sh/.bat/.cmd/.py/executável
- [x] Detection de agents (claude, opencode, cursor-agent, codex, gemini)
- [x] Testes unitários (internal/skills)
- [ ] Progressive disclosure (discovery → activation → execution) *parcial: loader carrega metadata sem ler instruções completas

### Fase 4: Sistema de Evals ✓
- [x] Formato de teste TOML (test.md) com frontmatter
- [x] Graders: tool_used, file_content, git_dirty, file_exists, file_not_exists, exit_code
- [x] Fixture projects (bases/) — SetupBase, CopyFixtures
- [x] InitGit para fixtures (init, config user, add, commit)
- [x] Comando `hotstack eval <skill> [case]` real
- [x] Runner com tmpDir por caso, avaliação de graders, relatório
- [x] Testes unitários (internal/evals): parse, graders, result

### Fase 5: Distribuição ✓
- [x] GitHub Actions workflow (`.github/workflows/release.yml`)
- [x] Cross-compile: linux/darwin/windows x amd64/arm64
- [x] Upload artifacts + checksums + release notes automáticos
- [x] `install.sh` para macOS/Linux (curl, deteta platforma, instala em ~/.local/bin)
- [x] `install.ps1` para Windows (PowerShell, adiciona ao PATH)
- [x] Makefile com `build-all`, `test`, `vet`, ldflags com versão

## Decisões Técnicas

| Componente | Tecnologia |
|------------|------------|
| Linguagem | **Go** |
| Config | **TOML** |
| Distribuição | **GitHub Releases** |
| CI/CD | GitHub Actions |
| Testes | `testing` + `testify` |
| CLI Framework | `cobra` ou `bubbletea` |
| TOML Parser | `github.com/BurntSushi/toml` |
| File Watcher | `fsnotify` |

## Sistema de Evals (Detalhes)

### Formato do test.md

```toml
+++
base = "catalogue"  # Projeto fixture base
intent = "Descrição da tarefa para o agent"

[[graders]]
type = "tool_used"        # Verifica se tool foi usada
tool = "Bash"
input_match = "npm test"

[[graders]]
type = "file_content"     # Verifica conteúdo de ficheiro
path = "package.json"
pattern = "\"version\": \"2\\.0\\.0\""
match = "contains"        # contains | not_contains

[[graders]]
type = "git_dirty"        # Verifica estado do git
min = 1                   # Mínimo de ficheiros alterados
max = 5                   # Máximo de ficheiros alterados
+++

Descrição do comportamento esperado do agent.
```

### Tipos de Graders

| Grader | Descrição | Params |
|--------|-----------|--------|
| `tool_used` | Tool foi invocada | `tool`, `input_match`, `min`, `max` |
| `file_content` | Ficheiro contém padrão | `path`, `pattern`, `match` |
| `git_dirty` | Git working tree alterado | `min`, `max` |
| `file_exists` | Ficheiro foi criado | `path` |
| `file_not_exists` | Ficheiro NÃO foi criado | `path` |
| `exit_code` | Exit code do último comando | `code` |

### Estrutura de Evals

```
evals/
├── bases/                    # Projetos fixture base
│   ├── php-composer/
│   ├── node-npm/
│   └── rust-cargo/
├── <skill-name>/             # Evals por skill
│   ├── <test-case>/
│   │   ├── test.md           # Teste TOML
│   │   └── fixtures/         # Ficheiros extras
│   └── ...
└── internal/                 # Runner Go
    ├── grader/
    │   ├── tool_used.go
    │   ├── file_content.go
    │   └── git_dirty.go
    ├── case.go
    └── judge.go
```

### Comando

```bash
hotstack eval deps-upgrade                    # Roda todos os evals da skill
hotstack eval deps-upgrade --case raises-minor  # Roda caso específico
hotstack eval --all                           # Roda todos os evals
```

## Stack Definida

### Linguagem: Go

```bash
# Estrutura do projeto
myskills/
├── cmd/
│   └── myskills/
│       └── main.go           # Entry point
├── internal/
│   ├── cli/                  # Comandos cobra
│   │   ├── init.go
│   │   ├── sync.go
│   │   ├── list.go
│   │   ├── run.go
│   │   └── eval.go
│   ├── rules/                # Parser de regras
│   ├── skills/               # Loader de skills
│   ├── evals/                # Runner de evals
│   │   ├── grader.go
│   │   ├── case.go
│   │   └── judge.go
│   └── sync/                 # Sync para editores
├── pkg/
│   └── toml/                 # Helpers TOML
├── evals/                    # Evals (fixture projects)
│   ├── bases/
│   └── <skill>/
├── go.mod
├── go.sum
└── Makefile
```

### Dependências Go

```go
module hotstack

go 1.22

require (
    github.com/spf13/cobra v1.8.0    // CLI framework
    github.com/BurntSushi/toml v1.3.2 // TOML parser
    github.com/fatih/color v1.16.0    // Output colorido
    github.com/stretchr/testify v1.9.0 // Testes
)
```

### GitHub Actions (CI/CD)

```yaml
# .github/workflows/release.yml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, darwin, windows]
        goarch: [amd64, arm64]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} \
          go build -o myskills-${{ matrix.goos }}-${{ matrix.goarch }} ./cmd/myskills
      - uses: softprops/action-gh-release@v1
        with:
          files: myskills-*
```

### Formato TOML (config)

```toml
# myskills.toml
[project]
name = "my-project"
description = "Projeto exemplo"

[agents]
cursor = true
copilot = true
claude = true
opencode = true

[[skills]]
name = "deps-upgrade"
path = "./skills/deps-upgrade"
description = "Upgrade dependencies safely"
```
