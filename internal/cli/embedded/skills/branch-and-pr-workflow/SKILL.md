---
name: branch-and-pr-workflow
description: Automatiza o fluxo completo de branches e PRs — analisa mudanças pendentes, agrupa em branches lógicas, cria branches, commits, push, cria PRs e faz merge. Use quando o usuário pedir para "criar branches separadas", "separar mudanças em PRs", "organizar commits em branches", ou quando houver muitas mudanças misturadas no working directory que precisam ser divididas. Também use quando o usuário disser "faz branches", "cria PRs", "separa isso em branches", ou similar.
---

# Workflow de Branches e PRs

Automatiza o ciclo completo: analisar → agrupar → branch → commit → push → PR → merge.

## Quando usar

- Working directory com muitas mudanças misturadas que precisam ser separadas
- Usuário pede para "criar branches separadas" ou "organizar em PRs"
- Usuário pede para "fazer push" ou "criar PR" de mudanças pendentes
- Usuário quer fazer merge de PRs depois de tudo limpo

## Fluxo completo

### 1. Analisar mudanças

```bash
git status --short
git diff --stat
```

Agrupar ficheiros por lógica:
- **Módulo** (User, Accounting, RH)
- **Camada** (Models, Repos, Services, Controllers, Views)
- **Tipo** (feat, fix, refactor, chore)
- **Feature nova** vs melhoria existente

### 2. Criar branches

Para cada grupo lógico:

```bash
# Criar branch a partir de master actualizado
git checkout -b <tipo>/<nome> master

# Adicionar apenas os ficheiros do grupo
git add <ficheiros>

# Commit com mensagem Conventional Commits
git commit -m "<tipo>(<escopo>): <descrição>"

# Voltar a master
git checkout master
```

### 3. Push

```bash
git push origin <todas-as-branches>
```

### 4. Criar PRs

Para cada branch:

```bash
gh pr create \
  --base master \
  --head <branch> \
  --title "<tipo>(<escopo>): <descrição>" \
  --body "## O que muda
<descrição>

## Como testar
<passos>

## Notas
<observações>"
```

### 5. Verificar PRs

```bash
gh pr list --state open --json number,title,headRefName,statusCheckRollup
```

Esperar que todos os checks passem (CodeRabbit, CI, etc.).

### 6. Merge

**IMPORTANTE:** Usar `--merge` (não `--squash`) para preservar histórico das branches:

```bash
# Ordem de merge (respeitar dependências)
git pull origin master
gh pr merge <PR1> --merge --delete-branch
git pull origin master
gh pr merge <PR2> --merge --delete-branch
# ... etc
```

Depois de cada merge, fazer `git pull` para actualizar master antes do próximo merge.

### 7. Limpar branches locais

```bash
git checkout master
git branch -d <branch1> <branch2> ...
```

Ou usar `--delete-branch` no merge para apagar automático.

## Convenções

### Branches

Formato: `<tipo>/<descrição-curta-em-kebab-case>`

- `feat/` — nova funcionalidade
- `fix/` — correção de bug
- `refactor/` — refatoração sem mudar comportamento
- `chore/` — manutenção, dependências, configs
- `docs/` — apenas documentação

### Commits

Conventional Commits:
```
<tipo>(<escopo>): <descrição no imperativo>
```

### PRs

- Título segue mesma convenção dos commits
- Body com "O que muda", "Como testar", "Notas"
- Merge com `--merge` para preservar histórico

## Exemplo prático

Entrada: working directory com 80+ ficheiros modificados

1. Analisar e agrupar em 8 branches
2. Criar branches, commits, push
3. Criar 8 PRs
4. Verificar checks
5. Merge em ordem (respeitando dependências)
6. Limpar branches

Resultado: master com histórico completo de branches via merge commits.

## Notas importantes

- **Sempre usar `--merge`** no `gh pr merge` para preservar histórico
- **Fazer `git pull`** após cada merge para manter master actualizado
- **Verificar checks** antes de mergear (CodeRabbit, CI)
- **Respeitar ordem de dependências** entre branches
