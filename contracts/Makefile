# Trace Wallet - Smart Contract Project Makefile
# This Makefile provides targets for building, testing, deploying, and cleaning the project.

.PHONY: all help build test test-coverage test-gas test-watch deploy-base-test deploy-base deploy-polygon-test deploy-polygon clean lint install

# Default target
all: install build test

# Help message
help:
	@echo "Trace Wallet Makefile Commands:"
	@echo "  make install            - Install project dependencies"
	@echo "  make build              - Compile all smart contracts"
	@echo "  make test               - Run all tests"
	@echo "  make test-coverage      - Run tests with coverage report"
	@echo "  make test-gas           - Run tests with gas reporting"
	@echo "  make test-watch         - Run tests in watch mode"
	@echo "  make lint               - Check code for style and best practices"
	@echo "  make deploy-base-test   - Deploy to Base testnet"
	@echo "  make deploy-base        - Deploy to Base mainnet"
	@echo "  make deploy-polygon-test - Deploy to Polygon zkEVM testnet"
	@echo "  make deploy-polygon     - Deploy to Polygon zkEVM mainnet"
	@echo "  make clean              - Remove build artifacts and cache"
	@echo "  make all                - Build and test the project"
	@echo "  make help               - Display this help message"

# Install dependencies
install:
	@echo "Installing project dependencies..."
	npm install

# Build target
build:
	@echo "Building smart contracts..."
	npm run compile

# Test targets
test:
	@echo "Running tests..."
	npm run test

test-coverage:
	@echo "Running tests with coverage..."
	npm run test:coverage

test-gas:
	@echo "Running tests with gas reporting..."
	npm run test:gas

test-watch:
	@echo "Running tests in watch mode..."
	npm run test:watch

# Lint target (add hardhat-solhint plugin to package.json if not already there)
lint:
	@echo "Linting solidity files..."
	npx solhint 'solidity/**/*.sol'

# Deployment targets
deploy-base-test:
	@echo "Deploying to Base testnet..."
	npm run deploy:base-test

deploy-base:
	@echo "Deploying to Base mainnet..."
	@echo "⚠️  WARNING: This will deploy to MAINNET. Are you sure? (y/n) " && read ans && [ $${ans:-n} = y ]
	npm run deploy:base

deploy-polygon-test:
	@echo "Deploying to Polygon zkEVM testnet..."
	npm run deploy:polygon-test

deploy-polygon:
	@echo "Deploying to Polygon zkEVM mainnet..."
	@echo "⚠️  WARNING: This will deploy to MAINNET. Are you sure? (y/n) " && read ans && [ $${ans:-n} = y ]
	npm run deploy:polygon

# Clean target
clean:
	@echo "Cleaning build artifacts and cache..."
	rm -rf artifacts
	rm -rf cache
	rm -rf coverage
	rm -rf coverage.json
	rm -rf typechain
	rm -rf typechain-types
	rm -rf node_modules
	@echo "Build artifacts and cache removed."
