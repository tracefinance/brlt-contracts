# BRLT Token Project

[![Smart Contract CI](https://github.com/tracefinance/fx-multisig-wallet/actions/workflows/compile-and-test.yml/badge.svg)](https://github.com/tracefinance/fx-multisig-wallet/actions/workflows/compile-and-test.yml)
[![codecov](https://codecov.io/gh/tracefinance/fx-multisig-wallet/graph/badge.svg?token=fsCmAuBO0b)](https://codecov.io/gh/tracefinance/fx-multisig-wallet)

An ERC20 stablecoin pegged to BRL (Brazilian Real), implemented as a UUPS upgradeable smart contract. Features include minting, burning, pausing, blacklisting capabilities, and EIP-2612 permit functionality.

## Features

-   **ERC20 Standard:** Fully compliant with the ERC20 token standard.
-   **BRL Pegged:** Designed to maintain a 1:1 peg with the Brazilian Real.
-   **UUPS Upgradeable:** Contract logic can be upgraded without changing the token address.
-   **Access Control:** Granular role-based permissions for administrative functions (minter, burner, pauser, blacklister, upgrader) using OpenZeppelin AccessControl.
-   **Mintable & Burnable:** Tokens can be minted by authorized minters and burned by authorized burners.
-   **Pausable:** Token transfers and other operations can be paused by authorized pausers.
-   **Blacklisting:** Addresses can be blacklisted, preventing them from sending or receiving tokens.
-   **EIP-2612 Permit:** Supports gasless approvals via signatures.

## Project Structure

```
├── solidity/
│   ├── BRLT.sol             # Main BRLT token contract (UUPS Upgradeable)
│   └── mocks/
│       ├── BRLTv2.sol       # Mock for testing BRLT upgrades
│       └── MockToken.sol    # Generic test ERC-20 token
├── test/
│   └── BRLT.test.js         # Test suite for BRLT token
├── scripts/
│   └── deploy-task.js     # Hardhat task for deploying BRLT
├── ignition/
│   └── modules/             # Ignition deployment modules (currently unused for BRLT)
├── tasks/
│   └── BRLT_Token_Development.md # Development task tracking
├── .openzeppelin/             # Hardhat Upgrades plugin manifest files
├── DEPLOYMENT.MD              # Detailed deployment and upgrade instructions
└── hardhat.config.js          # Hardhat configuration
```

## Prerequisites

-   Node.js >= 14.0.0
-   npm >= 6.0.0

## Installation

```shell
# Install dependencies
npm install
```

## Configuration

1.  Copy `.env.example` to `.env` and update with your specific values:
    ```shell
    cp .env.example .env
    ```
2.  Ensure your `.env` file includes:
    ```env
    # Network RPC URLs (replace with your actual provider URLs)
    BASE_RPC_URL=https://mainnet.base.org
    BASE_TESTNET_RPC_URL=https://sepolia.base.org
    POLYGON_ZKEVM_RPC_URL=https://zkevm-rpc.com
    POLYGON_AMOY_RPC_URL=https://rpc-amoy.polygon.technology/
    SEPOLIA_RPC_URL=https://rpc.sepolia.org # For Ethereum Sepolia

    # API Keys for contract verification
    ETHERSCAN_API_KEY=your_etherscan_api_key # Used for Ethereum Mainnet & Sepolia
    BASESCAN_API_KEY=your_basescan_api_key
    POLYGONSCAN_API_KEY=your_polygonscan_api_key

    # Private Key for deployment (use with caution, preferably from a secure vault or hardware wallet for mainnet)
    PRIVATE_KEY=your_deployer_private_key_here

    # Optional: Initial Admin for BRLT contract roles (defaults to deployer if not set)
    BRLT_INITIAL_ADMIN=your_initial_admin_address_here
    ```

## Testing

```shell
# Run all tests
npm test

# Run tests with coverage report
npm run test:coverage

# Run tests with gas reporting (ensure REPORT_GAS=true in .env or hardhat.config.js if needed by tests)
npm run test:gas

# Run tests in watch mode (development)
npm run test:watch
```

## Deployment

The BRLT contract can be deployed to various networks using the `deploy-contract` Hardhat task. The `--contract-name BRLT` parameter is required.

### Ethereum Sepolia Testnet
```shell
npm run deploy:sepolia
```

### Base Testnet (Base Sepolia)
```shell
npm run deploy:base-test
```

### Base Mainnet
```shell
npm run deploy:base
```

### Polygon Amoy Testnet
```shell
npm run deploy:polygon-amoy
```

### Polygon zkEVM Mainnet
```shell
npm run deploy:polygon
```

### Deployment Verification
The deployment task automatically attempts to:
1. Deploy the BRLT contract as a UUPS proxy.
2. Wait for block confirmations.
3. Verify the implementation contract and link the proxy on the respective block explorer.

Refer to `DEPLOYMENT.MD` for more detailed deployment and upgrade procedures.

## Network Configurations

The project supports multiple networks. Key details:

### Ethereum
-   **Sepolia (Testnet)**
    -   Chain ID: `11155111`
    -   Explorer: [https://sepolia.etherscan.io](https://sepolia.etherscan.io)

### Base
-   **Mainnet**
    -   Chain ID: `8453`
    -   Explorer: [https://basescan.org](https://basescan.org)
-   **Sepolia (Testnet)**
    -   Chain ID: `84532`
    -   Explorer: [https://sepolia.basescan.org](https://sepolia.basescan.org)


### Polygon
-   **zkEVM Mainnet**
    -   Chain ID: `1101`
    -   Explorer: [https://zkevm.polygonscan.com](https://zkevm.polygonscan.com)
-   **Amoy (Testnet)**
    -   Chain ID: `80002`
    -   Explorer: [https://www.oklink.com/amoy](https://www.oklink.com/amoy)


## Development

For local development and testing, you can use the Hardhat Network, a local Ethereum network designed for development.

### Running a Local Development Node

1.  **Start a local Hardhat node:**
    ```bash
    npx hardhat node
    ```
    This will start a local Ethereum node with several pre-funded accounts, useful for testing and development.

2.  **Deploy the `BRLT` contract to the local node:**
    In a new terminal, run:
    ```bash
    npm run deploy:localhost -- --contract-name BRLT
    ```
    Take note of the deployed proxy and implementation addresses.

### Interacting with the Contract Locally

Once deployed locally, you can interact with the `BRLT` token using the Hardhat console or custom scripts.

**Using Hardhat Console:**

1.  Connect to your local node:
    ```bash
    npx hardhat console --network localhost
    ```
2.  Inside the console, get an instance of your deployed BRLT contract. You'll need the BRLT contract's ABI (from `artifacts/solidity/BRLT.sol/BRLT.json`) and its deployed proxy address (from the deployment step):
    ```javascript
    // Example: Get the BRLT contract factory
    const BRLT = await ethers.getContractFactory("BRLT");
    // Attach to the deployed proxy address (replace with your actual proxy address)
    const brltProxyAddress = "0xYourBRLTProxyAddressHere"; // Get this from deployment output
    const brlt = BRLT.attach(brltProxyAddress);
    ```
3.  Now you can interact with the contract. Examples:
    ```javascript
    // Get token name
    await brlt.name();
    // Get total supply
    await brlt.totalSupply();
    // Get balance of an account (e.g., the first Hardhat node account)
    const signers = await ethers.getSigners();
    await brlt.balanceOf(signers[0].address);

    // If signers[0] has MINTER_ROLE, you can mint:
    // await brlt.connect(signers[0]).mint(signers[1].address, ethers.parseUnits("1000", 18));
    // await brlt.balanceOf(signers[1].address);

    // If signers[0] has PAUSER_ROLE, you can pause/unpause:
    // await brlt.connect(signers[0]).pause();
    // await brlt.paused();
    // await brlt.connect(signers[0]).unpause();
    ```
    Remember that for operations requiring specific roles (mint, pause, etc.), the signer used (`signers[0]` in these examples, which is the default if no `.connect()` is used for write operations) must possess that role. Roles are granted during the `initialize` function to the `initialAdmin` (which is the deployer by default in local deployments).

**Using Interaction Scripts:**

The project includes scripts for more complex interactions in the `scripts/interactions/` directory.

*   **`queryBRLT.js`**: This script queries basic information about the BRLT contract, such as its name, symbol, total supply, and the roles of pre-defined accounts. 
    To run it (after deploying to localhost):
    ```bash
    npx hardhat run scripts/interactions/queryBRLT.js --network localhost
    ```
    You might need to update the `brltProxyAddress` within the script if it's hardcoded and doesn't dynamically fetch it.

*   **Custom Scripts**: You can develop further scripts to automate sequences of interactions, such as minting to multiple users, performing transfers, checking blacklist status, etc., for testing or administrative tasks.

## Contributing

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feat/your-amazing-feature`).
3.  Commit your changes (`git commit -m 'feat: add some amazing feature'`).
4.  Push to the branch (`git push origin feat/your-amazing-feature`).
5.  Create a new Pull Request.
