#!/bin/bash

# kUber Installation Script
# This script downloads and installs the latest kUber release

set -e

# Configuration
REPO="anindyar/kuber"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="kuber"
GITHUB_API_URL="https://api.github.com/repos/${REPO}/releases/latest"
GITHUB_RELEASE_URL="https://github.com/${REPO}/releases/latest/download"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Utility functions
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    
    case "$os" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            error "Unsupported OS: $os"
            exit 1
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        armv7*)
            ARCH="arm"
            ;;
        *)
            error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
check_prerequisites() {
    info "Checking prerequisites..."
    
    if ! command_exists curl && ! command_exists wget; then
        error "Neither curl nor wget found. Please install one of them."
        exit 1
    fi
    
    if ! command_exists tar; then
        error "tar command not found. Please install tar."
        exit 1
    fi
    
    # Check for kubectl
    if ! command_exists kubectl; then
        warning "kubectl not found. kUber requires kubectl to function properly."
        warning "Please install kubectl: https://kubernetes.io/docs/tasks/tools/"
        echo
        read -p "Continue anyway? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    else
        success "kubectl found"
    fi
}

# Get latest release version
get_latest_version() {
    info "Fetching latest release information..."
    
    if command_exists curl; then
        VERSION=$(curl -sSL "$GITHUB_API_URL" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/' | head -n 1)
    elif command_exists wget; then
        VERSION=$(wget -qO- "$GITHUB_API_URL" | grep '"tag_name":' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/' | head -n 1)
    fi
    
    if [ -z "$VERSION" ]; then
        error "Failed to get latest version"
        exit 1
    fi
    
    info "Latest version: $VERSION"
}

# Download and extract binary
download_and_install() {
    local download_url="${GITHUB_RELEASE_URL}/${BINARY_NAME}-${PLATFORM}.tar.gz"
    local temp_dir=$(mktemp -d)
    local temp_file="${temp_dir}/${BINARY_NAME}-${PLATFORM}.tar.gz"
    
    info "Downloading kUber $VERSION for $PLATFORM..."
    
    if command_exists curl; then
        curl -sSL -o "$temp_file" "$download_url"
    elif command_exists wget; then
        wget -q -O "$temp_file" "$download_url"
    fi
    
    if [ ! -f "$temp_file" ]; then
        error "Download failed"
        exit 1
    fi
    
    info "Extracting archive..."
    tar -xzf "$temp_file" -C "$temp_dir"
    
    # Find the binary (it should be in the extracted directory)
    local binary_path
    if [ -f "${temp_dir}/${BINARY_NAME}" ]; then
        binary_path="${temp_dir}/${BINARY_NAME}"
    else
        # Look for binary in subdirectories
        binary_path=$(find "$temp_dir" -name "$BINARY_NAME" -type f | head -n 1)
    fi
    
    if [ -z "$binary_path" ] || [ ! -f "$binary_path" ]; then
        error "Binary not found in archive"
        exit 1
    fi
    
    # Check if we need sudo for installation
    if [ ! -w "$INSTALL_DIR" ]; then
        warning "Installing to $INSTALL_DIR requires sudo privileges"
        sudo install -m 755 "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    else
        install -m 755 "$binary_path" "$INSTALL_DIR/$BINARY_NAME"
    fi
    
    # Cleanup
    rm -rf "$temp_dir"
    
    success "kUber installed to $INSTALL_DIR/$BINARY_NAME"
}

# Verify installation
verify_installation() {
    info "Verifying installation..."
    
    if command_exists "$BINARY_NAME"; then
        local installed_version
        installed_version=$($BINARY_NAME --version 2>/dev/null || echo "unknown")
        success "kUber is installed and accessible: $installed_version"
    else
        warning "kUber installed but not in PATH"
        warning "You may need to add $INSTALL_DIR to your PATH or restart your terminal"
    fi
}

# Print usage information
print_usage() {
    echo
    success "Installation complete! 🎉"
    echo
    info "Quick Start:"
    echo "  1. Ensure kubectl is configured: kubectl cluster-info"
    echo "  2. Launch kUber: kuber"
    echo "  3. Use Tab to navigate, 'h' for help, 'q' to quit"
    echo
    info "Documentation: https://github.com/${REPO}#readme"
    info "Report issues: https://github.com/${REPO}/issues"
    echo
}

# Handle cleanup on script exit
cleanup() {
    local exit_code=$?
    if [ $exit_code -ne 0 ]; then
        error "Installation failed"
    fi
    exit $exit_code
}

# Main installation function
main() {
    echo -e "${BLUE}"
    echo "╔═══════════════════════════════════════╗"
    echo "║        kUber Installation Script      ║"
    echo "║     An Uber Kubernetes Manager       ║"
    echo "╚═══════════════════════════════════════╝"
    echo -e "${NC}"
    echo
    
    # Set trap for cleanup
    trap cleanup EXIT INT TERM
    
    detect_platform
    check_prerequisites
    get_latest_version
    download_and_install
    verify_installation
    print_usage
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "kUber Installation Script"
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --version, -v  Show script version"
        echo
        echo "Environment Variables:"
        echo "  INSTALL_DIR    Installation directory (default: /usr/local/bin)"
        echo
        exit 0
        ;;
    --version|-v)
        echo "kUber Installation Script v1.0.0"
        exit 0
        ;;
    "")
        # Default behavior - run installation
        main
        ;;
    *)
        error "Unknown option: $1"
        echo "Use --help for usage information"
        exit 1
        ;;
esac