#!/bin/bash
# Script para instalar ferramentas de desenvolvimento Go
# Executado durante o build da imagem Docker ou manualmente

set -e

echo "📦 Instalando ferramentas de desenvolvimento Go..."

# Lista de ferramentas a instalar
tools=(
    "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
    "github.com/pressly/goose/v3/cmd/goose@latest"
    "github.com/air-verse/air@latest"
    "github.com/securego/gosec/v2/cmd/gosec@latest"
    "mvdan.cc/gofumpt@latest"
    "golang.org/x/tools/cmd/goimports@latest"
    "golang.org/x/vuln/cmd/govulncheck@latest"
    "github.com/swaggo/swag/cmd/swag@latest"
    "github.com/go-delve/delve/cmd/dlv@latest"
)

# Instalar cada ferramenta
for tool in "${tools[@]}"; do
    tool_name=$(basename "$tool" | cut -d'@' -f1)
    echo "  → Instalando $tool_name..."
    go install "$tool" || echo "  ⚠️  Falha ao instalar $tool_name"
done

echo "✅ Instalação concluída!"
echo ""
echo "Ferramentas instaladas em: $GOPATH/bin"
ls -lh "$GOPATH/bin" 2>/dev/null || ls -lh /go/bin 2>/dev/null || true
