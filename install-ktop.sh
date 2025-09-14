#!/bin/bash
set -e

# kTop Installer Script
# Usage: curl -fsSL https://raw.githubusercontent.com/anindyar/kuber/main/install-ktop.sh | bash

REPO_URL="https://github.com/anindyar/kuber"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR=$(mktemp -d)
BINARY_NAME="ktop"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

trap cleanup EXIT

check_requirements() {
    print_status "Checking requirements..."
    
    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21+ and try again."
        print_status "Install Go from: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    GO_MAJOR=$(echo $GO_VERSION | cut -d. -f1)
    GO_MINOR=$(echo $GO_VERSION | cut -d. -f2)
    
    if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
        print_warning "Go version $GO_VERSION detected. kTop requires Go 1.21+."
        print_status "Your version should work, but consider updating if you encounter issues."
    else
        print_success "Go $GO_VERSION detected"
    fi
    
    # Check if git is available
    if ! command -v git >/dev/null 2>&1; then
        print_error "Git is not installed. Please install git and try again."
        exit 1
    fi
    
    # Check if kubectl is available (optional but recommended)
    if ! command -v kubectl >/dev/null 2>&1; then
        print_warning "kubectl not found. kTop uses kubectl for some operations."
        print_status "Install kubectl from: https://kubernetes.io/docs/tasks/tools/"
    else
        print_success "kubectl found"
    fi
}

detect_os_arch() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) 
            print_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac
    
    print_success "Detected OS: $OS, Architecture: $ARCH"
}

check_permissions() {
    if [ ! -w "$INSTALL_DIR" ] && [ "$EUID" -ne 0 ]; then
        print_warning "No write permission to $INSTALL_DIR"
        print_status "Trying to install with sudo..."
        SUDO_CMD="sudo"
    else
        SUDO_CMD=""
    fi
}

install_ktop() {
    print_status "Installing kTop..."
    
    cd "$TEMP_DIR"
    
    # Clone the repository
    print_status "Cloning kTop repository..."
    git clone --depth 1 "$REPO_URL" kuber
    
    cd kuber
    
    # Build kTop
    print_status "Building kTop binary..."
    if ! go build -ldflags "-s -w" -o "$BINARY_NAME" ./cmd/ktop; then
        print_error "Failed to build kTop"
        exit 1
    fi
    
    # Install binary
    print_status "Installing kTop to $INSTALL_DIR..."
    if ! $SUDO_CMD mv "$BINARY_NAME" "$INSTALL_DIR/"; then
        print_error "Failed to install kTop to $INSTALL_DIR"
        print_status "Trying alternative installation to ~/bin..."
        
        # Try installing to user's bin directory
        USER_BIN="$HOME/bin"
        mkdir -p "$USER_BIN"
        mv "$BINARY_NAME" "$USER_BIN/"
        
        # Add to PATH if not already there
        SHELL_RC=""
        case $SHELL in
            */bash) SHELL_RC="$HOME/.bashrc" ;;
            */zsh) SHELL_RC="$HOME/.zshrc" ;;
            */fish) SHELL_RC="$HOME/.config/fish/config.fish" ;;
        esac
        
        if [ -n "$SHELL_RC" ] && [ -f "$SHELL_RC" ]; then
            if ! grep -q "$USER_BIN" "$SHELL_RC"; then
                echo 'export PATH="$HOME/bin:$PATH"' >> "$SHELL_RC"
                print_status "Added $USER_BIN to PATH in $SHELL_RC"
                print_warning "Please run: source $SHELL_RC or restart your terminal"
            fi
        fi
        
        print_success "kTop installed to $USER_BIN/ktop"
        INSTALL_DIR="$USER_BIN"
    else
        print_success "kTop installed to $INSTALL_DIR/ktop"
    fi
    
    # Make sure binary is executable
    chmod +x "$INSTALL_DIR/ktop"
}

verify_installation() {
    print_status "Verifying installation..."
    
    if command -v ktop >/dev/null 2>&1; then
        VERSION_OUTPUT=$(ktop --version 2>/dev/null || echo "kTop installed successfully")
        print_success "kTop is available in PATH"
        print_status "Version: $VERSION_OUTPUT"
    else
        print_warning "kTop not found in PATH. You may need to:"
        print_status "1. Restart your terminal, or"
        print_status "2. Run: source ~/.bashrc (or your shell's config file), or"
        print_status "3. Add $INSTALL_DIR to your PATH manually"
    fi
}

show_usage() {
    cat << 'EOF'

🚀 kTop Installation Complete!

Quick Start:
  ktop                          # Launch with default context
  ktop --context=my-cluster     # Use specific context
  ktop --kubeconfig=/path/config # Use custom kubeconfig

Key Features:
  • Real-time cluster monitoring
  • Resource navigation with Tab/Arrow keys  
  • Pod and deployment log viewing
  • Search functionality in logs
  • Read-only security

Navigation:
  Enter    - Navigate forward/select
  ↑/↓      - Navigate lists
  Tab      - Switch panes
  l        - View logs (pods/deployments)
  d        - View resource details
  /        - Search logs
  Esc      - Go back
  q        - Quit

For help: ktop --help
Documentation: https://github.com/anindyar/kuber/tree/main/cmd/ktop

EOF
}

main() {
    echo "🏗️  kTop Installer"
    echo "=================="
    
    detect_os_arch
    check_requirements
    check_permissions
    install_ktop
    verify_installation
    show_usage
    
    print_success "Installation completed successfully! 🎉"
}

main "$@"