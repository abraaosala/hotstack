---
name: clean-architecture
description: Clean Architecture — separação de responsabilidades por camadas (Domain, Application, Infrastructure, Presentation), Dependency Rule, Ports and Adapters (Hexagonal Architecture), use cases, DTOs, mappers, testabilidade e estrutura de pastas. Use esta skill sempre que o usuário pedir para estruturar um projeto, organizar pastas e camadas, separar responsabilidades, implementar use cases, criar ports/adapters, definir a arquitetura de uma aplicação, ou quando estiver revisando código que mistura lógica de negócio com infraestrutura, framework ou apresentação — mesmo que não mencione "Clean Architecture" ou "Hexagonal" explicitamente. Também use ao definir a estrutura de um monolito, microserviço, ou ao migrar código legado para arquitetura mais testável.
---

# Clean Architecture

Guia prático de Clean Architecture baseado nos princípios de Robert C. Martin (Uncle Bob), Ports and Adapters (Hexagonal) e separação de responsabilidades por camadas.

## Quando usar cada seção

- Vai decidir onde colocar lógica de negócio vs lógica de acesso a dados? → **As Camadas**
- Vai criar um use case / caso de uso? → **Use Cases (Application Layer)**
- Vai integrar com banco de dados, API externa ou framework? → **Ports e Adapters**
- Vai definir a estrutura de pastas do projeto? → **Estrutura de Pastas**
- Vai testar a lógica de negócio sem depender de infraestrutura? → **Testabilidade**
- Vai revisar se o código está acoplado ao framework ou driver? → **Dependency Rule**
- Vai migrar código legado para arquitetura limpa? → **Estratégia de Migração**

---

## As Camadas

```
┌───────────────────────────────────────────┐
│          Presentation (Controllers)       │  ← Recebe input HTTP/CLI
├───────────────────────────────────────────┤
│          Application (Use Cases)          │  ← Orquestra o fluxo
├───────────────────────────────────────────┤
│          Domain (Entities + Rules)        │  ← Coração do sistema
├───────────────────────────────────────────┤
│          Infrastructure (Drivers)         │  ← Framework, DB, APIs
└───────────────────────────────────────────┘
```

Cada camada depende **apenas** das camadas internas. A camada de apresentação depende da application; a application depende do domain. A infraestrutura depende de todas (é a mais externa).

### Domain (Núcleo)

Contém entidades, value objects, regras de negócio, interfaces de ports (para persistência, eventos, etc.). **Nunca** depende de frameworks, drivers ou bibliotecas externas.

```php
namespace App\Domain\Order;

use App\Domain\Customer\CustomerId;
use App\Domain\Shared\Money;

final class Order
{
    private array $items = [];
    private OrderStatus $status = OrderStatus::Pending;

    public function addItem(ProductId $productId, int $quantity, Money $price): void
    {
        if ($this->status !== OrderStatus::Pending) {
            throw new \DomainException('Só é possível adicionar itens a pedidos pendentes.');
        }
        $this->items[] = new OrderItem($productId, $quantity, $price);
    }

    public function confirm(): void
    {
        if (empty($this->items)) {
            throw new \DomainException('Pedido não pode ser confirmado sem itens.');
        }
        $this->status = OrderStatus::Confirmed;
    }

    public function total(): Money
    {
        return array_reduce(
            $this->items,
            fn (Money $sum, OrderItem $item) => $sum->add($item->subtotal()),
            Money::zero('BRL'),
        );
    }
}
```

### Application (Use Cases)

Orquestra o fluxo: recebe input, chama methods do domínio, chama ports (repositórios, gateways), retorna output. **Não contém regras de negócio** — apenas coordenação.

```php
namespace App\Application\Order;

use App\Domain\Order\Order;
use App\Domain\Order\OrderRepository;
use App\Domain\Customer\CustomerId;

final class CreateOrder
{
    public function __construct(
        private readonly OrderRepository $repository,
    ) {
    }

    public function __invoke(CreateOrderRequest $request): OrderResponse
    {
        $order = new Order(
            id: OrderId::generate(),
            customerId: new CustomerId($request->customerId),
        );

        foreach ($request->items as $item) {
            $order->addItem(
                productId: new ProductId($item->productId),
                quantity: $item->quantity,
                price: new Money($item->priceAmount, $item->priceCurrency),
            );
        }

        $this->repository->save($order);

        return new OrderResponse(
            id: $order->id()->value(),
            total: $order->total()->amount(),
        );
    }
}
```

### Infrastructure (Drivers/Adapters)

Implementa interfaces definidas no domínio (ports): repositórios, gateways de pagamento, clients HTTP, etc. É onde frameworks, ORMs, drivers de banco vivem.

```php
namespace App\Infrastructure\Persistence;

use App\Domain\Order\Order;
use App\Domain\Order\OrderRepository;

final class DoctrineOrderRepository implements OrderRepository
{
    public function __construct(
        private readonly \Doctrine\ORM\EntityManagerInterface $em,
    ) {
    }

    public function findById(OrderId $id): ?Order
    {
        return $this->em->find(Order::class, $id->value());
    }

    public function save(Order $order): void
    {
        $this->em->persist($order);
        $this->em->flush();
    }
}
```

### Presentation (Controllers/Handlers)

Recebe input do mundo exterior (HTTP, CLI, queues), valida, e delega para use cases. **Não contém lógica de negócio** — apenas validação de input e formatação de output.

```php
namespace App\Infrastructure\Http;

use App\Application\Order\CreateOrder;
use App\Application\Order\CreateOrderRequest;

final class OrderController
{
    public function __construct(
        private readonly CreateOrder $createOrder,
    ) {
    }

    public function store(\Psr\Http\Message\ServerRequestInterface $request): array
    {
        $body = json_decode((string) $request->getBody(), true);

        $result = ($this->createOrder)(new CreateOrderRequest(
            customerId: $body['customer_id'],
            items: $body['items'],
        ));

        return ['status' => 'created', 'order' => $result];
    }
}
```

---

## Dependency Rule

A regra fundamental: **dependências apontam para dentro**. Camadas externas dependem das internas, nunca ao contrário.

```
Presentation → Application → Domain ← Infrastructure
                                   ↑
                            (ports/interfaces)
```

O domínio define **ports** (interfaces); a infraestrutura implementa **adapters**. Isso permite trocar drivers (Doctrine → Eloquent, Guzzle → Symfony HttpClient) sem mudar uma linha do domínio.

```php
// Domínio define o port (interface)
namespace App\Domain\Order;

interface OrderRepository
{
    public function findById(OrderId $id): ?Order;
    public function save(Order $order): void;
}

// Infraestrutura implementa o adapter
namespace App\Infrastructure\Persistence;

final class EloquentOrderRepository implements \App\Domain\Order\OrderRepository
{
    public function __construct(
        private readonly \App\Infrastructure\Database\Models\OrderModel $model,
    ) {
    }

    public function findById(OrderId $id): ?\App\Domain\Order\Order
    {
        $record = $this->model->find($id->value());
        return $record?->toDomain();
    }

    public function save(\App\Domain\Order\Order $order): void
    {
        $this->model->fromDomain($order)->save();
    }
}
```

---

## Ports and Adapters (Hexagonal)

Toda interação externa é um **port** (interface) com um **adapter** (implementação):

| Port (interface) | Adapter (implementação) |
|---|---|
| `OrderRepository` | `DoctrineOrderRepository` |
| `PaymentGateway` | `StripePaymentAdapter` |
| `EventPublisher` | `RabbitMQEventAdapter` |
| `EmailSender` | `SmtpEmailAdapter` |
| `HttpClient` | `GuzzleHttpClientAdapter` |

### Input Ports (Use Cases)

Os use cases em si são **input ports** — o sistema expõe capacidades, não implementações.

```php
// Port de entrada (use case)
interface CreateOrderUseCase
{
    public function __invoke(CreateOrderRequest $request): OrderResponse;
}

// Implementação do use case
final class CreateOrder implements CreateOrderUseCase
{
    public function __construct(
        private readonly OrderRepository $orders,
        private readonly CustomerRepository $customers,
    ) {
    }

    public function __invoke(CreateOrderRequest $request): OrderResponse
    {
        // ...
    }
}
```

### Output Ports

Interfaces que a infraestrutura deve implementar para fornecer serviços ao domínio/application.

```php
// Output port
interface EventPublisher
{
    public function publish(object $event): void;
}

// Adapter
final class InMemoryEventPublisher implements EventPublisher
{
    private array $published = [];

    public function publish(object $event): void
    {
        $this->published[] = $event;
    }

    public function published(): array
    {
        return $this->published;
    }
}
```

---

## Estrutura de Pastas

```
src/
├── Domain/                    # Núcleo — zero dependências externas
│   ├── Order/
│   │   ├── Order.php          # Aggregate root
│   │   ├── OrderItem.php      # Entidade interna
│   │   ├── OrderId.php        # Value object
│   │   ├── OrderStatus.php    # Enum
│   │   ├── OrderRepository.php # Port (interface)
│   │   └── OrderConfirmed.php # Domain event
│   ├── Customer/
│   │   ├── Customer.php
│   │   └── CustomerId.php
│   └── Shared/
│       └── Money.php
│
├── Application/               # Use cases — orquestração
│   ├── Order/
│   │   ├── CreateOrder.php
│   │   ├── CreateOrderRequest.php
│   │   └── OrderResponse.php
│   └── Customer/
│       └── RegisterCustomer.php
│
├── Infrastructure/            # Drivers — frameworks, drivers, adapters
│   ├── Persistence/
│   │   ├── DoctrineOrderRepository.php
│   │   └── EloquentOrderRepository.php
│   ├── Http/
│   │   ├── OrderController.php
│   │   └── StripePaymentAdapter.php
│   ├── Event/
│   │   └── RabbitMQEventPublisher.php
│   └── Cli/
│       └── GenerateReportCommand.php
│
└── ...entry points (index.php, artisan, etc.)
```

### Regras de organização

- **Um arquivo = uma classe** (PSR-4).
- **Um aggregate por subdiretório** dentro de `Domain/`.
- **Use cases** em `Application/` — um por arquivo, nome no imperativo (`CreateOrder`, `CancelOrder`, `RegisterCustomer`).
- **Ports** (interfaces) ficam **dentro do aggregate** que define (`Domain/Order/OrderRepository.php`).
- **Adapters** ficam em `Infrastructure/` — agrupados por tipo (`Persistence/`, `Http/`, `Event/`).
- **Entry points** (controllers, commands) ficam em `Infrastructure/` — eles dependem dos use cases, nunca do domínio diretamente.

---

## DTOs e Mappers

### Request DTO (Input)

Objeto imutável que encapsula o input de um use case. Validação pode acontecer aqui ou no controller.

```php
final readonly class CreateOrderRequest
{
    public function __construct(
        public readonly string $customerId,
        /** @var array<array{productId: string, quantity: int, priceAmount: int, priceCurrency: string}> */
        public readonly array $items,
    ) {
        if (empty($this->items)) {
            throw new \InvalidArgumentException('Pedido deve ter ao menos um item.');
        }
    }
}
```

### Response DTO (Output)

Objeto imutável que encapsula o output do use case. Nunca retornar entidades do domínio diretamente para a presentation layer.

```php
final readonly class OrderResponse
{
    public function __construct(
        public readonly string $id,
        public readonly int $total,
        public readonly string $status,
    ) {
    }
}
```

### Mappers (Domain ↔ Model)

Separar a conversão entre domínio e modelo de persistência (ORM model, database row).

```php
namespace App\Infrastructure\Persistence\Mappers;

use App\Domain\Order\Order;
use App\Infrastructure\Persistence\Models\OrderModel;

final class OrderMapper
{
    public static function toDomain(OrderModel $model): Order
    {
        return Order::reconstitute(
            id: new OrderId($model->id),
            customerId: new CustomerId($model->customer_id),
            items: array_map(
                fn (OrderItemModel $item) => OrderItem::reconstitute(
                    productId: new ProductId($item->product_id),
                    quantity: $item->quantity,
                    price: new Money($item->price_amount, $item->price_currency),
                ),
                $model->items->toArray(),
            ),
            status: OrderStatus::from($model->status),
        );
    }

    public static function toModel(Order $order): OrderModel
    {
        return new OrderModel([
            'id' => $order->id()->value(),
            'customer_id' => $order->customerId()->value(),
            'status' => $order->status()->value,
        ]);
    }
}
```

---

## Testabilidade

A grande vantagem da Clean Architecture: testar a lógica de negócio **sem banco, sem HTTP, sem framework**.

### Teste de Use Case com mocks

```php
final class CreateOrderTest extends \PHPUnit\Framework\TestCase
{
    private InMemoryOrderRepository $orders;
    private CreateOrder $useCase;

    protected function setUp(): void
    {
        $this->orders = new InMemoryOrderRepository();
        $this->useCase = new CreateOrder($this->orders);
    }

    public function testCreatesOrderWithItems(): void
    {
        $result = ($this->useCase)(new CreateOrderRequest(
            customerId: 'cust-123',
            items: [
                ['productId' => 'prod-1', 'quantity' => 2, 'priceAmount' => 1000, 'priceCurrency' => 'BRL'],
            ],
        ));

        $this->assertSame('BRL', $result->currency);
        $this->assertCount(1, $this->orders->all());
    }

    public function testRejectsEmptyOrder(): void
    {
        $this->expectException(\InvalidArgumentException::class);

        ($this->useCase)(new CreateOrderRequest(
            customerId: 'cust-123',
            items: [],
        ));
    }
}
```

### Repositório in-memory para testes

```php
namespace App\Infrastructure\Persistence\InMemory;

use App\Domain\Order\Order;
use App\Domain\Order\OrderRepository;

final class InMemoryOrderRepository implements OrderRepository
{
    private array $orders = [];

    public function findById(OrderId $id): ?Order
    {
        return $this->orders[$id->value()] ?? null;
    }

    public function save(Order $order): void
    {
        $this->orders[$order->id()->value()] = $order;
    }

    public function all(): array
    {
        return $this->orders;
    }
}
```

---

## Estratégia de Migração (Código Legado)

1. **Identificar o domínio**: separar regras de negócio do framework/infraestrutura.
2. **Criar entities/value objects** para os conceitos centrais do domínio.
3. **Extrair use cases** do controller/service gigante — um por operação.
4. **Criar ports** para dependências externas (repositório, gateway, etc.).
5. **Implementar adapters** para o sistema atual (em vez de reescrever tudo de uma vez).
6. **Mover controllers** para chamar use cases em vez de services de infraestrutura.
7. **Iterar**: não reescrever tudo de uma vez — migrar componente por componente.

Regra prática: **não refactor de uma vez**. Mover código para a camada correta quando o arquivo já estiver sendo modificado por outro motivo.

---

## Checklist de Revisão de Clean Architecture

**Dependency Rule**
- [ ] O domínio não importa nada de `Infrastructure/`, `Application/` ou frameworks.
- [ ] A application importa apenas `Domain/`.
- [ ] Controllers delegam para use cases, nunca chamam repositórios diretamente.

**Ports e Adapters**
- [ ] Interfaces de persistência/HTTP/events vivem no `Domain/`, não em `Infrastructure/`.
- [ ] Implementações estão em `Infrastructure/`.
- [ ] É possível trocar driver (Doctrine ↔ Eloquent) sem alterar domínio.

**Use Cases**
- [ ] Um use case = uma operação (SRP).
- [ ] Use cases são fáceis de testar (sem banco, sem HTTP).
- [ ] Request/Response são DTOs imutáveis, não arrays associativos.

**Domain**
- [ ] Entidades e value objects são imutáveis (ou têm methods de comportamento, não setters públicos).
- [ ] Regras de negócio vivem no domínio, não no controller ou no repositório.
- [ ] Aggregate Root é o único ponto de acesso externo.

**Testes**
- [ ] Testes de domínio não dependem de banco, HTTP ou framework.
- [ ] Testes de use case usam repositories in-memory ou mocks.
- [ ] Cobertura de testes cobre invariantes do domínio (não só o CRUD).
