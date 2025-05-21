# Trace Wallet - Smart Contract Project Makefile
# This Makefile provides targets for building, testing, deploying, and cleaning the project.

.PHONY: all help build test test-coverage test-gas test-watch deploy-base-test deploy-base deploy-polygon-test deploy-polygon deploy-localhost clean lint install \
        query-brlt manage-roles mint-tokens burn-tokens toggle-pause manage-blacklist get-status approve-tokens transfer-tokens prepare-upgrade-brlt apply-upgrade-brlt

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
	@echo "  make deploy-localhost   - Deploy to localhost"
	@echo "  make query-brlt ARGS='...' - Query BRLT contract info (pass --proxy, --account)"
	@echo "  make manage-roles ARGS='...' - Grant/revoke BRLT roles (pass --proxy, --role, --account, --action)"
	@echo "  make mint-tokens ARGS='...' - Mint BRLT tokens (pass --proxy, --recipient, --amount)"
	@echo "  make burn-tokens ARGS='...' - Burn BRLT tokens (pass --proxy, --from, --amount)"
	@echo "  make toggle-pause ARGS='...' - Pause/unpause BRLT contract (pass --proxy, --action)"
	@echo "  make manage-blacklist ARGS='...' - Blacklist/unblacklist account (pass --proxy, --account, --action)"
	@echo "  make get-status ARGS='...' - Get comprehensive BRLT contract status (pass --proxy and other optional checks)"
	@echo "  make approve-tokens ARGS='...' - Approve BRLT token spending (pass --proxy, --spender, --amount)"
	@echo "  make transfer-tokens ARGS='...' - Transfer BRLT tokens (pass --proxy, --to, --amount)"
	@echo "  make prepare-upgrade-brlt ARGS='...' - Prepare BRLT contract upgrade (pass --proxy, --new-impl-name)"
	@echo "  make apply-upgrade-brlt ARGS='...' - Apply BRLT contract upgrade (pass --proxy, --new-impl-address)"
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

deploy-localhost:
	@echo "Deploying to localhost..."
	npm run deploy:localhost

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

# Interaction script targets
# Users should pass script-specific arguments via ARGS variable
# Example: make query-brlt ARGS='--proxy 0x123... --account 0x456... --network localhost'

query-brlt:
	@echo "Querying BRLT contract. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <address> --account <address>"
	npm run query:brlt -- $(ARGS)

manage-roles:
	@echo "Managing BRLT roles. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --role <ROLE_NAME> --account <target_acc> --action <grant|revoke>"
	npm run manage-roles -- $(ARGS)

mint-tokens:
	@echo "Minting BRLT tokens. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --recipient <recv_acc> --amount <num>"
	npm run mint-tokens -- $(ARGS)

burn-tokens:
	@echo "Burning BRLT tokens. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --from <from_acc> --amount <num>"
	@echo "Note: The 'from_acc' must have approved the script signer before running."
	npm run burn-tokens -- $(ARGS)

toggle-pause:
	@echo "Toggling BRLT contract pause state. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --action <pause|unpause>"
	npm run toggle-pause -- $(ARGS)

manage-blacklist:
	@echo "Managing BRLT account blacklist. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --account <target_acc> --action <blacklist|unblacklist>"
	npm run manage-blacklist -- $(ARGS)

get-status:
	@echo "Getting BRLT contract status. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr>"
	@echo "Optional ARGS: --check-balances <acc1,acc2> --check-blacklist <acc1,acc2> --check-allowances <owner1:spender1,owner2:spender2>"
	npm run get-status -- $(ARGS)

approve-tokens:
	@echo "Approving BRLT token spending. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --spender <spender_addr> --amount <num>"
	npm run approve-tokens -- $(ARGS)

transfer-tokens:
	@echo "Transferring BRLT tokens. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --to <recipient_addr> --amount <num>"
	npm run transfer-tokens -- $(ARGS)

prepare-upgrade-brlt:
	@echo "Preparing BRLT contract upgrade. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --new-impl-name <NewContractName>"
	npm run prepare-upgrade-brlt -- $(ARGS)

apply-upgrade-brlt:
	@echo "Applying BRLT contract upgrade. Pass script arguments via ARGS."
	@echo "Required ARGS: --proxy <addr> --new-impl-address <new_implementation_address>"
	npm run apply-upgrade-brlt -- $(ARGS)

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
