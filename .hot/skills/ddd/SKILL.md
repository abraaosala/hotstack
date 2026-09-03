---
name: ddd
description: Domain-Driven Design (DDD) — modelagem de domínio, Entities, Value Objects, Aggregates, Repositories, Domain Events, Domain Services, Bounded Contexts, Ubiquitous Language, Anti-Corruption Layers e Strategic Design. Use esta skill sempre que o usuário pedir para modelar domínio, criar entidades/值 objects, definir agregados, trabalhar com domain events, estruturar bounded contexts, usar repository pattern, implementar domain services, ou quando estiver projetando a camada de domínio de uma aplicação — mesmo que não mencione "DDD" ou "modelagem de domínio" explicitamente. Também use ao revisar código que mistura lógica de negócio com infraestrutura, ou quando o usuário perguntar sobre "separar responsabilidades", "camadas de domínio", "regras de negócio", ou "acoplamento entre módulos".
---

# Domain-Driven Design (DDD)

Guia prático de Domain-Driven Design baseado nos conceitos de Eric Evans, Vaughn Vernon e/eventuais recursos modernos de modelagem de domínio.

## Quando usar cada seção

- Vai criar uma entidade, value object ou aggregate? → **Building Blocks**
- Vai definir limites entre módulos/subdomínios? → **Bounded Contexts**
- Vai nomear conceitos do domínio com precisão? → **Ubiquitous Language**
- Vai implementar persistência sem acoplar domínio? → **Repositories**
- Vai lidar com comunicação assíncrona entre contextos? → **Domain Events**
- Vai integrar com um contexto externo sem expor seu domínio? → **Anti-Corruption Layer**
- Vai decidir onde colocar lógica que não pertence a uma entidade? → **Domain Services**
- Vai revisar se o código realmente modela o domínio? → **Checklist de Revisão**

---

## Ubiquitous Language

A linguagem usada por desenvolvedores e especialistas do domínio deve ser **a mesma**. Nomes de classes, métodos, variáveis e até commits devem usar o vocabulário do domínio, não termos técnicos genéricos.

- Usar `Payment` em vez de `DataObject`.
- Usar `Invoice.issue()` em vez de `Invoice.setStatus(1)`.
- Evitar abreviações ambíguas (`usr`, `mgr`) — legibilidade > brevidade.
- Quando um termo tiver múltiplos significados no domínio, resolver a ambiguidade com o especialista e usar o nome mais preciso.

---

## Building Blocks

### Entities (Entidades)

Objetos com **identidade única** que persiste ao longo do tempo. Duas entidades com o mesmo conteúdo são entidades diferentes se tiverem IDs diferentes.

```php
final class Order
{
    public function __construct(
        public readonly OrderId $id,
        private readonly CustomerId $customerId,
        private array $items = [],
        private OrderStatus $status = OrderStatus::Pending,
    ) {
    }

    public function addItem(ProductId $productId, int $quantity, Money $price): void
    {
        $this->items[] = new OrderItem($productId, $quantity, $price);
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

Regras:
- Entities encapsulam **regras de negócio**, não apenas dados.
- Mudanças de estado acontecem via **métodos de comportamento**, não setters públicos.
- IDs devem ser value objects (não tipos primitivos como `string` ou `int`).

### Value Objects (Objetos de Valor)

Objetos **imutáveis** identificados pelo seu **valor**, não por uma identidade. Dois value objects com o mesmo conteúdo são considerados iguais.

```php
final readonly class Money
{
    public function __construct(
        private int $amount,
        private string $currency,
    ) {
        if ($amount < 0) {
            throw new \InvalidArgumentException('Valor não pode ser negativo.');
        }
    }

    public function add(Money $other): self
    {
        $this->assertSameCurrency($other);
        return new self($this->amount + $other->amount, $this->currency);
    }

    public function equals(Money $other): bool
    {
        return $this->amount === $other->amount
            && $this->currency === $other->currency;
    }

    public static function zero(string $currency): self
    {
        return new self(0, $currency);
    }
}
```

Regras:
- **Sempre imutáveis** (`readonly` ou equivalente).
- Comparação por valor, não por referência (`equals()`, não `===`).
- Validar invariantes no construtor — um value object inválido não deveria existir.
- Não expor setters públicos.

### Aggregates (Agregados)

Grupo de entidades e value objects tratados como **unidade para persistência e改变 de estado**. Uma sólida Aggregate Root (raiz do agregado) é o ponto de entrada para todas as改变.

```php
final class Order // ← Order é a Aggregate Root
{
    // ...

    public function addItem(ProductId $productId, int $quantity, Money $price): void
    {
        if ($this->status !== OrderStatus::Pending) {
            throw new \DomainException('Não é possível adicionar itens a um pedido já processado.');
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
}
```

Regras:
- Só a **Aggregate Root** pode ser acessada diretamente por outsiders.
- Entidades internas (ex.: `OrderItem`) são acessadas apenas via methods da Root.
- Uma transação deve modificar **um só aggregate** por vez.
- Referências entre aggregates devem ser por **ID** (não por referência direta a outro aggregate).

```php
// ✅ Correto: referência por ID
final class Order
{
    public function __construct(
        private readonly OrderId $id,
        private readonly CustomerId $customerId, // referencia outro contexto por ID
    ) {
    }
}

// ❌ Errado: referência direta a outro aggregate
final class Order
{
    public function __construct(
        private readonly OrderId $id,
        private Customer $customer, // acoplamento direto — errado
    ) {
    }
}
```

---

## Domain Events (Eventos de Domínio)

Fatos que **já aconteceram** no domínio, usados para comunicar改变 entre agregados e contextos de forma **desacoplada**.

```php
final readonly class OrderConfirmed
{
    public function __construct(
        public readonly OrderId $orderId,
        public readonly CustomerId $customerId,
        public readonly Money $total,
        public readonly \DateTimeImmutable $occurredAt,
    ) {
    }
}
```

Regras:
- Eventos são **passado** (facts), não comandos — nome no **pretérito** (`OrderConfirmed`, não `ConfirmOrder`).
- São **imutáveis** (readonly/valores).
- Disparados **dentro** do aggregate, publicados **depois** da persistência (para evitar efeitos colaterais se a transação falhar).
- Evitar eventos que contenham referências a objetos complexos — preferir IDs.

```php
final class Order
{
    private array $domainEvents = [];

    public function confirm(): void
    {
        if (empty($this->items)) {
            throw new \DomainException('Pedido vazio.');
        }
        $this->status = OrderStatus::Confirmed;
        $this->domainEvents[] = new OrderConfirmed(
            orderId: $this->id,
            customerId: $this->customerId,
            total: $this->total(),
            occurredAt: new \DateTimeImmutable(),
        );
    }

    public function pullDomainEvents(): array
    {
        $events = $this->domainEvents;
        $this->domainEvents = [];
        return $events;
    }
}
```

---

## Repositories

Abstração para **persistência** do aggregate. A interface pertence ao domínio; a implementação pertence à infraestrutura.

```php
// Domínio (interface)
namespace App\Domain\Order;

interface OrderRepository
{
    public function findById(OrderId $id): ?Order;
    public function save(Order $order): void;
}

// Infraestrutura (implementação)
namespace App\Infrastructure\Persistence;

final class DoctrineOrderRepository implements \App\Domain\Order\OrderRepository
{
    public function __construct(
        private readonly \Doctrine\ORM\EntityManagerInterface $em,
    ) {
    }

    public function findById(OrderId $id): ?\App\Domain\Order\Order
    {
        return $this->em->find(\App\Domain\Order\Order::class, $id->value());
    }

    public function save(\App\Domain\Order\Order $order): void
    {
        $this->em->persist($order);
        $this->em->flush();
    }
}
```

Regras:
- A interface vive no **domínio**; a implementação na **infraestrutura**.
- O domínio nunca depende de Doctrine, Eloquent, MongoDB, etc.
- Repositories retornam **aggregates completos** (não objects parciais).
- Queries complexas (relatórios, buscas por múltiplos campos) podem usar **Query Services** separados, não o Repository.

---

## Domain Services (Serviços de Domínio)

Lógica de negócio que **não pertence naturalmente a uma entidade ou value object** — tipicamente envolve múltiplos aggregates ou operações cross-domain.

```php
final readonly class PricingService
{
    public function calculateTotal(Order $order, Customer $customer): Money
    {
        $subtotal = $order->total();
        $discount = $customer->isVip()
            ? $subtotal->percentage(10)
            : Money::zero($subtotal->currency());

        return $subtotal->subtract($discount);
    }
}
```

Regras:
- Domain Services são **stateless** (sem estado).
- Não devem conter lógica de infraestrutura (database, HTTP, queues).
- Devem ser **fáceis de testar** — apenas dependem de objetos do domínio.
- Se a lógica envolve apenas um aggregate, considere movê-la para o próprio aggregate (método de comportamento).

---

## Bounded Contexts (Contextos Delimitados)

Cada subdomínio/team feature tem seu **próprio modelo**, linguagem e limites claros. O mesmo conceito pode ter significados diferentes em contextos diferentes.

```
┌─────────────────────┐  ┌─────────────────────┐
│   Order Context      │  │   Shipping Context   │
│                      │  │                      │
│  Order               │  │  Shipment            │
│  ├─ OrderItem        │  │  ├─ TrackingEvent    │
│  ├─ Money            │  │  ├─ Address          │
│  └─ CustomerId ──────┼──┼──▶ Customer (cópia)  │
│                      │  │                      │
└─────────────────────┘  └─────────────────────┘
```

Regras:
- Cada context tem seu **próprio modelo** — não partilhar entidades entre contextos.
- Comunicação entre contextos: **Domain Events** (async) ou **API calls** (sync), nunca acesso direto a banco de dados de outro contexto.
- Cada context pode ter seu próprio banco de dados (ou schema separado).
- Evitar "God Models" que representam tudo em um só lugar.

---

## Anti-Corruption Layer (ACL)

Camada de tradução entre o teu domínio e um contexto externo (API legada, sistema de terceiros, banco de dados antigo). Protege o teu modelo de contaminação por conceitos de fora.

```php
// O sistema legado usa "USR_ID", "USR_NM" — nomes terríveis
final class LegacyUserAdapter
{
    public function __construct(
        private readonly LegacyUserGateway $gateway,
    ) {
    }

    public function findById(UserId $id): ?User
    {
        $raw = $this->gateway->fetchUser($id->value());
        if ($raw === null) {
            return null;
        }

        return new User(
            id: new UserId($raw['USR_ID']),
            name: $raw['USR_NM'],
            email: $raw['USR_EM'],
        );
    }
}
```

Regras:
- O domínio **nunca** conhece os termos/objetos do sistema externo.
- A ACL traduz **na entrada e na saída**.
- Ideal para migrações e integrações com sistemas legados.

---

## Checklist de Revisão de Código DDD

Ao revisar código de domínio, verificar:

**Entidades e Value Objects**
- [ ] Entidades têm identidade clara (não comparar por conteúdo).
- [ ] Value objects são imutáveis e comparados por valor.
- [ ] Invariantes são validados no construtor (value objects) ou em methods de comportamento (entidades).
- [ ] Nomes usam Ubiquitous Language (não termos técnicos genéricos).

**Aggregates**
- [ ] Aggregate Root é o único ponto de acesso externo.
- [ ] Entidades internas são acessíveis apenas via methods da Root.
- [ ] Referências entre aggregates são por ID, não por referência direta.
- [ ] Uma transação modific apenas um aggregate.

**Domain Events**
- [ ] Eventos são fatos (passado), não comandos.
- [ ] São imutáveis e contêm apenas IDs (não objetos complexos).
- [ ] Disparados dentro do aggregate, publicados depois da persistência.

**Repositories**
- [ ] Interface vive no domínio, implementação na infraestrutura.
- [ ] O domínio não depende de ORM, drivers de banco, etc.
- [ ] Retorna aggregates completos.

**Domain Services**
- [ ] São stateless.
- [ ] Não contêm lógica de infraestrutura.
- [ ] Operam com múltiplos aggregates (se operasse com um só, seria método da entidade).

**Bounded Contexts**
- [ ] Cada context tem seu próprio modelo — não partilhar entidades entre contextos.
- [ ] Comunicação cross-context é por eventos ou APIs, não acesso direto a banco.
- [ ] ACL é usada ao integrar com sistemas externos.
