#!/bin/bash

set -e

REPO="g0ulartleo/mirante"
BINARY_NAME="mirante"
INSTALL_DIR="/usr/local/bin"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_platform() {
    local os
    local arch

    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)

    if [[ "$os" == *"mingw"* ]] || [[ "$os" == *"msys"* ]] || [[ "$os" == *"cygwin"* ]]; then
        OS="windows"
        BINARY_EXT=".exe"
        INSTALL_DIR="$HOME/.local/bin"
    else
        case $os in
            linux)
                OS="linux"
                BINARY_EXT=""
                ;;
            darwin)
                OS="darwin"
                BINARY_EXT=""
                ;;
            *)
                log_error "Unsupported operating system: $os"
                exit 1
                ;;
        esac
    fi

    case $arch in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            log_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac

    PLATFORM="${OS}-${ARCH}"
    log_info "Detected platform: $PLATFORM"
}

get_latest_version() {
    log_info "Fetching latest release version..."

    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        log_error "Neither curl nor wget is available. Please install one of them."
        exit 1
    fi

    if [ -z "$VERSION" ]; then
        log_error "Failed to get latest version"
        exit 1
    fi

    log_info "Latest version: $VERSION"
}

install_binary() {
    local binary_name="${BINARY_NAME}-${PLATFORM}${BINARY_EXT}"
    local download_url="https://github.com/$REPO/releases/download/$VERSION/$binary_name"
    local temp_file="/tmp/$binary_name"

    log_info "Downloading $binary_name..."

    if command -v curl >/dev/null 2>&1; then
        curl -sL "$download_url" -o "$temp_file"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "$temp_file"
    fi

    if [ ! -f "$temp_file" ]; then
        log_error "Failed to download binary"
        exit 1
    fi

    if [[ "$OS" != "windows" ]]; then
        chmod +x "$temp_file"
    fi

    if [[ "$OS" == "windows" ]]; then
        mkdir -p "$INSTALL_DIR"
    fi

    if [ -w "$INSTALL_DIR" ]; then
        log_info "Installing $BINARY_NAME to $INSTALL_DIR..."
        mv "$temp_file" "$INSTALL_DIR/$BINARY_NAME${BINARY_EXT}"
    else
        log_info "Installing $BINARY_NAME to $INSTALL_DIR (requires sudo)..."
        sudo mv "$temp_file" "$INSTALL_DIR/$BINARY_NAME${BINARY_EXT}"
    fi

    log_info "Installation completed successfully!"
}

verify_installation() {
    local binary_path="$INSTALL_DIR/$BINARY_NAME${BINARY_EXT}"

    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log_info "$BINARY_NAME is now available in your PATH"
        log_info "Run '$BINARY_NAME help' to get started"
    elif [[ "$OS" == "windows" ]] && [ -f "$binary_path" ]; then
        log_info "$BINARY_NAME installed to $binary_path"
        log_warn "Run '$binary_path login' to get started"

        if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            log_warn "$INSTALL_DIR is not in your PATH"
            log_info "To add to PATH permanently, run:"
            log_info "setx PATH \"%PATH%;$INSTALL_DIR\""
        fi
    else
        log_warn "$BINARY_NAME was installed but is not in your PATH"

        if [[ "$OS" != "windows" ]] && [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
            log_warn "$INSTALL_DIR is not in your PATH"
            log_info "Adding $INSTALL_DIR to PATH..."

            SHELL_PROFILE=""
            if [[ "$SHELL" == *"zsh"* ]]; then
                SHELL_PROFILE="$HOME/.zshrc"
            elif [[ "$SHELL" == *"bash"* ]]; then
                SHELL_PROFILE="$HOME/.bashrc"
                if [ ! -f "$SHELL_PROFILE" ]; then
                    SHELL_PROFILE="$HOME/.bash_profile"
                fi
            fi

            if [ -n "$SHELL_PROFILE" ]; then
                echo "" >> "$SHELL_PROFILE"
                echo "# Added by Mirante CLI installer" >> "$SHELL_PROFILE"
                echo "export PATH=\"$INSTALL_DIR:\$PATH\"" >> "$SHELL_PROFILE"
                log_info "Added $INSTALL_DIR to $SHELL_PROFILE"
                log_warn "Please restart your terminal or run: source $SHELL_PROFILE"
            else
                log_warn "Could not detect shell profile. Please manually add $INSTALL_DIR to your PATH"
            fi
        else
            log_warn "You may need to restart your terminal for changes to take effect"
        fi
    fi
}

main() {
    log_info "Installing Mirante CLI..."

    detect_platform
    get_latest_version
    install_binary
    verify_installation

    log_info "Installation completed!"
}

main "$@"
