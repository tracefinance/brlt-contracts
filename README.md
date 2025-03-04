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

The Go backend provides:

- RESTful API for wallet management
- Database integration for transaction history
- Configuration management
- Migration support
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
├── cmd/                    # Command-line applications
│   ├── server/             # Main server application
│   └── genkey/             # Encryption key generation utility
├── internal/               # Private application code
│   ├── api/                # API handlers
│   ├── config/             # Configuration management
│   ├── db/                 # Database access layer
│   └── keystore/         # Key management module
├── migrations/             # Database migrations
├── contracts/              # Smart contract codebase
│   ├── solidity/           # Solidity contracts
│   │   └── MultiSigWallet.sol  # Main wallet contract
│   ├── test/               # Contract test suite
│   └── scripts/            # Deployment scripts
└── ui/                     # Frontend application
    ├── public/             # Static assets
    └── src/                # React components and pages
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
