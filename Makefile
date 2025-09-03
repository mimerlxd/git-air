# Git Air - Makefile
# Simple build and installation targets

.PHONY: build install clean test help

# Default target
all: build

# Build the git-air binary
build:
	@echo "🔨 Building git-air..."
	go mod tidy
	go build -o git-air .
	@echo "✅ Build completed: ./git-air"

# Install git-air to /usr/local/bin
install: build
	@echo "📦 Installing git-air to /usr/local/bin..."
	sudo cp git-air /usr/local/bin/git-air
	sudo chmod +x /usr/local/bin/git-air
	@echo "✅ git-air installed successfully"
	@echo "💡 You can now run 'git-air' from anywhere"

# Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	rm -f git-air
	@echo "✅ Clean completed"

# Test the installation script
test-install:
	@echo "🧪 Testing installation script..."
	chmod +x install.sh
	@echo "✅ Installation script is executable"
	@echo "💡 To test full installation: ./install.sh"

# Run git-air locally
run: build
	@echo "🚀 Starting git-air..."
	./git-air

# Show help
help:
	@echo "Git Air - Available targets:"
	@echo ""
	@echo "  build        Build the git-air binary"
	@echo "  install      Build and install to /usr/local/bin"
	@echo "  clean        Remove build artifacts"
	@echo "  test-install Test the installation script"
	@echo "  run          Build and run git-air locally"
	@echo "  help         Show this help message"
	@echo ""
	@echo "Quick start:"
	@echo "  make build   # Build the binary"
	@echo "  make run     # Build and run locally"
	@echo "  make install # Install system-wide"