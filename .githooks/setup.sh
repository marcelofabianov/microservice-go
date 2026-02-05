#!/bin/bash
#
# Script para instalar/desinstalar Git hooks
#

set -e

# Cores para output
COLOR_RESET='\033[0m'
COLOR_BOLD='\033[1m'
COLOR_GREEN='\033[32m'
COLOR_YELLOW='\033[33m'
COLOR_RED='\033[31m'
COLOR_BLUE='\033[34m'

GITHOOKS_DIR=".githooks"
GIT_DIR=".git/hooks"

install_hooks() {
    echo -e "${COLOR_BOLD}${COLOR_BLUE}🔧 Instalando Git hooks...${COLOR_RESET}\n"
    
    # Verifica se estamos em um repositório Git
    if [ ! -d ".git" ]; then
        echo -e "${COLOR_RED}✗ Não é um repositório Git!${COLOR_RESET}"
        exit 1
    fi
    
    # Cria diretório de hooks se não existir
    mkdir -p "$GIT_DIR"
    
    # Instala cada hook
    for hook in "$GITHOOKS_DIR"/*; do
        if [ -f "$hook" ]; then
            hook_name=$(basename "$hook")
            target="$GIT_DIR/$hook_name"
            
            # Remove hook existente se for um symlink
            if [ -L "$target" ]; then
                rm "$target"
            fi
            
            # Cria symlink
            ln -sf "../../$hook" "$target"
            chmod +x "$hook"
            
            echo -e "${COLOR_GREEN}✓ Instalado: ${hook_name}${COLOR_RESET}"
        fi
    done
    
    echo -e "\n${COLOR_BOLD}${COLOR_GREEN}✅ Hooks instalados com sucesso!${COLOR_RESET}\n"
    echo -e "${COLOR_BLUE}Hooks ativos:${COLOR_RESET}"
    echo -e "  ${COLOR_GREEN}pre-commit${COLOR_RESET}  - Formatação, linting e testes rápidos"
    echo -e "  ${COLOR_GREEN}pre-push${COLOR_RESET}    - Verificação completa de qualidade"
    echo -e "  ${COLOR_GREEN}commit-msg${COLOR_RESET}  - Validação de Conventional Commits"
    echo ""
}

uninstall_hooks() {
    echo -e "${COLOR_BOLD}${COLOR_YELLOW}🗑️  Desinstalando Git hooks...${COLOR_RESET}\n"
    
    for hook in "$GITHOOKS_DIR"/*; do
        if [ -f "$hook" ]; then
            hook_name=$(basename "$hook")
            target="$GIT_DIR/$hook_name"
            
            if [ -L "$target" ]; then
                rm "$target"
                echo -e "${COLOR_YELLOW}✓ Removido: ${hook_name}${COLOR_RESET}"
            fi
        fi
    done
    
    echo -e "\n${COLOR_GREEN}✓ Hooks desinstalados${COLOR_RESET}\n"
}

list_hooks() {
    echo -e "${COLOR_BOLD}${COLOR_BLUE}📋 Git Hooks Disponíveis:${COLOR_RESET}\n"
    
    for hook in "$GITHOOKS_DIR"/*; do
        if [ -f "$hook" ]; then
            hook_name=$(basename "$hook")
            target="$GIT_DIR/$hook_name"
            
            if [ -L "$target" ]; then
                status="${COLOR_GREEN}✓ Instalado${COLOR_RESET}"
            else
                status="${COLOR_RED}✗ Não instalado${COLOR_RESET}"
            fi
            
            echo -e "  ${COLOR_BOLD}${hook_name}${COLOR_RESET} - $status"
        fi
    done
    echo ""
}

show_help() {
    echo -e "${COLOR_BOLD}Git Hooks Setup${COLOR_RESET}\n"
    echo -e "Uso: $0 [comando]\n"
    echo -e "Comandos:"
    echo -e "  ${COLOR_GREEN}install${COLOR_RESET}    - Instala os Git hooks"
    echo -e "  ${COLOR_GREEN}uninstall${COLOR_RESET}  - Remove os Git hooks"
    echo -e "  ${COLOR_GREEN}list${COLOR_RESET}       - Lista hooks disponíveis e status"
    echo -e "  ${COLOR_GREEN}help${COLOR_RESET}       - Mostra esta mensagem"
    echo ""
}

# Processa comando
case "${1:-install}" in
    install)
        install_hooks
        ;;
    uninstall)
        uninstall_hooks
        ;;
    list)
        list_hooks
        ;;
    help)
        show_help
        ;;
    *)
        echo -e "${COLOR_RED}Comando inválido: $1${COLOR_RESET}\n"
        show_help
        exit 1
        ;;
esac
