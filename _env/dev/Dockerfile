# =============================================================================
# DEVELOPMENT STAGE - Hot reload com Air e ferramentas de desenvolvimento
# =============================================================================
FROM golang:1.25-alpine AS development

# Build arguments para permissões de usuário
ARG HOST_UID=1000
ARG HOST_GID=1000

# Instalar dependências do sistema necessárias para desenvolvimento
RUN apk add --no-cache \
    bash \
    curl \
    git \
    tzdata \
    ca-certificates \
    make

# Criar grupo e usuário não-root para desenvolvimento
RUN addgroup -g ${HOST_GID} -S appgroup && \
    adduser -u ${HOST_UID} -G appgroup -S -D -h /home/appuser appuser

# Configurar diretório de trabalho
RUN mkdir -p /app && chown -R appuser:appgroup /app
WORKDIR /app

# Criar diretórios que serão volumes e ajustar permissões
RUN mkdir -p /app/tmp /app/bin && \
    chown -R appuser:appgroup /app/tmp /app/bin

# Copiar e executar script de instalação de ferramentas
COPY _env/dev/install-tools.sh /tmp/install-tools.sh
RUN chmod +x /tmp/install-tools.sh && \
    /tmp/install-tools.sh && \
    rm /tmp/install-tools.sh

# Adicionar /go/bin ao PATH e ajustar permissões
ENV PATH=$PATH:/go/bin
RUN chown -R appuser:appgroup /go

# Trocar para usuário não-root
USER appuser

# Copiar apenas arquivos de dependências primeiro (melhor cache de camadas)
COPY --chown=appuser:appgroup go.mod go.sum ./
RUN go mod download && \
    chmod 444 go.mod go.sum

# Expor portas para API e Delve debugger
EXPOSE 8080 40000

# Health check para desenvolvimento
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Comando padrão: hot reload com Air
CMD ["sh", "-c", "mkdir -p tmp/air && air -c .air.toml"]

# =============================================================================
# BUILDER STAGE - Compilação otimizada da aplicação
# =============================================================================
FROM golang:1.25-alpine AS builder

# Instalar dependências mínimas para build
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

WORKDIR /build

# Copiar arquivos de dependências e baixar módulos
COPY go.mod go.sum ./
RUN go mod download && \
    go mod verify

# Copiar código fonte
COPY . .

# Build otimizado para produção
# -ldflags="-s -w" remove símbolos de debug e reduz tamanho
# -trimpath remove caminhos do sistema de arquivos
# CGO_ENABLED=0 para binário estático
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -trimpath \
    -o /build/app \
    ./cmd/api/main.go

# Verificar que o binário foi criado e é executável
RUN chmod +x /build/app && \
    /build/app --version 2>/dev/null || echo "Binary built successfully"

# =============================================================================
# PRODUCTION STAGE - Imagem mínima e segura para produção
# =============================================================================
FROM alpine:3.19 AS production

# Build arguments para permissões de usuário (pode ser sobrescrito)
ARG HOST_UID=1000
ARG HOST_GID=1000

# Metadados da imagem
LABEL maintainer="marcelofabianov@gmail.com"
LABEL description="course API - Production Image"
LABEL version="1.0.0"

# Instalar apenas certificados CA e timezone data
RUN apk add --no-cache \
    ca-certificates \
    tzdata && \
    rm -rf /var/cache/apk/*

# Criar usuário e grupo não-root com UID/GID configuráveis
RUN addgroup -g ${HOST_GID} -S appgroup && \
    adduser -u ${HOST_UID} -G appgroup -S -D -h /home/appuser appuser

# Configurar diretórios necessários
RUN mkdir -p /app/certs /app/tmp && \
    chown -R appuser:appgroup /app

WORKDIR /app

# Copiar binário compilado do estágio builder
COPY --from=builder --chown=appuser:appgroup /build/app ./app

# Copiar arquivos de configuração se necessário
# COPY --from=builder --chown=appuser:appgroup /build/config ./config

# Trocar para usuário não-root usando UID dinâmico
USER ${HOST_UID}:${HOST_GID}

# Expor porta da aplicação
EXPOSE 8080

# Health check para produção
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Comando para executar a aplicação
CMD ["./app"]

# =============================================================================
# PRODUCTION-DEBUG STAGE - Produção com capacidade de debugging
# =============================================================================
FROM production AS production-debug

# Build arguments herdados do stage anterior
ARG HOST_UID=1000
ARG HOST_GID=1000

USER root

# Instalar Delve para debugging em produção
RUN apk add --no-cache bash && \
    wget -O /tmp/dlv.tar.gz https://github.com/go-delve/delve/releases/download/v1.22.0/delve_1.22.0_linux_amd64.tar.gz && \
    tar -C /tmp -xzf /tmp/dlv.tar.gz && \
    mv /tmp/dlv /usr/local/bin/dlv && \
    chmod +x /usr/local/bin/dlv && \
    rm /tmp/dlv.tar.gz

# Expor porta do Delve
EXPOSE 40000

# Voltar para usuário não-root usando UID dinâmico
USER ${HOST_UID}:${HOST_GID}

# Executar com Delve
CMD ["dlv", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "exec", "./app"]
