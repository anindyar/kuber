.PHONY: build test lint clean install run help

# Build variables
BINARY_NAME=kuber
BUILD_DIR=bin
GO_VERSION=1.21
MAIN_PATH=./cmd/kuber

# Version information
VERSION ?= $(shell git describe --tags --always --dirty)
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT ?= $(shell git rev-parse HEAD)

# Go build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the application
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

test: ## Run all tests
	@echo "Running tests..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Test coverage report: coverage.html"

test-contract: ## Run contract tests only
	@echo "Running contract tests..."
	@go test -v ./tests/contract/...

test-integration: ## Run integration tests only
	@echo "Running integration tests..."
	@go test -v ./tests/integration/...

test-unit: ## Run unit tests only
	@echo "Running unit tests..."
	@go test -v ./tests/unit/...

test-performance: ## Run performance tests
	@echo "Running performance tests..."
	@go test -v ./tests/performance/...

lint: ## Run golangci-lint
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	@golangci-lint run ./...

format: ## Format code with gofmt and goimports
	@echo "Formatting code..."
	@gofmt -s -w .
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	@goimports -w .

clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@go clean

install: build ## Install the binary to $GOPATH/bin
	@echo "Installing $(BINARY_NAME)..."
	@go install $(LDFLAGS) $(MAIN_PATH)

run: build ## Build and run the application
	@echo "Running $(BINARY_NAME)..."
	@$(BUILD_DIR)/$(BINARY_NAME)

deps: ## Download and verify dependencies
	@echo "Downloading dependencies..."
	@go mod download
	@go mod verify
	@go mod tidy

deps-update: ## Update dependencies
	@echo "Updating dependencies..."
	@go get -u ./...
	@go mod tidy

check: lint test ## Run linting and tests

release-build: ## Build release version with optimizations
	@echo "Building release version..."
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 go build -a -installsuffix cgo $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Release build complete: $(BUILD_DIR)/$(BINARY_NAME)"

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .

# Development helpers
dev-setup: ## Set up development environment
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/charmbracelet/teatest@latest
	@echo "Development tools installed"

watch: ## Watch for changes and rebuild (requires entr)
	@echo "Watching for changes..."
	@which entr > /dev/null || (echo "entr not found. Install with your package manager" && exit 1)
	@find . -name "*.go" | entr -r make run

# Cross-platform build targets
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64
RELEASE_DIR := release

build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(RELEASE_DIR)
	@$(foreach PLATFORM,$(PLATFORMS), \
		GOOS=$(word 1,$(subst /, ,$(PLATFORM))) \
		GOARCH=$(word 2,$(subst /, ,$(PLATFORM))) \
		CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(RELEASE_DIR)/$(BINARY_NAME)-$(PLATFORM)$(if $(findstring windows,$(PLATFORM)),.exe,) \
		$(MAIN_PATH) && echo "Built $(BINARY_NAME)-$(PLATFORM)" || exit 1;)

package-releases: build-all ## Package releases for distribution
	@echo "Packaging releases..."
	@cd $(RELEASE_DIR) && \
	for binary in $(BINARY_NAME)-*; do \
		if [ -f "$$binary" ]; then \
			platform=$$(echo "$$binary" | sed 's/$(BINARY_NAME)-//'); \
			if [[ "$$binary" == *".exe" ]]; then \
				platform=$$(echo "$$platform" | sed 's/.exe$$//'); \
				zip "$$binary.zip" "$$binary"; \
			else \
				tar -czf "$$binary.tar.gz" "$$binary"; \
			fi; \
			echo "Packaged $$binary"; \
		fi; \
	done

release: clean package-releases ## Create a complete release
	@echo "Release $(VERSION) created in $(RELEASE_DIR)/"
	@ls -la $(RELEASE_DIR)/

install-system: build ## Install to /usr/local/bin (requires sudo)
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "Installation complete. You can now run '$(BINARY_NAME)' from anywhere."

uninstall-system: ## Uninstall from /usr/local/bin
	@echo "Uninstalling $(BINARY_NAME)..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(BINARY_NAME) uninstalled."

# Show build info
info: ## Show build information
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Go Version: $(GO_VERSION)"