# Makefile for vault0 project

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOINSTALL = $(GOCMD) install
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod
GOCLEAN = $(GOCMD) clean

# Binary names
SERVER_BIN = vault0
GENKEY_BIN = genkey

# Build directory
BUILD_DIR = bin

# Source directories
SERVER_SRC = ./cmd/server
GENKEY_SRC = ./cmd/genkey

# Package name
PACKAGE = vault0

.PHONY: all build clean test deps install server genkey

# Default target
all: clean build

# Build all binaries
build: server genkey

# Build server binary
server:
	@echo "Building server binary..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(SERVER_BIN) $(SERVER_SRC)

# Build genkey binary
genkey:
	@echo "Building genkey binary..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(GENKEY_BIN) $(GENKEY_SRC)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	$(GOCLEAN) -cache

# Run tests
test:
	$(GOTEST) -v ./...

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install binaries
install:
	@echo "Installing binaries..."
	$(GOINSTALL) $(SERVER_SRC)
	$(GOINSTALL) $(GENKEY_SRC)

# Run server
run-server: server
	@echo "Running server..."
	@$(BUILD_DIR)/$(SERVER_BIN)

# Run genkey
run-genkey: genkey
	@echo "Running genkey..."
	@$(BUILD_DIR)/$(GENKEY_BIN) 