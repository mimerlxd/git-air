#!/bin/bash

# Git Air - One-line installer for MacOS and Linux Ubuntu
# Usage: curl -sSL https://raw.githubusercontent.com/your-repo/git-air/main/install.sh | bash

set -e

GIT_AIR_VERSION="latest"
GIT_AIR_REPO="https://github.com/your-repo/git-air"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR="/tmp/git-air-install-$$"
GO_VERSION="1.21.0"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}"
    echo "🚀 Git Air Installer"
    echo "==================="
    echo -e "${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

detect_os() {
    if [[ "$OSTYPE" == "darwin"* ]]; then
        OS="macos"
        ARCH=$(uname -m)
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        OS="linux"
        ARCH=$(uname -m)
        # Check if it's Ubuntu
        if ! command -v apt-get &> /dev/null; then
            print_error "This installer currently supports Ubuntu/Debian systems only"
            exit 1
        fi
    else
        print_error "Unsupported operating system: $OSTYPE"
        exit 1
    fi
    
    print_info "Detected OS: $OS ($ARCH)"
}

check_dependencies() {
    print_info "Checking dependencies..."
    
    # Check for curl
    if ! command -v curl &> /dev/null; then
        print_error "curl is required but not installed"
        exit 1
    fi
    
    # Check for git
    if ! command -v git &> /dev/null; then
        print_error "git is required but not installed"
        if [[ "$OS" == "linux" ]]; then
            print_info "Installing git..."
            sudo apt-get update && sudo apt-get install -y git
        else
            print_error "Please install git first"
            exit 1
        fi
    fi
    
    print_success "Dependencies check passed"
}

install_go() {
    if command -v go &> /dev/null; then
        GO_CURRENT_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
        print_info "Go is already installed (version $GO_CURRENT_VERSION)"
        return 0
    fi
    
    print_info "Installing Go compiler..."
    
    if [[ "$OS" == "macos" ]]; then
        if command -v brew &> /dev/null; then
            brew install go
        else
            print_warning "Homebrew not found. Installing Go manually..."
            GO_ARCHIVE="go${GO_VERSION}.darwin-${ARCH}.tar.gz"
            curl -sSL "https://golang.org/dl/${GO_ARCHIVE}" -o "/tmp/${GO_ARCHIVE}"
            sudo tar -C /usr/local -xzf "/tmp/${GO_ARCHIVE}"
            export PATH=$PATH:/usr/local/go/bin
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
            rm "/tmp/${GO_ARCHIVE}"
        fi
    elif [[ "$OS" == "linux" ]]; then
        GO_ARCHIVE="go${GO_VERSION}.linux-${ARCH}.tar.gz"
        curl -sSL "https://golang.org/dl/${GO_ARCHIVE}" -o "/tmp/${GO_ARCHIVE}"
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf "/tmp/${GO_ARCHIVE}"
        export PATH=$PATH:/usr/local/go/bin
        echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        rm "/tmp/${GO_ARCHIVE}"
    fi
    
    print_success "Go compiler installed"
}

download_and_compile() {
    print_info "Creating temporary directory..."
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    print_info "Downloading git-air source code..."
    git clone "$GIT_AIR_REPO" .
    
    print_info "Compiling git-air..."
    go mod tidy
    go build -o git-air .
    
    print_success "Compilation completed"
}

install_binary() {
    print_info "Installing git-air to $INSTALL_DIR..."
    
    # Check if we need sudo
    if [[ ! -w "$INSTALL_DIR" ]]; then
        sudo cp git-air "$INSTALL_DIR/git-air"
        sudo chmod +x "$INSTALL_DIR/git-air"
    else
        cp git-air "$INSTALL_DIR/git-air"
        chmod +x "$INSTALL_DIR/git-air"
    fi
    
    print_success "git-air installed to $INSTALL_DIR/git-air"
}

cleanup() {
    print_info "Cleaning up temporary files..."
    cd /
    rm -rf "$TEMP_DIR"
    print_success "Cleanup completed"
}

show_help() {
    echo -e "${GREEN}"
    echo "🎉 Git Air installation completed successfully!"
    echo ""
    echo "📚 Quick Start Guide:"
    echo "==================="
    echo ""
    echo "1. Navigate to your project directory:"
    echo "   cd /path/to/your/project"
    echo ""
    echo "2. Start git-air:"
    echo "   git-air"
    echo ""
    echo "✨ Git Air will automatically:"
    echo "• 🔍 Discover all Git repositories in current directory and subdirectories"
    echo "• 📝 Auto-commit changes every 30 seconds"
    echo "• 🚀 Push to ALL configured remotes immediately"
    echo "• 📡 Pull updates from remotes every minute"
    echo "• 📚 Handle monorepos and submodules"
    echo ""
    echo "🔧 Advanced Usage:"
    echo "• git-air -help          Show all options"
    echo "• git-air -log debug     Enable debug logging"
    echo "• git-air -scan \"path1,path2\"  Monitor specific paths"
    echo ""
    echo "🏠 For development servers:"
    echo "• Run git-air in your home directory to monitor all projects"
    echo "• Perfect for auto-syncing multiple repositories"
    echo ""
    echo "📖 Documentation: https://github.com/your-repo/git-air"
    echo -e "${NC}"
}

# Main installation flow
main() {
    print_header
    
    # Trap to ensure cleanup on exit
    trap cleanup EXIT
    
    detect_os
    check_dependencies
    install_go
    download_and_compile
    install_binary
    
    show_help
    
    print_success "Installation completed! You can now use 'git-air' command anywhere."
}

# Run main function
main "$@"