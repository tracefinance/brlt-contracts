# Vault0: Multi-Signature Crypto Wallet

Vault0 is a secure, dual-signature cryptocurrency wallet smart contract system with a robust recovery mechanism. It implements a multi-signature wallet that requires two authorizations (client and manager) to withdraw funds, with additional security features such as timelock-based recovery and token whitelisting.

## Project Overview

The Vault0 system consists of:

1. **Smart Contracts**: Solidity-based multi-signature wallet implementation
2. **Backend API**: Go-based server providing wallet management features
3. **Frontend UI**: Next.js with React 19 web interface for user interactions

## Key Features

- **Dual-Signature Security**: Requires both manager and client signatures to authorize withdrawals
- **Recovery Mechanism**: 72-hour timelock recovery process to recover funds if needed
- **Token Support**: Configurable token support with whitelisting functionality
- **Gas Optimization**: Optimized for efficient gas usage on Ethereum-compatible chains
- **Cross-Chain Compatibility**: Supports Base and Polygon zkEVM networks
- **Secure Key Management**: Built-in key management service with encrypted storage

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
Composed of foundational modules that serve as building blocks for the entire system:
- **Database Access**: Provides database connectivity and query execution
- **Wallet Operations**: Low-level wallet functionality
- **Keystore**: Secure storage and retrieval of cryptographic keys
- **Blockchain Interaction**: Core blockchain communication
- **Cryptography Utilities**: Encryption, decryption, and hashing functions
- **Configuration Management**: Application configuration

#### Layer 2: Service Layer
Contains business logic modules organized by domain, each encapsulating:
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
Exposes the service layer functionality externally:
- RESTful API endpoints
- Request/response handling
- Middleware (authentication, logging, error handling)
- Route management

The backend also provides:
- Database integration for transaction history
- Secure key management with AES-GCM encryption
- Modular architecture for multiple key storage mechanisms

### Frontend (React/Next.js)

The UI is built with:

- Next.js 15.2 with React 19
- TailwindCSS for styling
- TypeScript for type safety

## Supported Networks

Vault0 is designed to work on:

- **Base**: 
  - Mainnet (Chain ID: 8453)
  - Testnet (Chain ID: 84531)
- **Polygon zkEVM**: 
  - Mainnet (Chain ID: 1101)
  - Testnet (Chain ID: 1442)

## Technical Stack

### Smart Contracts
- **Language**: Solidity ^0.8.28
- **Framework**: Hardhat
- **Libraries**: OpenZeppelin Contracts
- **Testing**: Hardhat test suite with high coverage requirements (90%+)

### Backend
- **Language**: Go
- **Database**: SQLite
- **API**: RESTful API
- **Configuration**: Environment-based configuration
- **Encryption**: AES-GCM for key encryption
- **Key Management**: Module-based architecture supporting multiple implementations

### Frontend
- **Framework**: Next.js 15.2
- **UI Library**: React 19
- **Styling**: TailwindCSS 4
- **Language**: TypeScript

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
├── cmd/                                # Command-line applications
│   ├── server/                         # Main server application
│   │   └── main.go                     # Application entry point
│   └── genkey/                         # Encryption key generation utility
│       └── main.go                     # Key generation entry point
├── internal/                           # Private application code
│   ├── api/                            # Communication Layer (Layer 3)
│   │   ├── handlers/                   # API request handlers by domain
│   │   │   ├── user/                   # User-related endpoints
│   │   │   │   ├── handler.go          # User handler implementation
│   │   │   │   └── dto.go              # User request/response structures
│   │   │   ├── wallet/                 # Wallet-related endpoints
│   │   │   │   ├── handler.go          # Wallet handler implementation
│   │   │   │   └── dto.go              # Wallet request/response structures
│   │   │   ├── auth/                   # Authentication endpoints
│   │   │   │   ├── handler.go          # Auth handler implementation
│   │   │   │   └── dto.go              # Auth request/response structures
│   │   │   └── transaction/            # Transaction endpoints
│   │   │       ├── handler.go          # Transaction handler implementation
│   │   │       └── dto.go              # Transaction request/response structures
│   │   ├── middleware/                 # Request middleware components
│   │   │   ├── auth.go                 # Authentication middleware
│   │   │   ├── logging.go              # Request logging middleware
│   │   │   └── errors.go               # Error handling middleware
│   │   └── server.go                   # API server setup and configuration
│   ├── services/                       # Service Layer (Layer 2)
│   │   ├── user/                       # User management domain
│   │   │   ├── model.go                # User data models
│   │   │   ├── service.go              # User service implementation
│   │   │   └── repository.go           # User data access interface and implementation
│   │   ├── wallet/                     # Wallet operations domain
│   │   │   ├── model.go                # Wallet data models
│   │   │   ├── service.go              # Wallet service implementation
│   │   │   └── repository.go           # Wallet data access interface and implementation
│   │   ├── auth/                       # Authentication domain
│   │   │   ├── model.go                # Auth data models
│   │   │   ├── service.go              # Auth service implementation
│   │   │   └── jwt.go                  # JWT token utilities
│   │   ├── blockchain/                 # Blockchain operations domain
│   │   │   ├── model.go                # Blockchain data models
│   │   │   ├── service.go              # Blockchain service implementation
│   │   │   └── client.go               # Blockchain client interactions
│   │   └── transaction/                # Transaction processing domain
│   │       ├── model.go                # Transaction data models
│   │       ├── service.go              # Transaction service implementation
│   │       └── repository.go           # Transaction data access interface and implementation
│   ├── db/                             # Core Layer (Layer 1) - Database access
│   ├── keystore/                       # Core Layer (Layer 1) - Key management
│   ├── blockchain/                     # Core Layer (Layer 1) - Blockchain interaction
│   ├── crypto/                         # Core Layer (Layer 1) - Cryptography utilities
│   ├── wallet/                         # Core Layer (Layer 1) - Wallet operations
│   ├── config/                         # Core Layer (Layer 1) - Configuration
│   └── types/                          # Shared type definitions
├── migrations/                         # Database migrations
├── contracts/                          # Smart contract codebase
│   ├── solidity/                       # Solidity contracts
│   │   ├── MultiSigWallet.sol          # Main wallet contract
│   │   ├── interfaces/                 # Contract interfaces
│   │   │   ├── IERC20.sol              # ERC20 token interface
│   │   │   └── IMultiSigWallet.sol     # Wallet interface
│   │   └── libraries/                  # Contract libraries
│   │       └── AddressUtils.sol        # Address utility functions
│   ├── test/                           # Contract test suite
│   │   ├── MultiSigWallet.test.js      # Wallet contract tests
│   │   └── helpers/                    # Test helpers
│   │       └── setup.js                # Test setup utilities
│   ├── scripts/                        # Deployment scripts
│   │   ├── deploy.js                   # Main deployment script
│   │   └── verify.js                   # Contract verification script
│   └── hardhat.config.js               # Hardhat configuration
└── ui/                                 # Frontend application
    ├── public/                         # Static assets
    │   ├── favicon.ico                 # Site favicon
    │   └── images/                     # Image assets
    └── src/                            # React components and pages
        ├── components/                 # Reusable React components
        │   ├── common/                 # Common UI components
        │   ├── wallet/                 # Wallet-related components
        │   └── layout/                 # Layout components
        ├── pages/                      # Next.js pages
        │   ├── index.tsx               # Homepage
        │   ├── wallets/                # Wallet pages
        │   ├── auth/                   # Authentication pages
        │   └── api/                    # API routes
        ├── hooks/                      # React hooks
        ├── lib/                        # Utility libraries
        ├── context/                    # React context providers
        ├── styles/                     # CSS and style files
        └── types/                      # TypeScript type definitions
```

## Development Setup

### Prerequisites

- Node.js ≥ 14.0.0
- npm ≥ 6.0.0
- Go ≥ 1.16
- Solidity ^0.8.28
- Hardhat

### Quick Start

```bash
# Install all dependencies
make server-deps ui-deps contracts-deps

# Build all components
make build

# Clean all artifacts
make clean
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
# Install dependencies
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

## License

ISC

## Contributors

The Vault0 team
