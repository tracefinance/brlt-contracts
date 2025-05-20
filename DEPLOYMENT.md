# BRLT Token Deployment Guide

This guide provides instructions for deploying the BRLT ERC-20 token to supported networks.

## 1. Prerequisites

Before deploying, ensure you have the following installed and configured:

*   **Node.js**: Version 18.x or higher.
*   **npm** or **yarn**: For managing project dependencies.
*   **Git**: For cloning the repository.
*   A configured **`.env` file**: See section below.

## 2. `.env` File Configuration

Create a `.env` file in the project root by copying from `.env.example` (if available) or creating a new one. This file stores sensitive information and network configurations.

```env
# Wallet private key for deployment (DO NOT COMMIT THIS FILE)
PRIVATE_KEY="your_wallet_private_key"

# (Optional) Initial Admin for BRLT contract.
# If not set, the deployer address will be used as the initial admin.
BRLT_INITIAL_ADMIN="0xYourInitialAdminAddress"

# RPC URLs for target networks
BASE_RPC_URL="https://mainnet.base.org"
BASE_TESTNET_RPC_URL="https://goerli.base.org"
POLYGON_ZKEVM_RPC_URL="https://zkevm-rpc.com"
POLYGON_ZKEVM_TESTNET_RPC_URL="https://rpc.public.zkevm-test.net"

# Block Explorer API Keys for contract verification
BASESCAN_API_KEY="your_basescan_api_key"
POLYGONSCAN_API_KEY="your_polygonscan_api_key_for_zkevm"
```

**Note:**
*   Replace placeholder values with your actual data.
*   The `POLYGONSCAN_API_KEY` should be for Polygon zkEVM networks.
*   If `BRLT_INITIAL_ADMIN` is not specified, the address derived from `PRIVATE_KEY` (the deployer) will be granted all administrative roles for the BRLT contract.

## 3. Install Dependencies

Navigate to the project root in your terminal and install the necessary dependencies:

```bash
npm install
# or
# yarn install
```

## 4. Compile Contracts

Compile the smart contracts:

```bash
npm run compile
```

This command compiles all contracts in the `solidity/` directory and places artifacts in the `artifacts/` directory.

## 5. Deploy BRLT Token

The `scripts/deploy.js` script is used to deploy contracts. To deploy the BRLT token, use the `--contract BRLT` flag and specify the target network.

### Target Networks:

*   **Base Mainnet**: `base`
*   **Base Goerli Testnet**: `base-testnet`
*   **Polygon zkEVM Mainnet**: `polygon-zkevm`
*   **Polygon zkEVM Testnet**: `polygon-zkevm-testnet`
*   **Local Hardhat Network**: `hardhat` (for quick testing, no verification)

### Deployment Commands:

Replace `<network_name>` with one of the target network names from the list above.

**Example for Base Goerli Testnet:**

```bash
npm run deploy:base-test -- --contract BRLT
```
This command uses the pre-defined script `deploy:base-test` from `package.json` which is `hardhat run scripts/deploy.js --network base-testnet`. The extra `-- --contract BRLT` passes the argument to `scripts/deploy.js`.

**General Deployment Command Structure:**

```bash
npx hardhat run scripts/deploy.js --network <network_name> --contract BRLT
```

**Specific Network Deployment Commands:**

*   **Base Mainnet:**
    ```bash
    npx hardhat run scripts/deploy.js --network base --contract BRLT
    # Or using package.json script:
    # npm run deploy:base -- --contract BRLT
    ```

*   **Base Goerli Testnet:**
    ```bash
    npx hardhat run scripts/deploy.js --network base-testnet --contract BRLT
    # Or using package.json script:
    # npm run deploy:base-test -- --contract BRLT
    ```

*   **Polygon zkEVM Mainnet:**
    ```bash
    npx hardhat run scripts/deploy.js --network polygon-zkevm --contract BRLT
    # Or using package.json script:
    # npm run deploy:polygon -- --contract BRLT
    ```

*   **Polygon zkEVM Testnet:**
    ```bash
    npx hardhat run scripts/deploy.js --network polygon-zkevm-testnet --contract BRLT
    # Or using package.json script:
    # npm run deploy:polygon-test -- --contract BRLT
    ```

The script will output the deployed BRLT contract address.

## 6. Contract Verification

Contract verification on block explorers (Basescan, Polygonscan for zkEVM) is attempted automatically by the deployment script for live networks.

*   Ensure your `BASESCAN_API_KEY` and `POLYGONSCAN_API_KEY` are correctly set in the `.env` file.
*   The script will log the verification status.
*   If automatic verification fails, you can manually verify using Hardhat's verification plugin or the block explorer's UI, providing the contract address, source code (`solidity/BRLT.sol`), and constructor arguments (the `initialAdmin` address used).

## 7. Post-Deployment Steps

After deployment and verification:

1.  **Record Contract Address**: Note the deployed BRLT token address for the respective network. This address will be used for interactions and integration.
2.  **Role Management (if needed)**:
    *   The `initialAdmin` (either specified in `.env` or the deployer) receives `DEFAULT_ADMIN_ROLE`, `MINTER_ROLE`, `BURNER_ROLE`, `PAUSER_ROLE`, and `BLACKLISTER_ROLE`.
    *   If you need to assign these roles to other addresses (e.g., a separate operations wallet for minting, a Gnosis Safe for admin), the `initialAdmin` account will need to call `grantRole()` on the deployed BRLT contract for each role and target address.
    *   Consider renouncing roles from the `initialAdmin` if it's an EOA and roles are transferred to more secure multisigs.
3.  **Test Functionality**: Interact with the deployed contract on the testnet to confirm its core functionalities (minting, transferring, pausing, blacklisting if applicable) are working as expected.

This completes the deployment process for the BRLT token. 

## 8. Upgrading the BRLT Contract (UUPS)

The BRLT contract is designed using the UUPS (Universal Upgradeable Proxy Standard) pattern, allowing its logic to be upgraded without changing the contract address or losing state (like balances and roles).

The upgrade process is controlled by an address holding the `UPGRADER_ROLE`.

### Upgrade Steps:

1.  **Develop the New Implementation (`BRLTvN.sol`):**
    *   Create a new version of the BRLT contract (e.g., `solidity/BRLTv3.sol`).
    *   This new contract must inherit from the previous version (or its core components like `BRLTStorage`, if you separate storage) and `UUPSUpgradeable`.
    *   It must include an `_authorizeUpgrade(address newImplementation)` function, typically restricting it to the `UPGRADER_ROLE`.
    *   It can add new functions, modify existing ones, or add new state variables (respecting storage layout rules for upgrades).
    *   If new state variables are added, ensure they are initialized correctly, potentially using a `reinitializer` function (e.g., `initializeV3() public reinitializer(N)` where `N` is the new version number for reinitialization).
    *   **Important:** Thoroughly test the new implementation locally, including interaction with the proxy.

2.  **Create an Upgrade Script:**
    *   While upgrades can be done manually with Hardhat tasks, it's recommended to use a script for repeatability and clarity.
    *   Create a script in the `scripts/` directory (e.g., `scripts/upgradeBRLT.js`).
    *   This script will use `hre.upgrades.upgradeProxy()`.

    **Example `scripts/upgradeBRLT.js`:**
    ```javascript
    const hre = require("hardhat");

    async function main() {
      const { upgrades } = hre;
      const proxyAddress = "0xYourDeployedBRLTProxyAddress"; // Replace with your BRLT proxy address
      const NewBRLTContractFactory = await hre.ethers.getContractFactory("BRLTv3"); // Replace BRLTv3 with your new contract name

      console.log(`Upgrading BRLT proxy at ${proxyAddress} to new implementation...`);
      const upgradedContract = await upgrades.upgradeProxy(proxyAddress, NewBRLTContractFactory, {
        kind: 'uups',
        // call: { fn: 'initializeV3', args: [/* new arguments if any */] } // Optional: if your new version has a reinitializer
      });

      await upgradedContract.waitForDeployment();
      const newImplementationAddress = await upgrades.erc1967.getImplementationAddress(await upgradedContract.getAddress());

      console.log("BRLT proxy upgraded successfully.");
      console.log(`Proxy address: ${await upgradedContract.getAddress()}`);
      console.log(`New implementation address: ${newImplementationAddress}`);

      // Wait for block confirmations before verification
      if (hre.network.name !== "hardhat" && hre.network.name !== "localhost") {
        console.log("Waiting for 5 block confirmations...");
        // Access the deployment transaction from the upgraded contract instance if available,
        // or use a generic wait if not directly accessible post-upgrade via the `upgradedContract` object directly.
        // This part might need adjustment based on how `upgradeProxy` returns and what it makes available.
        // A simple way is to fetch the transaction that changed the implementation slot.
        // For now, a simple delay or manual check might be needed before verification.
        // await upgradedContract.deploymentTransaction().wait(5); // This might not be correct for upgrades

        // A more reliable way to wait for confirmations for the upgrade transaction itself
        // would involve capturing the transaction hash from the upgradeProxy call if possible,
        // or using a delay. Let's assume a delay for simplicity in this example:
        await new Promise(resolve => setTimeout(resolve, 30000)); // Wait 30 seconds

        console.log("Verifying new implementation contract on explorer...");
        try {
          await hre.run("verify:verify", {
            address: newImplementationAddress,
            // No constructor arguments for the implementation if it uses an initializer pattern for its own setup.
            // If it had a constructor that was run upon deployment (rare for UUPS implementations meant for proxies), list them.
          });
          console.log("New implementation contract verified successfully.");
        } catch (error) {
          console.error("Verification of new implementation failed:", error.message);
        }
      }
    }

    main()
      .then(() => process.exit(0))
      .catch((error) => {
        console.error(error);
        process.exit(1);
      });
    ```

3.  **Configure `.env` for the Upgrader:**
    *   Ensure the `PRIVATE_KEY` in your `.env` file corresponds to an account that holds the `UPGRADER_ROLE` on the BRLT proxy contract for the target network.

4.  **Run the Upgrade Script:**
    *   Execute the script for the target network.
        ```bash
        npx hardhat run scripts/upgradeBRLT.js --network <network_name>
        ```
    *   Replace `<network_name>` with the desired network (e.g., `base-testnet`).
    *   The script will deploy the new implementation, call `upgradeTo` on the proxy, and optionally call any reinitializer function.

5.  **Verify New Implementation (if not done by script):**
    *   The example script attempts to verify the new implementation contract.
    *   If verification fails or is not part of your script, you may need to manually verify the new implementation contract address on the block explorer. Provide its source code (e.g., `solidity/BRLTv3.sol`). Constructor arguments are usually not needed for UUPS implementation contracts if they rely solely on initializers.

6.  **Test Upgraded Contract:**
    *   Thoroughly test all functionalities on the upgraded contract on a testnet, paying close attention to:
        *   Preservation of existing state (balances, roles, allowances).
        *   Correct functioning of new features.
        *   Correct functioning of modified features.
        *   No regressions in existing, unchanged functionality.

### Security Considerations for Upgrades:

*   **UPGRADER_ROLE Security**: The account(s) holding the `UPGRADER_ROLE` must be highly secure (e.g., a Gnosis Safe multi-sig wallet). Compromise of this role means an attacker can change the contract logic.
*   **Timelocks**: For critical upgrades, consider using a Timelock contract. The `UPGRADER_ROLE` would be transferred to the Timelock. Upgrades would then require a two-step process: propose an upgrade, wait for a delay period, then execute. This gives users time to review upcoming changes.
*   **Audit New Code**: Always have new contract versions audited before deploying to mainnet.
*   **Communication**: Clearly communicate upcoming upgrades to your users/community, especially if they involve significant changes or require user action.

This section outlines the process for upgrading the BRLT token. Always proceed with caution, especially on mainnet. 