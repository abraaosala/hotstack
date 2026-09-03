---
name: php-psr-best-practices
description: Boas práticas de programação PHP seguindo os padrões PHP-FIG (PSR-1, PSR-4, PSR-12/PER Coding Style, PSR-3, PSR-7/15/17/18, PSR-11, PSR-14, PSR-20, entre outros) combinadas com os recursos modernos do PHP 8.4 e 8.5 (property hooks, visibilidade assimétrica, pipe operator, clone with, novas funções de array, etc.). Use esta skill sempre que o usuário pedir para escrever, revisar, refatorar ou dar boas práticas de código PHP, configurar um padrão de codificação para o time, criar classes/interfaces/DTOs em PHP, ou mencionar PSR, PHP-FIG, "código PHP moderno", PHP 8.4, PHP 8.5, ou compatibilidade com versões novas do PHP — mesmo que o usuário não cite "PSR" ou "boas práticas" explicitamente. Também use ao revisar Pull Requests com código PHP ou ao definir `.php-cs-fixer` / `phpcs` para um projeto.
---

# Boas Práticas de PHP: PSR + PHP 8.4/8.5

Guia de boas práticas de programação PHP baseado nos **PHP Standards Recommendations (PSR)** do PHP-FIG (https://www.php-fig.org/psr/) e nos recursos modernos disponíveis a partir do **PHP 8.4** e **PHP 8.5**.

## Quando usar cada seção

- Vai formatar/estilizar código (indentação, chaves, espaços)? → **PSR-12 / PER Coding Style** (`references/psr-12-style.md`)
- Vai nomear classes, métodos, constantes, organizar arquivos? → **PSR-1: Padrão Básico**
- Vai organizar namespaces e autoload via Composer? → **PSR-4: Autoloading**
- Vai criar uma interface/contrato (logger, cache, HTTP, container, eventos)? → **Outras PSRs de Interoperabilidade**
- Vai escrever código novo em PHP 8.4/8.5 e quer usar os recursos mais recentes? → **PHP 8.4** e **PHP 8.5** (`references/php-84-85-features.md`)

## PSR-1: Padrão Básico de Codificação

Base obrigatória antes de qualquer outra PSR:

- Arquivos PHP usam **apenas** as tags `<?php ?>` ou `<?= ?>` (nunca tags curtas `<? ?>` ou ASP).
- Arquivos em **UTF-8 sem BOM**.
- Um arquivo **deve** declarar símbolos (classes, funções, constantes) **ou** causar efeitos colaterais (output, `include`, alterar `ini_set`, conexões externas) — **nunca os dois** ao mesmo tempo.
- Namespaces e classes seguem PSR-4 (autoloading). Um arquivo = uma classe.
- `ClassName` em `StudlyCaps` (PascalCase).
- `CONST_NAME` em maiúsculas com underscore.
- `methodName()` em `camelCase`.
- Nomes de propriedades: qualquer convenção (`$camelCase`, `$snake_case`), mas consistente dentro do projeto.

## PSR-12 / PER Coding Style: Formatação

PSR-12 substitui a PSR-2 (deprecada) e é hoje mantida como **PER Coding Style**. Regras essenciais:

- Indentação: **4 espaços**, nunca tabs.
- Limite de linha: sem limite rígido; **120 caracteres** como limite suave; preferir até 80 quando possível.
- `declare(strict_types=1);` logo após a tag `<?php`, antes do `namespace`.
- Uma classe por arquivo; chave de abertura `{` de classes e métodos **em nova linha**.
- Chave de abertura de estruturas de controle (`if`, `for`, `while`...) **na mesma linha**.
- Palavras reservadas e tipos sempre em **minúsculo** (`int`, `bool`, `string`, nunca `INT`/`Integer`).
- Visibilidade (`public`/`protected`/`private`) **obrigatória** em todas as propriedades, constantes (PHP 7.1+) e métodos — nunca confiar no padrão implícito.
- `elseif` em vez de `else if`.
- Sem tag de fechamento `?>` em arquivos que contêm apenas PHP.

Exemplo mínimo correto:

```php
<?php

declare(strict_types=1);

namespace App\Domain;

final class Invoice
{
    public function __construct(
        private readonly string $id,
        private readonly float $total,
    ) {
    }

    public function isPaid(): bool
    {
        return $this->total <= 0.0;
    }
}
```

Detalhes completos de formatação (closures, argumentos multilinha, `use` de traits, ternário, operadores) estão em `references/psr-12-style.md`.

## PSR-4: Autoloading

- Namespace totalmente qualificado mapeia para caminho de diretório: `App\Domain\Invoice` → `src/Domain/Invoice.php` (conforme `composer.json`).
- Cada namespace de nível superior é um "vendor" ou pacote.
- Configuração no `composer.json`:

```json
{
    "autoload": {
        "psr-4": {
            "App\\": "src/"
        }
    }
}
```

- Depois de editar, rodar `composer dump-autoload`.
- Nunca misturar autoload PSR-4 com `require`/`include` manual para as mesmas classes.

## Outras PSRs de Interoperabilidade

Use a interface padrão do PHP-FIG em vez de inventar uma própria — isso permite trocar implementações (Monolog, Symfony, PSR-18 HTTP clients, etc.) sem acoplar o código da aplicação a uma biblioteca específica:

| PSR | Para quê | Pacote comum |
|---|---|---|
| **PSR-3** | Interface de logger (`LoggerInterface`) | Monolog |
| **PSR-4** | Autoloading | Composer |
| **PSR-6 / PSR-16** | Cache (genérico / simples) | Symfony Cache |
| **PSR-7** | Mensagens HTTP (request/response imutáveis) | Guzzle, Nyholm |
| **PSR-11** | Container de injeção de dependência (`ContainerInterface`) | PHP-DI, Symfony DI |
| **PSR-12 / PER** | Estilo de código | PHP-CS-Fixer, PHPCS |
| **PSR-14** | Event Dispatcher | Symfony EventDispatcher |
| **PSR-15** | Middlewares HTTP (`RequestHandlerInterface`) | Mezzio, Slim |
| **PSR-17** | Factories para objetos PSR-7 | Nyholm/psr7 |
| **PSR-18** | Cliente HTTP | Guzzle, Symfony HttpClient |
| **PSR-20** | Interface de relógio (`ClockInterface`), essencial para testar código dependente de tempo | Symfony Clock |

PSR-0 e PSR-2 estão **deprecadas** — não usar em código novo. PSR-5 (PHPDoc), PSR-19 (tags PHPDoc), PSR-21 (i18n) e PSR-22 (tracing) ainda estão em **Draft**, então tratar como referência, não como norma obrigatória.

Ao criar uma dependência nova (logger, cache, cliente HTTP), preferir depender da **interface PSR**, não da implementação concreta:

```php
final class PaymentService
{
    public function __construct(
        private readonly \Psr\Log\LoggerInterface $logger,
        private readonly \Psr\SimpleCache\CacheInterface $cache,
    ) {
    }
}
```

---

## PHP 8.4: recursos a favorecer em código novo

PHP 8.4 (lançado em novembro de 2024) trouxe mudanças que **reduzem boilerplate de getters/setters** e devem ser preferidas a padrões antigos ao escrever código novo:

- **Property hooks** (`get`/`set` na própria propriedade) — elimina a necessidade de métodos `getX()`/`setX()` manuais.
- **Visibilidade assimétrica** (`public private(set)`) — propriedade pública para leitura, mas só gravável internamente; substitui o padrão "propriedade privada + getter".
- **Novas funções de array**: `array_find()`, `array_any()`, `array_all()`, `array_find_key()` — preferir a `foreach` manual para buscas/validações em array.
- **`new` encadeado sem parênteses**: `new Slugger()->slugify($title)` — não é mais necessário envolver `new X()` em parênteses extras para encadear chamada.
- Parser HTML5 nativo (`Dom\HTMLDocument`), funções multibyte novas (`mb_trim`, `mb_str_pad`, etc.), modos de arredondamento explícitos em `round()`.

Exemplos e comparação "antes/depois" completos em `references/php-84-85-features.md`.

## PHP 8.5: recursos a favorecer em código novo

PHP 8.5 (lançado em novembro de 2025) adiciona:

- **Pipe operator (`|>`)** — encadeia chamadas de função sem aninhar nem variáveis intermediárias; preferir a `array_map`/composição manual quando a leitura melhorar.
- **`clone(...)` com sobrescrita de propriedades** ("clone with") — forma nativa do padrão "wither" para objetos imutáveis/`readonly`, sem precisar reescrever um método `withX()` para cada propriedade.
- **`#[\NoDiscard]`** — atributo para sinalizar que o valor de retorno de uma função não deve ser descartado (ex.: métodos que retornam `self` em builders imutáveis); usar em APIs de biblioteca onde ignorar o retorno é quase sempre um bug do chamador.
- **`array_first()` / `array_last()`** — preferir a `$array[array_key_first($array)]`.
- **Extensão `URI`** nativa (RFC 3986 + WHATWG) — preferir a `parse_url()` para manipular URLs de forma imutável e validada.
- Closures e first-class callables em expressões de constantes/atributos, introspecção do handler de erro atual, `fatal_error_backtraces`.

**Regra prática de adoção**: usar os recursos novos em código novo; só refatorar código legado quando ele for tocado por outro motivo (evitar PRs gigantes só de "modernização").

---

## Checklist rápido antes de finalizar um PR/arquivo PHP

- [ ] `declare(strict_types=1);` presente
- [ ] Tipagem em todos os parâmetros, propriedades e retornos (incluir tipos union/nullable quando aplicável)
- [ ] Visibilidade explícita em toda propriedade, constante e método
- [ ] Indentação de 4 espaços, sem tabs, sem trailing whitespace
- [ ] Um `use` por linha de import, sem barra invertida inicial
- [ ] Depende de interfaces PSR (Logger, Cache, Container, HTTP) em vez de implementações concretas
- [ ] Getters/setters triviais substituídos por property hooks / visibilidade assimétrica quando fizer sentido (PHP 8.4+)
- [ ] Sem `?>` de fechamento em arquivos 100% PHP
- [ ] Código passa em `phpcs`/`php-cs-fixer` configurado para PSR-12 (ou PER Coding Style)

## Referências

- `references/psr-12-style.md` — regras detalhadas de formatação PSR-12/PER (closures, argumentos multilinha, operadores, classes anônimas).
- `references/php-84-85-features.md` — exemplos de código "antes/depois" para cada recurso novo do PHP 8.4 e 8.5.
