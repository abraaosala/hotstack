# HotStack

Framework de skills para coding agents. Unifica regras AI (estilo Pair), executa skills (estilo Hodstack) e segue o padrão Agent Skills (SKILL.md).

## Instalação

### npm (recomendado)

```bash
npm install -g hotstack
```

### macOS / Linux

```bash
curl -fsSL https://github.com/abraaosala/hotstack/releases/latest/download/install.sh | sh
```

### Windows

```powershell
irm https://github.com/abraaosala/hotstack/releases/latest/download/install.ps1 | iex
```

### Local (dev)

```bash
make build   # ou: go build -o hotstack ./cli/hotstack
```

## Início rápido

```bash
hotstack init          # cria .hot/ com config.toml + PROJECT.md
# preenche .hot/PROJECT.md com o contexto do projeto
hotstack sync          # gera AGENTS.md, CLAUDE.md, .cursorrules, etc.
hotstack list          # lista skills
hotstack run <skill>   # executa uma skill
hotstack run <skill> --script report.sh   # executa um script bundlado
hotstack eval <skill>  # valida uma skill com evals
hotstack update        # atualiza para a versão mais recente
hotstack update --check # apenas verificar se há atualização
```

## Estrutura

```
.hot/
├── config.toml        # projeto + editores ativos
├── PROJECT.md         # contexto do projeto (evita alucinações)
├── rules/             # regras AI (01_general.md, etc.)
├── skills/            # skills (formato Agent Skills)
│   └── <skill>/
│       ├── SKILL.md
│       └── scripts/
└── evals/             # evals para validar skills
    ├── bases/         # projetos fixture
    └── <skill>/<case>/
        └── test.md    # TOML + graders
```

## Comandos

| Comando | Descrição |
|---------|-----------|
| `init` | Scaffolding do projeto HotStack |
| `sync` | Gera contexto para Cursor, Copilot, Claude, OpenCode, Junie, Windsurf, Cline |
| `list` | Lista skills com metadata |
| `run` | Executa skill (via agent) ou script bundlado |
| `eval` | Valida skill com graders |
| `update` | Atualiza para a versão mais recente |

## Evals

Cada caso de teste é um `test.md` com frontmatter TOML:

```toml
+++
base = "go-project"
intent = "Adicione uma função quadrado"

[[graders]]
type = "file_content"
path = "main.go"
pattern = "func quadrado"

[[graders]]
type = "git_dirty"
min = 1

[[graders]]
type = "exit_code"
command = "go build ./..."
code = 0

[[graders]]
type = "output_contains"
pattern = "build successful"
+++
```

### Tipos de Grader

| Grader | Descrição |
|--------|-----------|
| `tool_used` | Verifica se uma tool (bash, etc.) foi usada |
| `file_content` | Verifica se ficheiro contém padrão (regex ou string) |
| `file_exists` | Verifica se ficheiro existe |
| `file_not_exists` | Verifica se ficheiro NÃO existe |
| `git_dirty` | Verifica mudanças no working tree (min/max) |
| `exit_code` | Verifica exit code de comando |
| `output_contains` | Verifica padrão no output |
| `command_exists` | Verifica se comando existe no PATH |
| `snapshot` | Testes de snapshot (cria/compara) |

## Agentes suportados

Cursor, GitHub Copilot, Claude Code, OpenCode, Junie, Windsurf, Cline. Geração automática dos ficheiros de rules de cada um via `hotstack sync`, ativados no `config.toml`.

## Desenvolvimento

```bash
make test      # testes unitários
make vet       # go vet
make build-all # cross-compile para 6 plataformas
```

## Licença

MIT
