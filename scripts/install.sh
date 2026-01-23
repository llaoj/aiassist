#!/bin/bash

# aiassist installation script
# Usage: curl -fsSL https://raw.githubusercontent.com/llaoj/aiassist/main/scripts/install.sh | bash

set -e

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
REPO_OWNER="llaoj"
REPO_NAME="aiassist"
BINARY_NAME="aiassist"
INSTALL_DIR="/usr/local/bin"

# Print colored message
print_info() {
    echo -e "${BLUE}ℹ ${NC}$1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

# Detect OS
detect_os() {
    local os=""
    local arch=""
    
    # Detect OS type
    case "$(uname -s)" in
        Linux*)
            os="linux"
            ;;
        Darwin*)
            os="darwin"
            ;;
        CYGWIN*|MINGW*|MSYS*)
            os="windows"
            ;;
        FreeBSD*)
            os="freebsd"
            ;;
        *)
            print_error "Unsupported operating system: $(uname -s)"
            exit 1
            ;;
    esac
    
    # Detect architecture
    case "$(uname -m)" in
        x86_64|amd64)
            arch="amd64"
            ;;
        aarch64|arm64)
            arch="arm64"
            ;;
        armv7l|armv6l)
            arch="arm"
            ;;
        i386|i686)
            arch="386"
            ;;
        *)
            print_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
    
    OS_TYPE="$os"
    ARCH_TYPE="$arch"
    
    print_info "Detected OS: ${OS_TYPE}, Architecture: ${ARCH_TYPE}"
}

# Get latest release version from GitHub
get_latest_version() {
    print_info "Fetching latest release version..."
    
    # Try to get latest release from GitHub API
    if command -v curl >/dev/null 2>&1; then
        VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    elif command -v wget >/dev/null 2>&1; then
        VERSION=$(wget -qO- "https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    if [ -z "$VERSION" ]; then
        print_warning "Could not fetch latest version, using 'latest'"
        VERSION="latest"
    else
        print_success "Latest version: ${VERSION}"
    fi
}

# Download binary
download_binary() {
    local download_url=""
    local tmp_dir=$(mktemp -d)
    local file_extension=""
    
    # Set file extension based on OS
    if [ "$OS_TYPE" = "windows" ]; then
        file_extension=".exe"
    else
        file_extension=""
    fi
    
    # Construct download URL
    if [ "$VERSION" = "latest" ]; then
        download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/${BINARY_NAME}-${OS_TYPE}-${ARCH_TYPE}${file_extension}"
    else
        download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${BINARY_NAME}-${OS_TYPE}-${ARCH_TYPE}${file_extension}"
    fi
    
    print_info "Downloading from: ${download_url}"
    
    # Download binary
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$download_url" -o "${tmp_dir}/${BINARY_NAME}${file_extension}"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$download_url" -O "${tmp_dir}/${BINARY_NAME}${file_extension}"
    else
        print_error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    if [ $? -ne 0 ]; then
        print_error "Failed to download binary"
        rm -rf "$tmp_dir"
        exit 1
    fi
    
    DOWNLOADED_FILE="${tmp_dir}/${BINARY_NAME}${file_extension}"
    print_success "Download completed"
}

# Install binary
install_binary() {
    print_info "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
    
    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        SUDO=""
    else
        if command -v sudo >/dev/null 2>&1; then
            SUDO="sudo"
            print_warning "Need sudo permission to install to ${INSTALL_DIR}"
        else
            print_error "No write permission to ${INSTALL_DIR} and sudo not found"
            print_info "Please install manually or run with appropriate permissions"
            exit 1
        fi
    fi
    
    # Make binary executable
    chmod +x "$DOWNLOADED_FILE"
    
    # Move to install directory
    $SUDO mv -f "$DOWNLOADED_FILE" "${INSTALL_DIR}/${BINARY_NAME}"
    
    if [ $? -eq 0 ]; then
        print_success "${BINARY_NAME} installed successfully to ${INSTALL_DIR}/${BINARY_NAME}"
    else
        print_error "Installation failed"
        exit 1
    fi
    
    # Cleanup
    rm -rf "$(dirname "$DOWNLOADED_FILE")"
}

# Verify installation
verify_installation() {
    print_info "Verifying installation..."
    
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        print_success "${BINARY_NAME} is installed and available in PATH"
        
        # Try to get version
        version_output=$("$BINARY_NAME" version 2>/dev/null || echo "")
        if [ -n "$version_output" ]; then
            echo ""
            echo "$version_output"
        fi
    else
        print_warning "${BINARY_NAME} is installed but not in PATH"
        print_info "You may need to add ${INSTALL_DIR} to your PATH:"
        echo "  export PATH=\"\$PATH:${INSTALL_DIR}\""
    fi
}

# Main installation flow
main() {
    echo ""
    echo "╔════════════════════════════════════════╗"
    echo "║   aiassist Installation Script         ║"
    echo "╚════════════════════════════════════════╝"
    echo ""
    
    # Detect OS and architecture
    detect_os
    
    # Get latest version
    get_latest_version
    
    # Download binary
    download_binary
    
    # Install binary
    install_binary
    
    # Verify installation
    verify_installation
    
    echo ""
    print_success "Installation completed!"
    echo ""
    print_info "Get started by running: ${BINARY_NAME} --help"
    echo ""
}

# Run main function
main
