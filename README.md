# Vault0: Multi-Signature Crypto Wallet

Vault0 is a secure, dual-signature cryptocurrency wallet smart contract system with a robust recovery mechanism. It implements a multi-signature wallet that requires two authorizations (client and manager) to withdraw funds, with additional security features such as timelock-based recovery and token whitelisting.

## Project Overview

The Vault0 system consists of:

1. **Smart Contracts**: Solidity-based multi-signature wallet implementation
2. **Backend API**: Go-based server providing wallet management features
3. **Frontend UI**: Nuxt 3 web interface for user interactions

## Key Features

- **Wallets, keys, users and signers management**: Comprehensive management of crypto assets and access control
- **Vaults with multi-sig smart contracts**:
  * Recovery mechanism with 72-hour timelock process
  * Token whitelisting and management
- **EVM compatible**: Supports Ethereum, Base, Arbitrum, Binance Chain, and Tron networks
- **Realtime transaction monitoring**: Track and verify transactions as they happen
- **Full dashboard with web3 support for signatures**: Complete UI for managing wallets and signing transactions

## Progress

- [x] Backend structure in golang with three-layer architecture:
  * Layer 1 (Core): Key management, signing, blockchain integration, price feed, database access, cryptography utilities
  * Layer 2 (Services): User service, wallet service, blockchain service, transaction processing
  * Layer 3 (API): RESTful endpoints, middleware, request/response handling
- [x] Solidity smart contract with:
  * Multi-signature support requiring dual authorization
  * 72-hour timelock recovery mechanism
  * Native coin and ERC20 token support
  * Token whitelisting functionality
  * Gas-optimized operations
- [ ] Full featured dashboard with web3 signing
  - [x] Wallet interface with balances and real-time transactions
  - [x] User, wallet, signer, key and token management
  - [ ] Vault management
  - [ ] Vault interface with balances and real-time transactions
  - [ ] Withdraw interface with signing support
- [ ] Authentication

## Future Roadmap

- [ ] Bridge support (create one vault and send/receive from any supported blockchain)
- [ ] Swap support
- [ ] Public API

## Technical Architecture

### Smart Contract Layer

The core of Vault0 is the `MultiSigWallet.sol` contract, which:

- Implements a dual-signature system for withdrawals
- Provides a 72-hour recovery timelock mechanism
- Supports both native coins (ETH) and ERC20 tokens
- Includes whitelisting and token management functions
- Uses OpenZeppelin contracts for security best practices

### Backend API (Go)

The Go backend follows a three-layer architecture:

#### Layer 1: Core/Infrastructure Layer
Composed of foundational modules that serve as building blocks for the entire system, organized in the `internal/core` directory:
- **Database Access**: Provides database connectivity and query execution
- **Wallet Operations**: Low-level wallet functionality
- **Keystore**: Secure storage and retrieval of cryptographic keys
- **Blockchain Interaction**: Core blockchain communication
- **Cryptography Utilities**: Encryption, decryption, and hashing functions
- **Contract Interaction**: Smart contract operation abstraction
- **Key Generation**: Cryptographic key generation utilities

#### Layer 2: Service Layer
Contains business logic modules organized by domain in the `internal/services` directory, each encapsulating:
- Domain-specific data models
- Business operations and validation rules
- Repository interfaces for data access

Services include:
- User management
- Wallet administration
- Authentication and authorization
- Transaction processing
- Blockchain service integration

#### Layer 3: Communication Layer
Exposes the service layer functionality externally through the `internal/api` directory:
- RESTful API endpoints
- Request/response handling
- Middleware (authentication, logging, error handling)
- Route management

The backend also provides:
- Database integration for transaction history
- Secure key management with AES-GCM encryption
- Modular architecture for multiple key storage mechanisms
- OAuth2 support for secure authentication

### Frontend (Nuxt 3)

The UI is built with:

- Nuxt 3 (Vue.js framework)
- TailwindCSS with shadcn/ui components
- TypeScript for type safety
- Token-based pagination

## Supported Networks

Vault0 is designed to work on EVM-compatible networks including:

- **Ethereum**: 
  - Mainnet
  - Testnets (Sepolia, Goerli)
- **Base**: 
  - Mainnet (Chain ID: 8453)
  - Testnet (Chain ID: 84531)
- **Arbitrum**: 
  - Mainnet
  - Testnet
- **Binance Smart Chain**: 
  - Mainnet
  - Testnet
- **Tron**: 
  - Mainnet
  - Testnet

## Technical Stack

### Smart Contracts
- **Language**: Solidity ^0.8.28
- **Framework**: Hardhat
- **Libraries**: OpenZeppelin Contracts
- **Testing**: Hardhat test suite with high coverage requirements (90%+)

### Backend
- **Language**: Go
- **Database**: SQLite
- **Web Framework**: Gin
- **API**: RESTful API with token-based pagination
- **Configuration**: Environment-based configuration
- **Encryption**: AES-GCM for key encryption
- **Key Management**: Module-based architecture supporting multiple implementations

### Frontend
- **Framework**: Nuxt 3
- **UI Library**: Vue.js with Composition API
- **Components**: shadcn/ui components (located in `~/components/ui/`)
- **Styling**: TailwindCSS
- **Language**: TypeScript
- **API Integration**: Nuxt plugin with dedicated client modules

## Security Features

- **Dual-Signature Requirement**: Prevents single-point compromise
- **Timelock Recovery**: 72-hour delay for recovery operations
- **Withdrawal Expiration**: 24-hour expiration for withdrawal requests
- **Reentrancy Protection**: Guards against reentrancy attacks
- **Token Whitelisting**: Controlled token support
- **Event Logging**: Comprehensive event logging for all operations
- **Encrypted Key Storage**: Private keys stored using AES-GCM encryption
- **Environment-Based Encryption Key**: Database encryption key provided via environment variables

## Project Structure

```
vault0/
├── ui/              # Frontend Nuxt application
│   ├── components/  # Vue components (auto-imported)
│   │   └── ui/      # shadcn/ui components
│   ├── composables/ # Vue composables (auto-imported)
│   ├── lib/         # Utility functions
│   ├── pages/       # Nuxt pages
│   ├── plugins/     # Nuxt plugins (including API client)
│   ├── public/      # Static assets
│   ├── server/      # Server-side code
│   └── types/       # TypeScript type definitions
├── contracts/       # Smart contracts
│   ├── solidity/    # Contract source files
│   ├── test/        # Contract tests
│   └── scripts/     # Deployment and utility scripts
├── cmd/             # Command-line applications
├── internal/        # Backend application code
│   ├── api/         # Communication Layer (Layer 3)
│   ├── core/        # Core/Infrastructure Layer (Layer 1)
│   └── services/    # Service Layer (Layer 2)
└── migrations/      # Database migrations
```

## Development Setup

### Prerequisites

- Visual Studio Code with Dev Containers plugin or Cursor Code Editor
- Docker and Docker Compose
- Git

### Development Container Setup

This project uses dev containers for consistent development environments across all team members.

1. Install Visual Studio Code or Cursor Code Editor with the Dev Containers extension
2. Clone the repository
3. Open the project in your editor
4. When prompted, click "Reopen in Container" or use the command palette to select "Dev Containers: Reopen in Container"
5. Wait for the container to build (this may take a few minutes the first time)

### Key Development Commands

```bash
# Install all dependencies
make deps

# Build and run the backend server
make server

# Build and start the frontend UI
make ui

# Compile and test smart contracts
make contracts
```

### Smart Contract Development

```bash
# Install dependencies
make contracts-deps

# Compile contracts
make contracts

# Run tests
make contracts-test

# Run coverage tests
make contracts-test-coverage

# Lint contracts
make contracts-lint

# Clean contract artifacts
make contracts-clean
```

### Backend Development

```bash
# Install Go dependencies
make server-deps

# Generate an encryption key for development
make genkey
./bin/genkey

# Set the encryption key in your environment
export DB_ENCRYPTION_KEY='generated-key-from-above-command'

# Build server
make server

# Run server
make server-dev

# Run tests
make server-test
```

### Frontend Development

```bash
# Install dependencies (always run from ui/ directory)
cd ui && npm install

# Or use the make command
make ui-deps

# Start development server
make ui-dev

# Build for production
make ui

# Start production server
make ui-start

# Lint code
make ui-lint

# Clean UI build artifacts
make ui-clean
```

## Deployment

### Smart Contract Deployment

```bash
# Deploy to Base testnet
make contracts-deploy-base-test

# Deploy to Base mainnet
make contracts-deploy-base

# Deploy to Polygon zkEVM testnet
make contracts-deploy-polygon-test

# Deploy to Polygon zkEVM mainnet
make contracts-deploy-polygon
```

### Version Control

```bash
# Reset to last commit (caution: removes all untracked files)
make git-reset
```

## Contributors

The Vault0 team

## License

This project is licensed under the MIT License.

Copyright (c) 2025 Vault0

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
