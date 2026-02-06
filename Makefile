.PHONY: help setup-dev clean-setup build up down restart logs clean reset-db test lint security-scan \
	build-fast up-tools dev down-volumes restart-api logs-api logs-postgres logs-redis ps \
	db-migrate-up db-migrate-down db-migrate-status db-shell shell shell-root test-coverage \
	lint-fix lint-verbose lint-new fmt security-check mod-download mod-tidy mod-verify \
	quality-check quality-full ci build-prod build-prod-debug size-check \
	hooks-install hooks-uninstall hooks-list clean-all prune check-deps info swagger-init swagger-generate \
	compile compile-local compile-docker compile-container clean-binary run-binary

# =============================================================================
# VARIÁVEIS
# =============================================================================
PROJECT_NAME := course
COMPOSE_FILE := docker-compose.yml
ENV_FILE := .env
CERTS_DIR := certs
BIN_DIR := bin
TMP_DIR := tmp
DOCS_DIR := docs

# Binary build variables
BINARY_NAME := $(PROJECT_NAME)
MAIN_PATH := ./cmd/api/main.go
BUILD_DIR := ./$(BIN_DIR)
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags='-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)'
LDFLAGS_DOCKER := -ldflags='-s -w'

# Docker Compose command (suporta v1 e v2)
DOCKER_COMPOSE := $(shell command -v $(DOCKER_COMPOSE) 2>/dev/null || echo "docker compose")

# Cores para output (usando printf para compatibilidade)
COLOR_RESET := \033[0m
COLOR_BOLD := \033[1m
COLOR_GREEN := \033[32m
COLOR_YELLOW := \033[33m
COLOR_BLUE := \033[34m

# Helper para imprimir com cores
define print_color
	@printf "$(1)\n"
endef

# =============================================================================
# HELP - Lista todos os comandos disponíveis
# =============================================================================
help: ## Mostra esta mensagem de ajuda
	@printf "$(COLOR_BOLD)$(PROJECT_NAME) - Comandos Disponíveis$(COLOR_RESET)\n"
	@printf "\n"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(COLOR_GREEN)%-20s$(COLOR_RESET) %s\n", $$1, $$2}'
	@printf "\n"

# =============================================================================
# SETUP - Configuração inicial do ambiente
# =============================================================================
clean-setup: ## Remove configurações antigas da raiz
	@printf "$(COLOR_BLUE)→ Removendo configuração antiga da raiz...$(COLOR_RESET)\n"
	@rm -f .dockerignore docker-compose.yml Dockerfile .air.toml .air_dlv.toml
	@rm -rf $(BIN_DIR) $(CERTS_DIR) $(DOCS_DIR) $(TMP_DIR)
	@printf "$(COLOR_GREEN)✓ Configuração antiga removida!$(COLOR_RESET)\n"

setup-dev: clean-setup ## Configura ambiente de desenvolvimento completo
	@printf "$(COLOR_BLUE)→ Configurando ambiente de desenvolvimento...$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)1. Copiando arquivos de configuração...$(COLOR_RESET)\n"
	@cp -f _env/dev/.dockerignore .dockerignore
	@cp -f _env/dev/.env.example .env
	@cp -f _env/dev/docker-compose.dev.yml docker-compose.yml
	@cp -f _env/dev/Dockerfile Dockerfile
	@cp -f _env/dev/.air.toml .air.toml
	@cp -f _env/dev/.air_dlv.toml .air_dlv.toml
	@printf "$(COLOR_YELLOW)2. Configurando UID/GID do usuário no .env...$(COLOR_RESET)\n"
	@USER_UID=$$(id -u); \
	USER_GID=$$(id -g); \
	sed -i "s/^HOST_UID=.*/HOST_UID=$$USER_UID/" .env; \
	sed -i "s/^HOST_GID=.*/HOST_GID=$$USER_GID/" .env; \
	printf "  $(COLOR_GREEN)✓ HOST_UID=$$USER_UID$(COLOR_RESET)\n"; \
	printf "  $(COLOR_GREEN)✓ HOST_GID=$$USER_GID$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)3. Criando diretórios necessários...$(COLOR_RESET)\n"
	@mkdir -p $(BIN_DIR) $(CERTS_DIR) $(DOCS_DIR) $(TMP_DIR)
	@printf "$(COLOR_YELLOW)4. Gerando certificados TLS...$(COLOR_RESET)\n"
	@command -v mkcert >/dev/null 2>&1 || { printf "$(COLOR_YELLOW)⚠ mkcert não encontrado. Instale: https://github.com/FiloSottile/mkcert$(COLOR_RESET)\n"; exit 1; }
	@mkcert -key-file ./$(CERTS_DIR)/course-key.pem -cert-file ./$(CERTS_DIR)/course-cert.pem api.course.local localhost 127.0.0.1 ::1
	@printf "$(COLOR_YELLOW)5. Configurando permissões...$(COLOR_RESET)\n"
	@chmod 644 ./$(CERTS_DIR)/*.pem
	@printf "$(COLOR_YELLOW)6. Instalando Git hooks...$(COLOR_RESET)\n"
	@./.githooks/setup.sh install
	@printf "$(COLOR_YELLOW)7. Inicializando Swagger...$(COLOR_RESET)\n"
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g cmd/api/main.go -o ./$(DOCS_DIR) --parseDependency --parseInternal; \
		printf "  $(COLOR_GREEN)✓ Swagger inicializado em ./$(DOCS_DIR)$(COLOR_RESET)\n"; \
	else \
		printf "  $(COLOR_YELLOW)⚠ swag não encontrado localmente (será gerado no container)$(COLOR_RESET)\n"; \
		printf "  $(COLOR_BLUE)→ Use 'make swagger-init' após 'make up' para gerar docs$(COLOR_RESET)\n"; \
	fi
	@printf "$(COLOR_GREEN)✓ Ambiente configurado com sucesso!$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)→ Execute 'make build' para construir as imagens$(COLOR_RESET)\n"

# =============================================================================
# DOCKER - Comandos de gerenciamento de containers
# =============================================================================
build: ## Constrói as imagens Docker
	@printf "$(COLOR_BLUE)→ Construindo imagens Docker...$(COLOR_RESET)\n"
	@docker compose build --no-cache
	@printf "$(COLOR_GREEN)✓ Imagens construídas com sucesso!$(COLOR_RESET)\n"

build-fast: ## Constrói as imagens Docker usando cache
	@printf "$(COLOR_BLUE)→ Construindo imagens Docker (com cache)...$(COLOR_RESET)\n"
	@docker compose build
	@printf "$(COLOR_GREEN)✓ Imagens construídas com sucesso!$(COLOR_RESET)\n"

up: ## Inicia todos os containers em background
	@printf "$(COLOR_BLUE)→ Iniciando containers...$(COLOR_RESET)\n"
	@docker compose up -d
	@printf "$(COLOR_GREEN)✓ Containers iniciados!$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)→ Use 'make logs' para ver os logs$(COLOR_RESET)\n"

up-tools: ## Inicia containers incluindo ferramentas (adminer, redis-commander)
	@printf "$(COLOR_BLUE)→ Iniciando containers com ferramentas...$(COLOR_RESET)\n"
	@docker compose --profile tools up -d
	@printf "$(COLOR_GREEN)✓ Containers iniciados!$(COLOR_RESET)\n"

dev: ## Inicia ambiente de desenvolvimento e mostra logs
	@printf "$(COLOR_BLUE)→ Iniciando ambiente de desenvolvimento...$(COLOR_RESET)\n"
	@docker compose up

down: ## Para e remove todos os containers
	@printf "$(COLOR_BLUE)→ Parando containers...$(COLOR_RESET)\n"
	@docker compose down
	@printf "$(COLOR_GREEN)✓ Containers parados!$(COLOR_RESET)\n"

down-volumes: ## Para containers e remove volumes (CUIDADO: apaga dados)
	@printf "$(COLOR_YELLOW)⚠ ATENÇÃO: Isso irá remover todos os dados dos volumes!$(COLOR_RESET)\n"
	@read -p "Tem certeza? [y/N] " -n 1 -r; \
	printf "\n"; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker compose down -v; \
		printf "$(COLOR_GREEN)✓ Containers e volumes removidos!$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_BLUE)→ Operação cancelada$(COLOR_RESET)\n"; \
	fi

restart: ## Reinicia todos os containers
	@printf "$(COLOR_BLUE)→ Reiniciando containers...$(COLOR_RESET)\n"
	@docker compose restart
	@printf "$(COLOR_GREEN)✓ Containers reiniciados!$(COLOR_RESET)\n"

restart-api: ## Reinicia apenas o container da API
	@printf "$(COLOR_BLUE)→ Reiniciando API...$(COLOR_RESET)\n"
	@docker compose restart api
	@printf "$(COLOR_GREEN)✓ API reiniciada!$(COLOR_RESET)\n"

logs: ## Mostra logs de todos os containers
	@docker compose logs -f

logs-api: ## Mostra logs apenas da API
	@docker compose logs -f api

logs-postgres: ## Mostra logs do PostgreSQL
	@docker compose logs -f postgres

logs-redis: ## Mostra logs do Redis
	@docker compose logs -f redis

ps: ## Lista containers em execução
	@docker compose ps

# =============================================================================
# DATABASE - Comandos relacionados ao banco de dados
# =============================================================================
reset-db: ## Reseta o banco de dados (migrations + fixtures)
	@printf "$(COLOR_BLUE)→ Resetando banco de dados...$(COLOR_RESET)\n"
	@docker compose exec api goose -dir ./db/migrations postgres "postgresql://course:course123@postgres:5432/course?sslmode=disable" reset || true
	@docker compose exec api goose -dir ./db/migrations postgres "postgresql://course:course123@postgres:5432/course?sslmode=disable" up
	@printf "$(COLOR_YELLOW)→ Populando com fixtures de desenvolvimento...$(COLOR_RESET)\n"
	@docker compose exec api goose -dir ./db/fixtures postgres "postgresql://course:course123@postgres:5432/course?sslmode=disable" up
	@printf "$(COLOR_GREEN)✓ Banco de dados resetado com sucesso!$(COLOR_RESET)\n"

db-migrate-up: ## Executa migrations pendentes
	@printf "$(COLOR_BLUE)→ Executando migrations...$(COLOR_RESET)\n"
	@docker compose exec api goose -dir ./db/migrations postgres "postgresql://course:course123@postgres:5432/course?sslmode=disable" up
	@printf "$(COLOR_GREEN)✓ Migrations executadas!$(COLOR_RESET)\n"

db-migrate-down: ## Reverte a última migration
	@printf "$(COLOR_BLUE)→ Revertendo migration...$(COLOR_RESET)\n"
	@docker compose exec api goose -dir ./db/migrations postgres "postgresql://course:course123@postgres:5432/course?sslmode=disable" down
	@printf "$(COLOR_GREEN)✓ Migration revertida!$(COLOR_RESET)\n"

db-migrate-status: ## Mostra status das migrations
	@docker compose exec api goose -dir ./db/migrations postgres "postgresql://course:course123@postgres:5432/course?sslmode=disable" status

db-shell: ## Abre shell no PostgreSQL
	@docker compose exec postgres psql -U course -d course

# =============================================================================
# DEVELOPMENT - Comandos de desenvolvimento
# =============================================================================
shell: ## Abre shell no container da API
	@docker compose exec api /bin/bash

shell-root: ## Abre shell como root no container da API
	@docker compose exec -u root api /bin/bash

test: ## Executa testes
	@printf "$(COLOR_BLUE)→ Executando testes...$(COLOR_RESET)\n"
	@docker compose exec api go test -v -coverprofile=coverage.out ./...
	@docker compose exec api go tool cover -func=coverage.out
	@printf "$(COLOR_GREEN)✓ Testes concluídos!$(COLOR_RESET)\n"

test-coverage: ## Executa testes e gera relatório de cobertura HTML
	@printf "$(COLOR_BLUE)→ Gerando relatório de cobertura...$(COLOR_RESET)\n"
	@docker compose exec api go test -v -coverprofile=coverage.out ./...
	@docker compose exec api go tool cover -html=coverage.out -o coverage.html
	@printf "$(COLOR_GREEN)✓ Relatório gerado: coverage.html$(COLOR_RESET)\n"

lint: ## Executa linter (golangci-lint)
	@printf "$(COLOR_BLUE)→ Executando linter...$(COLOR_RESET)\n"
	@docker compose exec api golangci-lint run ./... --timeout=5m
	@printf "$(COLOR_GREEN)✓ Linting concluído!$(COLOR_RESET)\n"

lint-fix: ## Executa linter com correções automáticas
	@printf "$(COLOR_BLUE)→ Executando linter com correções automáticas...$(COLOR_RESET)\n"
	@docker compose exec api golangci-lint run ./... --fix --timeout=5m
	@printf "$(COLOR_GREEN)✓ Linting e correções concluídas!$(COLOR_RESET)\n"

lint-verbose: ## Executa linter com output detalhado
	@printf "$(COLOR_BLUE)→ Executando linter (modo verbose)...$(COLOR_RESET)\n"
	@docker compose exec api golangci-lint run ./... --verbose --timeout=5m
	@printf "$(COLOR_GREEN)✓ Linting concluído!$(COLOR_RESET)\n"

lint-new: ## Executa linter apenas em código novo (uncommited)
	@printf "$(COLOR_BLUE)→ Executando linter em código novo...$(COLOR_RESET)\n"
	@docker compose exec api golangci-lint run ./... --new --timeout=5m
	@printf "$(COLOR_GREEN)✓ Linting concluído!$(COLOR_RESET)\n"

fmt: ## Formata código com gofumpt
	@printf "$(COLOR_BLUE)→ Formatando código...$(COLOR_RESET)\n"
	@docker compose exec api gofumpt -l -w .
	@printf "$(COLOR_GREEN)✓ Código formatado!$(COLOR_RESET)\n"

security-scan: ## Executa scan de segurança com gosec
	@printf "$(COLOR_BLUE)→ Executando scan de segurança...$(COLOR_RESET)\n"
	@docker compose exec api gosec -fmt=json -out=security-report.json ./...
	@printf "$(COLOR_GREEN)✓ Scan concluído: security-report.json$(COLOR_RESET)\n"

security-check: ## Executa análise completa de segurança (gosec + govulncheck)
	@printf "$(COLOR_BLUE)→ Executando análise de segurança completa...$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)1. Executando gosec...$(COLOR_RESET)\n"
	@docker compose exec api gosec -fmt=json -out=security-report.json ./... || true
	@printf "$(COLOR_YELLOW)2. Verificando vulnerabilidades conhecidas...$(COLOR_RESET)\n"
	@docker compose exec api govulncheck ./... || true
	@printf "$(COLOR_GREEN)✓ Análise de segurança concluída!$(COLOR_RESET)\n"

mod-download: ## Baixa dependências Go
	@printf "$(COLOR_BLUE)→ Baixando dependências...$(COLOR_RESET)\n"
	@docker compose exec api go mod download
	@printf "$(COLOR_GREEN)✓ Dependências baixadas!$(COLOR_RESET)\n"

mod-tidy: ## Limpa dependências não utilizadas
	@printf "$(COLOR_BLUE)→ Limpando dependências...$(COLOR_RESET)\n"
	@docker compose exec api go mod tidy
	@printf "$(COLOR_GREEN)✓ Dependências limpas!$(COLOR_RESET)\n"

mod-verify: ## Verifica integridade das dependências
	@printf "$(COLOR_BLUE)→ Verificando dependências...$(COLOR_RESET)\n"
	@docker compose exec api go mod verify
	@printf "$(COLOR_GREEN)✓ Dependências verificadas!$(COLOR_RESET)\n"

# =============================================================================
# SWAGGER - Comandos de documentação da API
# =============================================================================
swagger-init: ## Inicializa/atualiza documentação Swagger da API
	@printf "$(COLOR_BLUE)→ Gerando documentação Swagger...$(COLOR_RESET)\n"
	@docker compose exec api swag init -g cmd/api/main.go -o ./$(DOCS_DIR) --parseDependency --parseInternal
	@printf "$(COLOR_GREEN)✓ Documentação Swagger gerada em ./$(DOCS_DIR)$(COLOR_RESET)\n"
	@printf "$(COLOR_YELLOW)→ Acesse: https://api.course.local:8080/swagger/index.html$(COLOR_RESET)\n"

swagger-generate: swagger-init ## Alias para swagger-init

# =============================================================================
# BUILD - Compilação de binários Go
# =============================================================================
compile: compile-docker ## Compila binário Go (usa Docker Go por padrão)

compile-local: ## Compila binário localmente (requer Go instalado)
	@printf "$(COLOR_BLUE)→ Compilando binário localmente...$(COLOR_RESET)\n"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -trimpath -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@chmod +x $(BUILD_DIR)/$(BINARY_NAME)
	@printf "$(COLOR_GREEN)✓ Binário compilado: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)→ Versão: $(VERSION) | Build: $(BUILD_TIME)$(COLOR_RESET)\n"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

compile-docker: ## Compila binário usando imagem Docker Go (não precisa de containers rodando)
	@printf "$(COLOR_BLUE)→ Compilando binário com Docker Go...$(COLOR_RESET)\n"
	@mkdir -p $(BUILD_DIR)
	@docker run --rm \
		-v $(PWD):/workspace \
		-w /workspace \
		-u $(shell id -u):$(shell id -g) \
		-e GOCACHE=/tmp/go-cache \
		-e GOMODCACHE=/tmp/go-mod \
		-e CGO_ENABLED=0 \
		-e GOOS=linux \
		-e GOARCH=amd64 \
		golang:1.25-alpine \
		sh -c "go mod download >/dev/null 2>&1 && go build $(LDFLAGS_DOCKER) -trimpath -o $(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)"
	@printf "$(COLOR_GREEN)✓ Binário compilado: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)→ Versão: $(VERSION) | Build: $(BUILD_TIME)$(COLOR_RESET)\n"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

compile-container: ## Compila binário dentro do container Docker (requer containers rodando)
	@printf "$(COLOR_BLUE)→ Compilando binário no container...$(COLOR_RESET)\n"
	@mkdir -p $(BUILD_DIR)
	@docker compose exec api sh -c "CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -trimpath -o /app/$(BIN_DIR)/$(BINARY_NAME) $(MAIN_PATH)"
	@docker compose exec api chmod +x /app/$(BIN_DIR)/$(BINARY_NAME)
	@printf "$(COLOR_GREEN)✓ Binário compilado: $(BUILD_DIR)/$(BINARY_NAME)$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)→ Versão: $(VERSION) | Build: $(BUILD_TIME)$(COLOR_RESET)\n"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

clean-binary: ## Remove binários compilados
	@printf "$(COLOR_BLUE)→ Removendo binários...$(COLOR_RESET)\n"
	@rm -f $(BUILD_DIR)/$(BINARY_NAME)
	@printf "$(COLOR_GREEN)✓ Binários removidos!$(COLOR_RESET)\n"

run-binary: compile ## Compila e executa o binário localmente
	@printf "$(COLOR_BLUE)→ Executando binário...$(COLOR_RESET)\n"
	@$(BUILD_DIR)/$(BINARY_NAME)

# =============================================================================
# QUALITY - Comandos de qualidade de código
# =============================================================================
quality-check: lint test ## Executa linter e testes (CI/CD ready)
	@printf "$(COLOR_GREEN)✓ Verificação de qualidade concluída!$(COLOR_RESET)\n"

quality-full: lint-verbose test-coverage security-check ## Análise completa de qualidade
	@printf "$(COLOR_GREEN)✓ Análise completa de qualidade concluída!$(COLOR_RESET)\n"

ci: quality-check ## Simula ambiente CI/CD localmente
	@printf "$(COLOR_GREEN)✓ Pipeline CI simulado com sucesso!$(COLOR_RESET)\n"

# =============================================================================
# PRODUCTION - Comandos de build para produção
# =============================================================================
build-prod: ## Constrói imagem de produção
	@printf "$(COLOR_BLUE)→ Construindo imagem de produção...$(COLOR_RESET)\n"
	@docker build -t $(PROJECT_NAME):latest --target production .
	@printf "$(COLOR_GREEN)✓ Imagem de produção construída!$(COLOR_RESET)\n"

build-prod-debug: ## Constrói imagem de produção com debug
	@printf "$(COLOR_BLUE)→ Construindo imagem de produção com debug...$(COLOR_RESET)\n"
	@docker build -t $(PROJECT_NAME):debug --target production-debug .
	@printf "$(COLOR_GREEN)✓ Imagem de produção com debug construída!$(COLOR_RESET)\n"

size-check: ## Verifica tamanho das imagens Docker
	@printf "$(COLOR_BLUE)→ Tamanho das imagens:$(COLOR_RESET)\n"
	@docker images | grep $(PROJECT_NAME) || printf "$(COLOR_YELLOW)Nenhuma imagem encontrada$(COLOR_RESET)\n"

# =============================================================================
# GIT HOOKS - Configuração de hooks
# =============================================================================
hooks-install: ## Instala Git hooks para qualidade
	@printf "$(COLOR_BLUE)→ Instalando Git hooks...$(COLOR_RESET)\n"
	@./.githooks/setup.sh install
	@printf "$(COLOR_GREEN)✓ Hooks instalados!$(COLOR_RESET)\n"

hooks-uninstall: ## Remove Git hooks
	@printf "$(COLOR_BLUE)→ Removendo Git hooks...$(COLOR_RESET)\n"
	@./.githooks/setup.sh uninstall
	@printf "$(COLOR_GREEN)✓ Hooks removidos!$(COLOR_RESET)\n"

hooks-list: ## Lista status dos Git hooks
	@./.githooks/setup.sh list

# =============================================================================
# CLEANUP - Limpeza de recursos
# =============================================================================
clean: ## Remove arquivos temporários e caches
	@printf "$(COLOR_BLUE)→ Limpando arquivos temporários...$(COLOR_RESET)\n"
	@rm -rf $(TMP_DIR)/* $(BIN_DIR)/* coverage.* security-report.json
	@docker compose exec api go clean -cache -testcache -modcache 2>/dev/null || true
	@printf "$(COLOR_GREEN)✓ Limpeza concluída!$(COLOR_RESET)\n"

clean-all: down-volumes clean ## Remove tudo: containers, volumes, arquivos temporários
	@printf "$(COLOR_BLUE)→ Removendo imagens Docker...$(COLOR_RESET)\n"
	@docker compose down --rmi local
	@printf "$(COLOR_GREEN)✓ Limpeza completa concluída!$(COLOR_RESET)\n"

prune: ## Remove recursos Docker não utilizados (CUIDADO)
	@printf "$(COLOR_YELLOW)⚠ Isso irá remover todos os recursos Docker não utilizados$(COLOR_RESET)\n"
	@read -p "Tem certeza? [y/N] " -n 1 -r; \
	printf "\n"; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker system prune -af --volumes; \
		printf "$(COLOR_GREEN)✓ Recursos removidos!$(COLOR_RESET)\n"; \
	else \
		printf "$(COLOR_BLUE)→ Operação cancelada$(COLOR_RESET)\n"; \
	fi

# =============================================================================
# UTILITIES - Utilitários diversos
# =============================================================================
check-deps: ## Verifica se dependências estão instaladas
	@printf "$(COLOR_BLUE)→ Verificando dependências...$(COLOR_RESET)\n"
	@command -v docker >/dev/null 2>&1 || { printf "$(COLOR_YELLOW)✗ Docker não encontrado$(COLOR_RESET)\n"; exit 1; }
	@docker compose version >/dev/null 2>&1 || command -v docker-compose >/dev/null 2>&1 || { printf "$(COLOR_YELLOW)✗ Docker Compose não encontrado$(COLOR_RESET)\n"; exit 1; }
	@command -v mkcert >/dev/null 2>&1 || printf "$(COLOR_YELLOW)⚠ mkcert não encontrado (opcional)$(COLOR_RESET)\n"
	@printf "$(COLOR_GREEN)✓ Dependências OK!$(COLOR_RESET)\n"

info: ## Mostra informações do ambiente
	@printf "$(COLOR_BOLD)Informações do Ambiente$(COLOR_RESET)\n"
	@printf "$(COLOR_BLUE)Docker:$(COLOR_RESET) $$(docker --version 2>/dev/null || echo 'não encontrado')"
	@printf "$(COLOR_BLUE)Docker Compose:$(COLOR_RESET) $$(docker compose version 2>/dev/null || echo 'não encontrado')"
	@printf "$(COLOR_BLUE)Projeto:$(COLOR_RESET) $(PROJECT_NAME)"
	@printf "$(COLOR_BLUE)Compose File:$(COLOR_RESET) $(COMPOSE_FILE)"
	@printf ""

.DEFAULT_GOAL := help
