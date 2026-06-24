#!/bin/bash
# wtool installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/Danceiny/wtool/main/scripts/install.sh | bash

set -euo pipefail

REPO="Danceiny/wtool"
BINARY_NAME="wtool"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
detect_platform() {
    local os arch
    
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    
    case "$arch" in
        x86_64)  arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo "Unsupported architecture: $arch" >&2; exit 1 ;;
    esac
    
    case "$os" in
        linux) os="linux" ;;
        darwin) os="darwin" ;;
        *) echo "Unsupported OS: $os" >&2; exit 1 ;;
    esac
    
    echo "${os}_${arch}"
}

# Get latest release version
get_latest_version() {
    local version
    version=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null | \
        grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        echo "Failed to get latest version" >&2
        exit 1
    fi
    
    echo "$version"
}

# Download and install
download_and_install() {
    local version platform download_url tmp_dir
    
    version="$1"
    platform="$2"
    
    download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}_${platform}.tar.gz"
    
    echo "Downloading wtool ${version} for ${platform}..."
    
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT
    
    curl -fsSL "$download_url" | tar -xz -C "$tmp_dir"
    
    if [ ! -f "$tmp_dir/$BINARY_NAME" ]; then
        echo "Download failed or binary not found in archive" >&2
        exit 1
    fi
    
    # Install
    echo "Installing to ${INSTALL_DIR}..."
    if [ -w "$INSTALL_DIR" ]; then
        mv "$tmp_dir/$BINARY_NAME" "$INSTALL_DIR/"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
    else
        echo "Need sudo to install to ${INSTALL_DIR}"
        sudo mv "$tmp_dir/$BINARY_NAME" "$INSTALL_DIR/"
        sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    echo "wtool ${version} installed successfully!"
    echo "Run 'wtool doctor' to verify installation."
}

# Install shell completions
install_completions() {
    local shell
    shell=$(basename "$SHELL")
    
    case "$shell" in
        bash)
            if [ -d "/usr/local/etc/bash_completion.d" ]; then
                wtool completion bash > /usr/local/etc/bash_completion.d/wtool 2>/dev/null || true
            elif [ -d "/etc/bash_completion.d" ]; then
                wtool completion bash > /etc/bash_completion.d/wtool 2>/dev/null || true
            fi
            ;;
        zsh)
            if [ -d "/usr/local/share/zsh/site-functions" ]; then
                wtool completion zsh > /usr/local/share/zsh/site-functions/_wtool 2>/dev/null || true
            fi
            ;;
    esac
}

# Main
main() {
    echo "=== wtool Installer ==="
    echo
    
    # Check if wtool already installed
    if command -v wtool >/dev/null 2>&1; then
        echo "wtool already installed: $(wtool --version)"
        read -p "Reinstall? [y/N] " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 0
        fi
    fi
    
    platform=$(detect_platform)
    version=$(get_latest_version)
    
    download_and_install "$version" "$platform"
    install_completions
    
    echo
    echo "Next steps:"
    echo "  1. Run 'wtool doctor' to check your environment"
    echo "  2. Run 'wtool setup hook --global' to install git hooks"
    echo "  3. Run 'wtool init' in any worktree to initialize it"
}

main "$@"
