#!/bin/bash
# wtool installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/Danceiny/wtool/main/scripts/install.sh | bash

set -euo pipefail

REPO="Danceiny/wtool"
BINARY_NAME="wtool"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
DEFAULT_VERSION="v0.1.0"

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

# Get latest release version (with multiple fallbacks)
get_latest_version() {
    local version=""
    
    # Try GitHub API first (often rate-limited)
    local api_response
    api_response=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" 2>/dev/null || true)
    if [ -n "$api_response" ]; then
        version=$(echo "$api_response" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    fi
    
    # Fallback 1: try gh CLI if available
    if [ -z "$version" ] && command -v gh >/dev/null 2>&1; then
        version=$(gh release list --repo "${REPO}" --limit 1 2>/dev/null | awk '{print $1}')
    fi
    
    # Fallback 2: scrape from GitHub releases page HTML
    if [ -z "$version" ]; then
        local html_version
        html_version=$(curl -fsSL "https://github.com/${REPO}/releases" 2>/dev/null | \
            grep -oE 'href="/'"${REPO}"'/releases/tag/[^"]+' | \
            head -1 | \
            sed -E 's|.*/tag/||')
        if [ -n "$html_version" ]; then
            version="$html_version"
        fi
    fi
    
    # Final fallback: use default version
    if [ -z "$version" ]; then
        version="${DEFAULT_VERSION}"
        echo "Warning: Could not fetch latest version from GitHub (API rate limited), using ${version}" >&2
        echo "To install a specific version, run: VERSION=v0.x.x curl ... | bash" >&2
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
    echo "URL: ${download_url}"
    
    tmp_dir=$(mktemp -d)
    trap 'rm -rf "$tmp_dir"' EXIT
    
    # Try to download with better error handling
    if ! curl -fsSL --retry 3 --retry-delay 2 "$download_url" | tar -xz -C "$tmp_dir" 2>/dev/null; then
        echo "Download failed. The release may not have binary artifacts yet." >&2
        echo "You can build from source instead:" >&2
        echo "  git clone https://github.com/${REPO}.git && cd wtool && make build && make install" >&2
        exit 1
    fi
    
    if [ ! -f "$tmp_dir/$BINARY_NAME" ]; then
        echo "Binary not found in archive" >&2
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
