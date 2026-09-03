---
name: github-pr-workflow
description: Boas práticas de workflow do GitHub para Pull Requests e code review — nomes de branch, mensagens de commit (Conventional Commits), criação de PR com `gh pr create`, template de descrição de PR, checklist de revisão de código, comentários de review, estratégias de merge (squash/merge/rebase) e como lidar com checks de CI. Use esta skill sempre que o usuário pedir para criar/abrir uma pull request, revisar código, escrever descrição de PR, definir convenção de branch ou commit, montar um checklist de code review, ou configurar um processo de PR para o time — mesmo que não use literalmente a palavra "workflow". Também use ao trabalhar com o `gh` CLI para PRs ou ao decidir estratégia de merge.
---

# Workflow de GitHub: Pull Requests e Code Review

Guia de boas práticas para todo o ciclo de vida de uma Pull Request (PR): criar a branch, commitar, abrir a PR, revisar o código e fazer o merge.

## Quando usar cada seção

- Vai criar uma branch ou commit? → **Convenções de Branch** e **Convenção de Commits**
- Vai abrir uma PR? → **Criando a Pull Request**
- Vai revisar código de outra pessoa? → **Checklist de Code Review**
- Vai comentar numa PR? → **Etiqueta de Comentários**
- Vai fazer merge? → **Estratégias de Merge**
- PR travada em CI? → **Lidando com Checks de CI**

Template pronto de descrição de PR está em `references/pr_template.md`.

## Convenções de Branch

Formato: `<tipo>/<descrição-curta-em-kebab-case>`

Tipos comuns:
- `feature/` ou `feat/` — nova funcionalidade
- `fix/` — correção de bug
- `hotfix/` — correção urgente em produção
- `chore/` — manutenção, dependências, configs
- `docs/` — apenas documentação
- `refactor/` — refatoração sem mudar comportamento

Exemplos: `feature/user-authentication`, `fix/login-timeout`, `docs/api-readme`.

Regras:
- Sempre a partir da branch principal atualizada (`main`/`master`/`develop`, conforme o repositório).
- Nomes curtos, descritivos, sem espaços ou maiúsculas.
- Se houver ticket/issue associado, incluir o número: `fix/1234-login-timeout`.

## Convenção de Commits

Recomenda-se **Conventional Commits**:

```
<tipo>(<escopo opcional>): <descrição no imperativo, minúscula>

[corpo opcional explicando o porquê]

[rodapé opcional: BREAKING CHANGE, Closes #123]
```

Tipos: `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`, `ci`, `build`.

Exemplos:
- `feat(auth): adiciona login via OAuth do Google`
- `fix(api): corrige timeout na rota de pagamentos`
- `chore: atualiza dependências do npm`

Regras práticas:
- Commits pequenos e atômicos — uma mudança lógica por commit.
- Assunto no imperativo ("adiciona", não "adicionado" ou "adicionando").
- Corpo do commit explica *por quê*, não apenas *o quê* (o diff já mostra o quê).
- Referenciar issues quando aplicável: `Closes #123`, `Refs #456`.

## Criando a Pull Request

Antes de abrir:
1. Atualize sua branch com a base (`git fetch && git rebase origin/main`, ou merge se o time preferir).
2. Rode testes e lint localmente.
3. Revise seu próprio diff antes de pedir revisão de outra pessoa.

Com o `gh` CLI:

```bash
gh pr create \
  --title "feat(auth): adiciona login via OAuth do Google" \
  --body-file pr_description.md \
  --base main \
  --reviewer usuario1,usuario2 \
  --label "feature"
```

Ou de forma interativa: `gh pr create --fill` (usa o último commit) ou apenas `gh pr create` para preencher tudo no prompt.

Título da PR: siga a mesma convenção dos commits (`tipo(escopo): descrição`), pois muitos repositórios usam o título da PR como mensagem do commit de squash.

Descrição da PR: use o template em `references/pr_template.md`. No mínimo, inclua:
- **O que muda** e por quê
- **Como testar**
- **Issue relacionada** (`Closes #123`)
- Screenshots/GIFs se houver mudança visual

Mantenha a PR pequena (idealmente < 400 linhas de diff). PRs grandes demoram mais para revisar e têm mais bugs escondidos — se estiver grande, considere dividir em PRs menores e incrementais.

## Checklist de Code Review

Ao revisar uma PR de outra pessoa, verificar:

**Corretude**
- [ ] O código faz o que a descrição da PR diz?
- [ ] Casos de borda (nulo, vazio, erro de rede, concorrência) foram tratados?
- [ ] Há testes cobrindo a mudança? Os testes realmente testam o comportamento certo?

**Design**
- [ ] A solução é razoavelmente simples para o problema? Não é over-engineering nem gambiarra?
- [ ] Segue os padrões e convenções já usados no repositório?
- [ ] Não duplica lógica que já existe em outro lugar?

**Segurança e desempenho**
- [ ] Entradas de usuário são validadas/sanitizadas?
- [ ] Não expõe segredos, chaves ou dados sensíveis?
- [ ] Nenhuma regressão óbvia de performance (loops desnecessários, N+1 queries)?

**Legibilidade**
- [ ] Nomes de variáveis/funções são claros?
- [ ] Comentários explicam o "porquê" apenas onde não é óbvio (não narram o código linha a linha)?

**Escopo**
- [ ] A PR faz uma coisa só, sem misturar refactor não relacionado?
- [ ] CI está verde (lint, testes, build)?

## Etiqueta de Comentários

- Comente o código, não a pessoa: "esse loop pode ser O(n²), dá pra simplificar?" em vez de "você escreveu isso errado".
- Diferencie bloqueante de sugestão. Prefixos úteis: `blocking:`, `nit:` (não bloqueante, cosmético), `question:`.
- Elogie o que está bom, não só aponte problemas.
- Se o comentário for longo ou opinativo demais, considere levar para uma call em vez de uma guerra de comentários.
- Como autor da PR, responda a todos os comentários (mesmo que só com "feito" ou explicando por que discorda) antes de pedir novo review.

## Estratégias de Merge

| Estratégia | Quando usar |
|---|---|
| **Squash and merge** | Padrão para a maioria dos times — histórico limpo, 1 commit por PR na branch principal |
| **Merge commit** | Quando o histórico de commits individuais da branch tem valor (ex.: monorepos com convenção própria) |
| **Rebase and merge** | Quando se quer histórico linear sem commit de merge, mas preservando commits individuais |

Regra geral: se o time não tem preferência definida, **squash and merge** é a opção mais segura e mais comum em projetos open source.

Depois do merge:
- Delete a branch (`gh pr merge --delete-branch` já faz isso).
- Confirme que o deploy/CI pós-merge passou.

## Lidando com Checks de CI

- PR vermelha por lint/formatação → rode o formatter localmente (`prettier`, `black`, etc.) e faça commit da correção.
- PR vermelha por teste → reproduza o teste localmente antes de tentar corrigir às cegas; não pule/desabilite o teste só para passar o CI.
- Flaky test (falha intermitente sem relação com a mudança) → re-rode o job (`gh run rerun <run-id>`) e, se persistir, abra uma issue separada em vez de mascarar o problema na PR atual.
- CI travado esperando aprovação de ambiente/segredo → verifique se é uma etapa manual esperada (ex.: deploy em produção) antes de assumir que é erro.

```bash
gh pr checks <numero-da-pr>      # ver status dos checks
gh run rerun <run-id>            # re-rodar um workflow que falhou
gh run view <run-id> --log-failed  # ver o log só da parte que falhou
```

## Referências

- `references/pr_template.md` — template pronto para descrição de PR, copiar e preencher.
