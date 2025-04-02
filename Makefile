# Makefile for vault0 project

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOCLEAN = $(GOCMD) clean
GOINSTALL = $(GOCMD) install

# Build flags
LDFLAGS = -ldflags="-s -w"
DEBUGFLAGS = -gcflags="all=-N -l"

# Binary names
SERVER_BIN = vault0
GENKEY_BIN = genkey
VERIFY_TOKENS_BIN = verify-tokens

# Build directory
BUILD_DIR = bin

# Source directories
SERVER_SRC = ./cmd/server
GENKEY_SRC = ./cmd/genkey
VERIFY_TOKENS_SRC = ./cmd/verify-tokens

# UI directory
UI_DIR = ./ui

# Contracts directory
CONTRACTS_DIR = ./contracts

# Package name
PACKAGE = vault0

.PHONY: all build clean server-build server-test server-test-coverage server-deps server-build-debug genkey-build genkey-install server server-clean git-reset git-status git-pull git-push ui-build ui-deps ui ui-start ui-lint ui-clean contracts contracts-deps contracts-test contracts-test-coverage contracts-lint contracts-clean contracts-deploy-base-test contracts-deploy-base contracts-deploy-polygon-test contracts-deploy-polygon count-lines count-lines-ui count-lines-backend count-lines-contracts count-lines-source count-lines-tests git-diff-setup verify-tokens verify-tokens-build swag-install server-docs delve-install wire-install wire server-install

# Count lines of code in the project
count-lines:
	@./scripts/count_lines.sh

# Default target
all: clean build

# Build all binaries
build: server-build genkey-build verify-tokens-build ui-build contracts-build

# Build server binary
server-build: wire
	@echo "Building server binary..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(SERVER_BIN) $(SERVER_SRC)

# Build server binary with debug symbols
server-build-debug: wire
	@echo "Building server binary with debug symbols..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(DEBUGFLAGS) -o $(BUILD_DIR)/$(SERVER_BIN) $(SERVER_SRC)

# Build genkey binary
genkey-build:
	@echo "Building genkey binary..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(GENKEY_BIN) $(GENKEY_SRC)

# Build verify-tokens binary
verify-tokens-build:
	@echo "Building verify-tokens binary..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(VERIFY_TOKENS_BIN) $(VERIFY_TOKENS_SRC)

# Run verify-tokens
verify-tokens: verify-tokens-build
	@echo "Running verify-tokens..."
	@$(BUILD_DIR)/$(VERIFY_TOKENS_BIN)

# Run tests
server-test:
	$(GOTEST) -v ./...

# Run tests with coverage
server-test-coverage:
	@echo "Running server tests with coverage..."
	$(GOTEST) -v -cover ./...
	@echo "For detailed coverage report:"
	@echo "$(GOTEST) -coverprofile=coverage.out ./... && go tool cover -html=coverage.out"

# Install dependencies
server-deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install all server development tools
server-install: wire-install delve-install swag-install
	@echo "All server development tools installed."
	$(GOMOD) tidy

# Run server
server: server-build
	@echo "Running server..."
	@$(BUILD_DIR)/$(SERVER_BIN)

# Clean build artifacts
server-clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN) -cache

# Install Wire tool
wire-install:
	@echo "Installing Wire tool..."
	$(GOINSTALL) github.com/google/wire/cmd/wire@latest

# Install Delve debugger tool
delve-install:
	@echo "Installing Delve debugger tool..."
	$(GOINSTALL) github.com/go-delve/delve/cmd/dlv@latest

# Generate Wire code
wire:
	@echo "Generating Wire dependency injection code..."
	cd internal/wire && wire

# Install Swaggo tool
swag-install:
	@echo "Installing Swaggo tool..."
	$(GOGET) github.com/swaggo/swag/cmd/swag

# Generate server-docs documentation
server-docs: 
	@echo "Generating server-docs OpenAPI documentation..."
	$$($(GOCMD) env GOPATH)/bin/swag init -g internal/api/docs.go -o internal/api/docs

# Combined clean target
clean: server-clean ui-clean
	@echo "All artifacts cleaned"
	
# Git commands
git-reset:
	@echo "Resetting git repository to last commit..."
	git reset --hard HEAD && git clean -fd

# Setup git diff with cat to prevent terminal from getting stuck
git-diff-setup:
	@echo "Setting up git diff with cat as pager for the current repository..."
	git config pager.diff cat
	git config pager.show cat
	@echo "Git diff is now set up to use cat instead of less. This prevents the terminal from getting stuck in pager mode."

# UI commands
ui-build:
	@echo "Building UI for production..."
	cd $(UI_DIR) && npm run build

ui-deps:
	@echo "Installing UI dependencies..."
	cd $(UI_DIR) && npm install

ui:
	@echo "Starting UI development server..."
	cd $(UI_DIR) && npm run dev

ui-start:
	@echo "Starting UI production server..."
	cd $(UI_DIR) && npm run start

ui-lint:
	@echo "Linting UI code..."
	cd $(UI_DIR) && npm run lint

ui-clean:
	@echo "Cleaning UI build artifacts..."
	rm -rf $(UI_DIR)/.next $(UI_DIR)/dist $(UI_DIR)/.turbo $(UI_DIR)/.vercel

# Contract commands
contracts:
	@echo "Building smart contracts..."
	cd $(CONTRACTS_DIR) && npm run compile

contracts-deps:
	@echo "Installing contract dependencies..."
	cd $(CONTRACTS_DIR) && npm install

contracts-test:
	@echo "Running contract tests..."
	cd $(CONTRACTS_DIR) && npm run test

contracts-test-coverage:
	@echo "Running contract tests with coverage..."
	cd $(CONTRACTS_DIR) && npm run test:coverage

contracts-lint:
	@echo "Linting contract code..."
	cd $(CONTRACTS_DIR) && npx solhint 'solidity/**/*.sol'

contracts-clean:
	@echo "Cleaning contract build artifacts..."
	cd $(CONTRACTS_DIR) && rm -rf artifacts cache coverage coverage.json typechain typechain-types

contracts-deploy-base-test:
	@echo "Deploying contracts to Base testnet..."
	cd $(CONTRACTS_DIR) && npm run deploy:base-test

contracts-deploy-base:
	@echo "Deploying contracts to Base mainnet..."
	cd $(CONTRACTS_DIR) && npm run deploy:base

contracts-deploy-polygon-test:
	@echo "Deploying contracts to Polygon zkEVM testnet..."
	cd $(CONTRACTS_DIR) && npm run deploy:polygon-test

contracts-deploy-polygon:
	@echo "Deploying contracts to Polygon zkEVM mainnet..."
	cd $(CONTRACTS_DIR) && npm run deploy:polygon
