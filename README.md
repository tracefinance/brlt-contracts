# Multi-Signature Crypto Wallet

[![Smart Contract CI](https://github.com/tracefinance/fx-multisig-wallet/actions/workflows/compile-and-test.yml/badge.svg)](https://github.com/tracefinance/fx-multisig-wallet/actions/workflows/compile-and-test.yml)
[![codecov](https://codecov.io/gh/tracefinance/fx-multisig-wallet/graph/badge.svg?token=fsCmAuBO0b)](https://codecov.io/gh/tracefinance/fx-multisig-wallet)

A secure Ethereum-compatible wallet smart contract that requires dual signatures for withdrawals and implements a secure recovery mechanism.

## Features

### Dual-Signature Mechanism
- Requires both manager and client signatures to withdraw funds
- Support for native coins (ETH) and ERC-20 tokens
- 24-hour expiration on withdrawal requests
- Comprehensive event logging

### Recovery Mechanism
- 72-hour timelock recovery process
- Only manager can initiate recovery
- Client can cancel recovery within timelock period
- Separate functions for native coin and token recovery
- Funds are sent to a predefined recovery address

## Project Structure

```
├── solidity/
│   ├── MultiSigWallet.sol   # Main wallet contract
│   └── MockToken.sol        # Test ERC-20 token
├── test/
│   └── MultiSigWallet.js    # Test suite
├── scripts/
│   └── deploy.js            # Deployment script
├── ignition/
│   └── modules/
│       └── MultiSigWallet.js # Deployment module
└── hardhat.config.js        # Hardhat configuration
```

## Prerequisites

- Node.js >= 14.0.0
- npm >= 6.0.0

## Installation

```shell
# Install dependencies
npm install
```

## Configuration

1. For development and testing, copy `.env.example` and update values if needed:
```shell
cp .env.example .env
```

2. For production deployment, create `.env` with your configuration:
```env
# Network RPC URLs
BASE_RPC_URL=your_base_rpc_url
BASE_TESTNET_RPC_URL=your_base_testnet_rpc_url
POLYGON_ZKEVM_RPC_URL=your_polygon_rpc_url
POLYGON_ZKEVM_TESTNET_RPC_URL=your_polygon_testnet_rpc_url

# API Keys
ETHERSCAN_API_KEY=your_etherscan_api_key
BASESCAN_API_KEY=your_basescan_api_key
POLYGONSCAN_API_KEY=your_polygonscan_api_key

# Private Key - Be careful with this!
PRIVATE_KEY=your_private_key_here

# Contract Parameters
CLIENT_ADDRESS=your_client_address
RECOVERY_ADDRESS=your_recovery_address
```

## Testing

```shell
# Run all tests
npm test

# Run tests with coverage report
npm run test:coverage

# Run tests with gas reporting
npm run test:gas

# Run tests in watch mode (development)
npm run test:watch
```

## Deployment

The contract can be deployed to various networks:

### Base Network

```shell
# Deploy to Base testnet
npm run deploy:base-test

# Deploy to Base mainnet
npm run deploy:base
```

### Polygon zkEVM

```shell
# Deploy to Polygon zkEVM testnet
npm run deploy:polygon-test

# Deploy to Polygon zkEVM mainnet
npm run deploy:polygon
```

### Deployment Verification

The deployment script will automatically:
1. Deploy the contract
2. Wait for confirmations
3. Verify the contract on the respective block explorer

## Network Configurations

The project supports multiple networks:

### Base
- Mainnet Chain ID: 8453
- Testnet Chain ID: 84531
- Explorer: https://basescan.org

### Polygon zkEVM
- Mainnet Chain ID: 1101
- Testnet Chain ID: 1442
- Explorer: https://zkevm.polygonscan.com

## Usage

### Setting Up
1. Deploy the contract with client and recovery addresses
2. The deploying address becomes the manager

### Withdrawing Funds
1. Either manager or client initiates withdrawal request
2. The other party signs the request
3. Funds are automatically transferred after both signatures

### Recovery Process
1. Manager initiates recovery
2. Client has 72 hours to cancel
3. After timelock, manager can:
   - Execute native coin recovery
   - Execute token recovery
   - Complete the recovery process

## Security Considerations

- Keep private keys secure
- Verify addresses before deployment
- Monitor events for unauthorized attempts
- Test thoroughly on testnet before mainnet deployment
- Use secure RPC endpoints

## Development

```shell
# Start local Hardhat node
npx hardhat node

# Run tests
npm test

# Deploy to local network
npx hardhat run scripts/deploy.js --network localhost
```

## Contributing

1. Fork the repository
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create a new Pull Request
