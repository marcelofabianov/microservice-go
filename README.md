# Boilerplate

## Tecnologias

**Core**

- Linguagem: Go 1.25
- Banco de Dados: PostgreSQL 18
- Cache: Redis 7
- Roteador HTTP: Chi v5
- Injeção de Dependência: Uber FX
- Acesso a Dados: pgx v5 + sqlx
- Migrações: Goose v3

**Segurança**

- TLS: TLS 1.3 exclusivo com cipher suites recomendadas
- Cookies Seguros: HttpOnly, Secure, SameSite=Lax
- Hash de Senha: Argon2 (doberman)
- Validação: go-playground/validator v10

**Desenvolvimento**

- Hot Reload: Air
- Debugger: Delve
- Logging: Slog
- Testes: Testify, Testcontainers-Go
- Linter: golangci-lint
- Formatter: gofumpt
- Security Scanner: gosec
- API Documentation: Swagger (swaggo/swag)

## 1. Ambiente

**Pré-requisitos**

- [Docker](https://docs.docker.com/get-docker/)
- [mkcert](https://github.com/FiloSottile/mkcert)
- [Make](https://www.make.sh/)
- [git](https://git-scm.com/downloads)

## 2. Instruções de Configuração

...
