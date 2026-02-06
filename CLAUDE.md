# CLAUDE.md - Project Intelligence

## Project Overview

Go REST API (module: `github.com/marcelofabianov/course`) with OAuth2 authentication. Runs in Docker with hot-reload (Air), PostgreSQL 18, Redis 7, TLS 1.3, and Uber FX dependency injection.

## Tech Stack

- **Go 1.25** | **PostgreSQL 18** (pgx v5 driver) | **Redis 7** (go-redis v9)
- **Router:** Chi v5 | **DI:** Uber FX | **Migrations:** Goose v3
- **Validation:** go-playground/validator v10 + wisp (Brazilian validators)
- **Errors:** github.com/marcelofabianov/fault (structured error handling with codes)
- **Value Objects:** github.com/marcelofabianov/wisp (UUID, Email, Phone, CPF, CNPJ, CEP, Audit, Role)
- **Password Hashing:** Argon2 (doberman) | **Logging:** slog
- **Testing:** testify, testcontainers-go, miniredis
- **Linter:** golangci-lint | **Formatter:** gofumpt | **Security:** gosec, govulncheck

## Architecture

### Pilares de Desenvolvimento

**Security-first, TDD, DDD, Data Integrity, Pattern Consistency** - nesta ordem de prioridade.

- **Security-first** - toda decisao considera seguranca como requisito primario
- **TDD** - testes escritos ANTES da implementacao, cobertura minima de 80%
- **DDD** - modelagem orientada ao dominio, bounded contexts isolados, linguagem ubiqua
- **Data Integrity** - garantir integridade de dados em todas as camadas (constraints no banco, validacao no domain, optimistic locking, transacoes atomicas, idempotencia em event handlers)
- **Pattern Consistency** - manter padroes ja estabelecidos no projeto. Antes de implementar, verificar como o codigo existente resolve problemas similares e seguir a mesma abordagem (estrutura de arquivos, naming, error handling, DI registration, port definitions, audit fields)

### Principios

SOLID, Clean Code, DRY, KISS - sem over-engineering.

### Padroes Arquiteturais

- **Clean Architecture** - camadas com dependencias apontando para dentro (domain no centro)
- **Hexagonal Architecture (Ports & Adapters)** - ports definem contratos, adapters implementam
- **Event-Driven Architecture (EDA)** - domain events para comunicacao entre bounded contexts

### Directory Structure

```
cmd/api/main.go              # Entrypoint - FX bootstrap
config/                       # Configuracao centralizada (viper + .env)
internal/                     # Codigo privado da aplicacao
  di/                         # Modulos FX (PkgModule, AppModule, feature modules)
    app.go                    # Router, Server, HealthCheckers providers
    pkg.go                    # Config, Logger, Database, Cache, Validation providers
    {feature}.go              # Feature-specific DI module
  {feature}/                  # Bounded Context (ex: user/)
    domain/                   # Entidades, Value Objects, Erros de dominio, Events
      user.go                 # Entity com NewUser(), metodos de comportamento
      role.go                 # Value Object com Scan/Value para DB
      err.go                  # Sentinel errors + factory functions com fault.Wrap
      event.go                # Domain Events
    port/                     # Interfaces (contratos)
      repository.go           # Repository ports (composicao de interfaces granulares)
      usecase.go              # UseCase ports (Input/Output + interface Execute)
    usecase/                  # Implementacao dos use cases
    handler/                  # HTTP handlers (adapter primario)
    storage/                  # Repository implementations (adapter secundario)
pkg/                          # Pacotes reutilizaveis (library code)
  logger/                     # Wrapper slog
  retry/                      # Strategy pattern (Exponential, Constant, Linear backoff)
  validation/                 # Wrapper go-playground/validator com sanitizacao
  database/                   # PostgreSQL wrapper com retry, pool, health check
  cache/                      # Redis wrapper com retry, pool, operacoes CRUD
  web/                        # HTTP layer
    server.go                 # TLS server
    router.go                 # Router interface
    response.go               # JSON response helpers (Success, Error, Created, etc.)
    context.go                # Context keys (Logger, RequestID)
    health.go                 # Liveness/Readiness handlers
    chi/router.go             # Chi router factory com middleware stack
    middleware/                # 15 middlewares (security, logging, rate limit, etc.)
_env/dev/                     # Dev environment configs (Dockerfile, docker-compose, .env)
db/migrations/                # Goose SQL migrations
db/fixtures/                  # Dev data
db/seeds/                     # Seed data
```

### DI Pattern (Uber FX)

- `PkgModule` fornece: Config, Logger, Database, Cache, Validation
- `AppModule` fornece: Router, Server, HealthCheckers
- Feature modules fornecem: Handlers (como `web.Router` via `AsRouter()`), UseCases, Repositories
- Registrar routers com: `fx.Provide(AsRouter(NewFeatureRouter))`
- Registrar health checkers com: `fx.Provide(AsHealthChecker(NewChecker))`

### Error Handling Pattern (fault)

```go
// Sentinel errors no topo do arquivo
var ErrSomething = fault.New("message", fault.WithCode(fault.Invalid))

// Factory functions para erros contextualizados
func NewErrSomething(detail string) error {
    return fault.Wrap(ErrSomething, "description",
        fault.WithCode(fault.DomainViolation),
        fault.WithContext("key", detail),
        fault.WithContext("aggregate", AGGREGATE_NAME),
    )
}
```

Codigos fault disponiveis: `fault.Invalid`, `fault.NotFound`, `fault.Conflict`, `fault.Internal`, `fault.InfraError`, `fault.DomainViolation`

### Domain Entity Pattern

- Entities usam value objects do `wisp` (UUID, Email, Phone, NonEmptyString, etc.)
- Constructor `NewEntity(input, ...)` valida e retorna `(*Entity, error)`
- Metodos de comportamento no entity (Update, ChangePassword, Activate, Deactivate, Delete)
- `wisp.Audit` embedado para CreatedAt, UpdatedAt, DeletedAt, ArchivedAt, CreatedBy, UpdatedBy
- Roles implementam `driver.Valuer` e `sql.Scanner` para persistencia

### Port Pattern (Hexagonal)

```go
// Interfaces granulares compostas
type CreateUserRepositoryPort interface {
    CreateUser(ctx context.Context, user *domain.User) error
}
type UserRepositoryPort interface {
    CreateUserRepositoryPort
    // ... compose more granular ports
}

// UseCase port com Input/Output structs
type RegisterUserUseCase interface {
    Execute(ctx context.Context, input *RegisterUserInput) (*RegisterUserOutput, error)
}
```

## Rules

### Security Rules (OBRIGATORIO)

1. **NUNCA** commitar secrets, passwords, API keys, ou certificados. Arquivos `.env`, `certs/`, `*.pem`, `*.key` estao no `.gitignore`
2. **SEMPRE** usar parametrized queries ($1, $2) para SQL - NUNCA concatenar strings em queries
3. **SEMPRE** validar e sanitizar input do usuario antes de processar
4. **SEMPRE** usar `context.Context` com timeout em operacoes de I/O (DB, Redis, HTTP)
5. **SEMPRE** usar `fault.Wrap` com codigo apropriado - NUNCA expor erros internos ao cliente
6. **NUNCA** logar dados sensiveis (password, token, secret, credit_card) - o validator ja sanitiza
7. **SEMPRE** usar Argon2 para hash de senhas - NUNCA MD5, SHA1, SHA256 ou bcrypt
8. **SEMPRE** manter TLS 1.2+ com cipher suites recomendadas (ja configurado em `web/server.go`)
9. **SEMPRE** usar cookies HttpOnly, Secure, SameSite=Strict para tokens CSRF
10. **SEMPRE** definir rate limiting em endpoints publicos
11. **SEMPRE** usar `crypto/rand` para geracao de tokens - NUNCA `math/rand` para criptografia
12. **SEMPRE** validar JWT com secrets de pelo menos 32 bytes
13. **NUNCA** desabilitar security headers em producao

### Quality Rules (OBRIGATORIO)

1. **TDD** - Escrever testes ANTES da implementacao. Todo codigo novo DEVE ter testes
2. **Table-driven tests** - Usar pattern de table-driven tests com `t.Run()` e subtests
3. **Test naming** - Nomes descritivos: `TestSubject_Scenario` ou `TestSubject_Method_Scenario`
4. **Coverage** - Manter cobertura de testes acima de 80%
5. **Interfaces** - Definir interfaces no pacote consumidor (port), NAO no pacote que implementa
6. **Erros como valores** - Usar sentinel errors com `errors.Is()` e `fault.Wrap()` para contexto
7. **Happy path left-aligned** - Tratar erros primeiro, happy path sem indentacao
8. **Early return** - Retornar cedo em condicoes de erro, evitar else desnecessario
9. **Naming** - Usar nomes claros e descritivos. Sem abreviacoes obscuras
10. **Package naming** - Singular, lowercase, sem underscores (ex: `handler`, `domain`, `usecase`)
11. **Formatacao** - Usar `gofumpt` (superset do gofmt). Rodar antes de commitar
12. **Linting** - Codigo deve passar `golangci-lint run ./...` sem erros
13. **Nao adicionar** comentarios obvios, docstrings em codigo que nao foi alterado, ou type annotations desnecessarias
14. **Example tests** - Adicionar `Example_` tests para documentacao de API publica em `pkg/`

### Performance Rules

1. **Connection pooling** - SEMPRE usar pool para DB e Redis (ja configurado)
2. **Context com timeout** - NUNCA fazer I/O sem timeout via context
3. **sync.Pool** - Usar para objetos frequentemente alocados (ex: gzip writers)
4. **Goroutine safety** - Proteger estado compartilhado com `sync.Mutex` ou `sync.RWMutex`
5. **Retry com backoff** - Usar `pkg/retry` com exponential backoff e jitter para operacoes de rede
6. **Circuit breaker** - Usar gobreaker para proteger contra cascading failures (ex: rate limiter)
7. **Benchmark tests** - Adicionar `Benchmark` tests para hot paths
8. **Evitar alocacoes** - Preferir `[]byte` sobre `string` em hot paths, pre-alocar slices com `make(s, 0, cap)`

### Architecture Rules

1. **Dependency Rule** - Dependencias SEMPRE apontam para dentro: handler -> usecase -> domain. NUNCA ao contrario
2. **Domain puro** - O pacote `domain/` NAO importa nada de `pkg/`, `config/`, ou frameworks. Apenas `wisp` e `fault` como excecoes aceitas
3. **Ports no consumidor** - Interfaces (ports) ficam em `port/`, definidas pelo lado que CONSOME
4. **Um usecase por arquivo** - Cada use case em seu proprio arquivo com Input/Output
5. **Handler fino** - Handlers apenas: parse request -> call usecase -> write response. ZERO logica de negocio
6. **Repository como adapter** - Implementacoes de repository em `storage/` implementam ports de `port/`
7. **DI via FX** - TODA injecao de dependencia via Uber FX modules. NUNCA `init()` para setup global (exceto `wisp.RegisterRoles`)
8. **Feature isolation** - Cada bounded context em `internal/{feature}/` com seu proprio domain, ports, usecases, handlers, storage
9. **Events para cross-context** - Comunicacao entre bounded contexts via domain events (EDA), NUNCA importar domain de outro context diretamente

### CRITICAL: Isolamento entre Bounded Contexts

**Um contexto NUNCA pode depender diretamente de outro contexto.** Bounded contexts DEVEM ser completamente desacoplados e independentes.

**Como funciona:**
- Cada contexto (`internal/{feature}/`) define seus proprios **Ports** em `port/` para qualquer dependencia externa que precise, incluindo servicos de outros contextos
- Se `order` precisa consultar dados de `user`, o contexto `order` define um port (ex: `port.UserQueryPort`) com a interface que ele precisa
- A implementacao concreta desse port (adapter) que conecta ao outro contexto e registrada EXCLUSIVAMENTE em `internal/di/`
- O modulo DI (`internal/di/{feature}.go`) e o **UNICO local autorizado** a fazer cross-context wiring, conectando ports de um contexto a implementacoes de outro

**Regras:**
- **PROIBIDO** importar pacotes de `internal/{outro_contexto}/` dentro de qualquer arquivo de um bounded context
- **PROIBIDO** references diretas entre contextos em domain, usecase, handler, ou storage
- **PERMITIDO** apenas em `internal/di/` fazer o wiring cross-context via Ports & Adapters
- Cada contexto deve funcionar de forma autonoma - se removermos outro contexto, o codigo do contexto atual DEVE compilar (apenas o DI quebraria)

**Exemplo de cross-context via Ports & Adapters (sincrono):**
```
internal/order/port/user_query.go     -> define UserQueryPort interface
internal/di/order.go                  -> registra adapter que implementa UserQueryPort usando user storage
internal/order/usecase/create_order.go -> recebe UserQueryPort via constructor injection
```

### CRITICAL: Event-Driven Architecture (EDA) entre Contextos

**EDA e a forma PREFERIDA de interacao entre bounded contexts.** Sempre que possivel, usar domain events ao inves de chamadas sincronas entre contextos.

**Principio:** Cada contexto emite eventos sobre o que aconteceu no seu dominio. Outros contextos reagem a esses eventos de forma autonoma, sem conhecer o emissor.

**Como funciona:**
- Cada contexto define seus **Domain Events** em `domain/event.go` - structs imutaveis que representam fatos que aconteceram
- O contexto que emite o evento NAO sabe e NAO se importa com quem consome
- O contexto que consome define um **Port** para o event handler e reage ao evento com sua propria logica
- O wiring entre evento e handler e feito EXCLUSIVAMENTE em `internal/di/`

**Regras EDA:**
- **Events sao fatos imutaveis** - nomeados no passado (ex: `UserRegistered`, `OrderCreated`, `PaymentCompleted`)
- **Fire-and-forget** - o emissor publica o evento e segue. NAO espera resposta dos consumidores
- **Idempotencia** - handlers de eventos DEVEM ser idempotentes (processar o mesmo evento 2x nao causa efeito colateral)
- **Autonomia** - se o consumidor falhar ao processar o evento, o emissor NAO e afetado
- **Sem acoplamento de dados** - o evento carrega apenas os dados minimos necessarios (IDs e dados essenciais), NAO objetos de dominio de outro contexto
- **Event bus via Port** - o mecanismo de publicacao/subscricao e definido como Port, permitindo trocar implementacao (in-memory, Redis Pub/Sub, Kafka, etc.)
- **Ordering** - NAO depender de ordem de entrega dos eventos. Handlers devem funcionar independente da ordem

**Quando usar EDA vs Ports & Adapters sincrono:**
- **EDA (preferido):** notificacoes, side effects, atualizacao de read models, propagacao de estado, workflows assincronos
- **Sincrono via Port:** quando o use case PRECISA do resultado imediato para continuar (ex: validar se usuario existe antes de criar pedido)

**Exemplo de cross-context via EDA:**
```
internal/user/domain/event.go          -> define UserRegistered event struct
internal/user/usecase/register_user.go -> publica UserRegistered via EventPublisherPort
internal/notification/port/event.go    -> define handler port para UserRegistered
internal/notification/usecase/...      -> reage ao evento enviando welcome email
internal/di/notification.go            -> registra subscription do handler ao evento
```

### Middleware Stack (ordem importa)

1. Recovery -> 2. RequestID -> 3. RealIP -> 4. Logger -> 5. SecurityHeaders -> 6. HTTPSOnly -> 7. CORS -> 8. RequestSize -> 9. Compression -> 10. RateLimit -> 11. Heartbeat

Dentro de `/api/v1`: Timeout -> AcceptJSON -> AllowContentType -> CSRF -> Routes

## Skills Usage Rules (OBRIGATORIO)

SEMPRE priorizar o uso de skills disponiveis. Invocar a skill ANTES de gerar qualquer resposta sobre o tema.

### Quando usar cada skill:

- **`golang-style`** - OBRIGATORIO antes de escrever ou editar qualquer arquivo `.go`. Garante happy path coding, error wrapping, sentinel errors, godoc comments
- **`golang-pro`** - Usar quando trabalhar com goroutines, channels, generics, gRPC, ou sistemas de alta performance
- **`golang-patterns`** - Usar para aplicar patterns idiomaticos Go, best practices, e convencoes
- **`golang-testing`** - Usar quando escrever testes. Garante table-driven tests, subtests, benchmarks, fuzzing, TDD
- **`software-architecture`** - Usar quando discutir ou implementar decisoes arquiteturais, analise de codigo, ou design de sistemas
- **`database-architect`** - PROATIVAMENTE usar para decisoes de modelagem de dados, schema, migrations, ou selecao de tecnologia de dados
- **`microservices-architect`** - Usar para design de sistemas distribuidos, service boundaries, saga patterns, event sourcing
- **`microservices-patterns`** - Usar para implementar patterns de microservicos, comunicacao event-driven, resiliencia
- **`feature-dev:feature-dev`** - Usar para desenvolvimento guiado de features com foco em arquitetura e entendimento do codebase
- **`find-skills`** - Usar quando o usuario busca funcionalidade que pode existir como skill instalavel

### Combinacoes comuns:

- **Nova feature**: `software-architecture` + `golang-style` + `golang-testing`
- **Novo endpoint**: `golang-style` + `golang-testing` + `feature-dev:feature-dev`
- **Schema/migration**: `database-architect` + `golang-style`
- **Bug fix**: `golang-style` + `golang-testing`
- **Performance**: `golang-pro` + `golang-patterns`

## Commands Reference

- `make setup-dev` - Setup completo do ambiente
- `make up` / `make down` - Start/stop containers
- `make test` - Rodar testes
- `make lint` - Rodar linter
- `make fmt` - Formatar codigo
- `make security-check` - gosec + govulncheck
- `make migrate-create NAME=x` - Criar migration
- `make migrate-up` / `make migrate-down` - Executar/reverter migrations
- `make shell` - Shell no container API

## Language

Responder SEMPRE em Portugues Brasil. Termos tecnicos e identificadores de codigo permanecem em ingles.
