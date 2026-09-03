# PHP 8.4 e 8.5 — Recursos Novos com Exemplos

Todos os exemplos seguem a formatação PSR-12/PER (ver `psr-12-style.md`).

## PHP 8.4 (lançado em 21/11/2024)

### Property hooks

Elimina getters/setters manuais para lógica simples de acesso.

```php
// Antes (PHP < 8.4)
final class User
{
    private string $first;
    private string $last;

    public function getFullName(): string
    {
        return "{$this->first} {$this->last}";
    }
}

// Depois (PHP 8.4)
final class User
{
    public string $fullName {
        get => "{$this->first} {$this->last}";
    }

    public function __construct(
        private string $first,
        private string $last,
    ) {
    }
}
```

Hooks também podem ter `set` com validação, substituindo setters que faziam apenas validação simples:

```php
final class Product
{
    public function __construct(
        public string $name,
        private float $price {
            set(float $value) {
                if ($value < 0) {
                    throw new \ValueError('Preço não pode ser negativo.');
                }
                $this->price = $value;
            }
        },
    ) {
    }
}
```

Nota: property hooks são incompatíveis com `readonly`. Para restringir escrita mantendo comportamento customizado, combinar com visibilidade assimétrica (abaixo) em vez de `readonly`.

### Visibilidade assimétrica

Propriedade com visibilidade diferente para leitura e escrita — substitui o padrão "propriedade privada + getter público".

```php
// Antes (PHP < 8.4)
final class BankAccount
{
    private float $balance;

    public function getBalance(): float
    {
        return $this->balance;
    }
}

// Depois (PHP 8.4)
final class BankAccount
{
    public private(set) float $balance = 0.0;
}
```

`public protected(set)` permite que subclasses escrevam, mas não código externo. Combinada com property hooks, dá controle total sobre leitura/escrita:

```php
final class Person
{
    private(set) \DateTimeImmutable $birthDate {
        set => $value > new \DateTimeImmutable()
            ? throw new \InvalidArgumentException('Data no futuro.')
            : $value;
    }
}
```

### Novas funções de array

Preferir a `foreach` manual para busca/validação em coleções:

```php
// Antes
$admin = null;
foreach ($users as $user) {
    if ($user->isAdmin) {
        $admin = $user;
        break;
    }
}

// Depois (PHP 8.4)
$admin = array_find($users, fn (User $u) => $u->isAdmin);
$hasAdmin = array_any($users, fn (User $u) => $u->isAdmin);
$allVerified = array_all($users, fn (User $u) => $u->isVerified);
$adminKey = array_find_key($users, fn (User $u) => $u->isAdmin);
```

### `new` encadeado sem parênteses extras

```php
// Antes
$slug = (new Slugger())->slugify($title);

// Depois (PHP 8.4)
$slug = new Slugger()->slugify($title);
```

### Outros recursos do 8.4

- Parser HTML5 nativo e compatível com o padrão (`Dom\HTMLDocument::createFromString()`), substituindo o antigo `DOMDocument::loadHTML()` baseado em libxml/HTML4.
- Funções multibyte novas: `mb_trim()`, `mb_ltrim()`, `mb_rtrim()`, `mb_str_pad()`, `mb_ucfirst()`, `mb_lcfirst()`.
- `round()` com modos de arredondamento explícitos via enum `RoundingMode`.
- Propriedades podem ser declaradas em **interfaces** (via property hooks), permitindo contratos com propriedades tipadas, não só métodos.
- Suporte a verbos HTTP adicionais em `$_POST`/`$_FILES` (ex.: `PUT`, `PATCH`, `DELETE` com corpo `multipart/form-data`).

## PHP 8.5 (lançado em 20/11/2025)

### Pipe operator (`|>`)

Encadeia chamadas sem aninhar nem criar variáveis intermediárias:

```php
// Antes
$output = strtolower(
    str_replace(' ', '-', trim($input))
);

// Depois (PHP 8.5)
$output = $input
    |> trim(...)
    |> fn ($s) => str_replace(' ', '-', $s)
    |> strtolower(...);
```

Usar quando a leitura ficar mais clara em pipeline; para transformações simples de 1-2 passos, uma chamada aninhada normal ainda pode ser mais direta — julgar caso a caso.

### `clone(...)` com sobrescrita de propriedades ("clone with")

Padrão nativo de "wither" para objetos imutáveis/`readonly`, sem precisar de um método `withX()` para cada propriedade:

```php
// Antes
final class Money
{
    public function __construct(
        public readonly int $amount,
        public readonly string $currency,
    ) {
    }

    public function withAmount(int $amount): self
    {
        return new self($amount, $this->currency);
    }
}

// Depois (PHP 8.5)
final class Money
{
    public function __construct(
        public readonly int $amount,
        public readonly string $currency,
    ) {
    }
}

$updated = clone($money, amount: 500);
```

### `#[\NoDiscard]`

Sinaliza que o retorno de uma função/método não deve ser ignorado — útil em builders imutáveis e funções puras onde ignorar o retorno quase sempre é um bug:

```php
final class QueryBuilder
{
    #[\NoDiscard('O retorno precisa ser usado; QueryBuilder não é mutável.')]
    public function where(string $condition): self
    {
        return clone($this, conditions: [...$this->conditions, $condition]);
    }
}

$qb->where('id = 1'); // PHP emite warning: retorno descartado
```

### `array_first()` / `array_last()`

```php
// Antes
$first = $items[array_key_first($items)];
$last = $items[array_key_last($items)];

// Depois (PHP 8.5)
$first = array_first($items);
$last = array_last($items);
```

### Extensão `URI` nativa

Substitui `parse_url()` por objetos imutáveis e validados (RFC 3986 e WHATWG):

```php
// Antes
$parts = parse_url($url); // array solto, sem validação forte

// Depois (PHP 8.5)
$uri = \Uri\Rfc3986\Uri::parse($url);
$host = $uri->getHost();
$withNewPath = $uri->withPath('/novo-caminho'); // imutável
```

### Outros recursos do 8.5

- Closures e first-class callables agora podem ser usados em expressões de constantes (inclusive dentro de atributos).
- `get_error_handler()` / `get_exception_handler()` para introspecção do handler atual.
- Promoção de propriedade `final` direto no construtor: `public function __construct(final public string $id) {}`.
- Atributos podem ser aplicados a constantes de classe.
- `fatal_error_backtraces` (INI) adiciona stack trace a erros fatais, respeitando `#[\SensitiveParameter]`.
- Flag `php --ini=diff` mostra apenas configurações do `php.ini` diferentes do padrão.
- `curl_multi_get_handles()`.

## Regra prática de adoção

- **Código novo**: usar os recursos acima sempre que simplificarem o código (menos boilerplate, mais imutabilidade, tipos mais expressivos).
- **Código legado**: não abrir PRs só para "modernizar" — refatorar oportunisticamente quando o arquivo já estiver sendo alterado por outro motivo.
- **Compatibilidade**: confirmar a versão mínima de PHP do projeto (`composer.json` → `"php": ">=8.4"`) antes de usar qualquer recurso desta lista; nada aqui funciona em PHP 8.3 ou anterior.
