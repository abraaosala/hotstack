# PSR-12 / PER Coding Style — Regras Detalhadas

Extensão do resumo no SKILL.md. Esta referência cobre casos específicos de formatação que surgem no dia a dia.

## Declaração de tipos

```php
declare(strict_types=1);
```

Logo após `<?php`, antes de `namespace` e `use`. Nunca esquecer — garante type safety em toda a file.

## Classes e interfaces

- Chave de abertura `{` em **nova linha**.
- Uma classe/interface por arquivo.
- `extends` e `implements` na mesma linha que a declaração da classe.

```php
final class UserRegistration implements SendsWelcomeEmail
{
    // ...
}
```

## Propriedades

- Visibilidade explícita obrigatória.
- Tipo obrigatório (PHP 7.4+).
- Um `use` por trait; visibilidade pode ser redefinida.

```php
final class Config
{
    use \App\Traits\HasTimeout;

    public private(set) string $name;

    protected int $timeout = 30;
}
```

## Métodos

- Chave de abertura `{` em **nova linha**.
- Tipo de retorno obrigatório.
- Visibilidade explícita.
- `static` antes do tipo de retorno.

```php
public static function fromArray(array $data): self
{
    return new self(
        name: $data['name'] ?? '',
    );
}
```

## Argumentos

- Um argumento por linha quando a lista ultrapassar 80 caracteres.
- Último argumento **sem** vírgula trailing.

```php
final class HttpClient
{
    public function __construct(
        private readonly string $baseUrl,
        private readonly int $timeout,
        private readonly bool $followRedirects,
    ) {
    }
}
```

## Closures

- `fn` para closures expressões (uma expressão, sem `return` explícito).
- `function` para closures com múltiplas instruções.
- Espaço após `fn`, antes do parêntese de argumentos.

```php
$double = fn (int $n): int => $n * 2;

$process = function (string $input): string {
    $trimmed = trim($input);
    return strtolower($trimmed);
};
```

## Arrays e listas

- Chaves de abertura `[` na mesma linha.
- Uma chave por linha para arrays associativos longos.
- Trailing comma em arrays multilinha.

```php
$config = [
    'host' => 'localhost',
    'port' => 3306,
    'charset' => 'utf8mb4',
];
```

## Operadores ternários

- Parênteses explícitos para ternários aninhados.
- Operador ternário com `?:` pode ser usado sem parênteses se legível.

```php
$level = $isAdmin ? 'admin' : ($isUser ? 'user' : 'guest');
```

## Expressões condicionais complexas

- Cada condição em sua própria linha quando houver 3+ condições.
- Operador lógico **no início** da linha seguinte (não no fim da anterior).

```php
if (
    $user->isActive()
    && $user->hasPermission('write')
    && $project->isWritable()
) {
    // ...
}
```

## `match` expression

- Pode ser usado em vez de `switch` para retornos simples.
- Chaves de abertura na mesma linha.
- `default` como último case.

```php
$status = match ($httpCode) {
    200, 201 => 'success',
    404 => 'not_found',
    500 => 'server_error',
    default => 'unknown',
};
```

## Enumerações

- `enum` em vez de constantes de classe para grupos finitos.
- `implements` para enums com comportamento extra (serialization, etc.).

```php
enum Status: string
{
    case Pending = 'pending';
    case Active = 'active';
    case Archived = 'archived';

    public function label(): string
    {
        return ucfirst($this->value);
    }
}
```

## Atributos

- Um atributo por linha quando houver múltiplos.
- Sem espaços extras entre colchetes.

```php
#[Route('/users', methods: ['GET', 'POST'])]
#[Middleware(AuthMiddleware::class)]
public function handle(): Response
{
    // ...
}
```

## Named arguments

- Usar quando o nome do argumento melhora a legibilidade.
- Nunca misturar argumentos posicionais e nomeados (exceto após `...`).

```php
$uri = new Uri(
    scheme: 'https',
    host: 'example.com',
    path: '/api/users',
);
```
