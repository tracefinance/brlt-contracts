# BRLT Token Implementation and Project Refinement

This document outlines the tasks for analyzing the current project, implementing the BRLT ERC-20 token, and setting up multi-chain deployment.

## Completed Tasks

*(No tasks completed yet)*

## In Progress Tasks

**Phase 1: Current Project Review & BRLT Specification**
1. [x] **Analyze Current Project:** Review existing project setup (Hardhat config, `package.json` scripts, existing `MultiSigWallet` tests, and deployment setup) to understand the current environment.
   - Hardhat config: Solidity 0.8.28, multi-chain (Base, Polygon zkEVM mainnets/testnets), Etherscan verification.
   - package.json: Scripts for test, compile, deploy (all use `scripts/deploy.js`). OZ 5.2.0 installed.
   - Tests: `MultiSigWallet.js`, `MultiSigWalletExtended.js` exist.
   - Deployment: `scripts/deploy.js` and `ignition/modules/MultiSigWallet.js` exist.
2. [x] **Verify Existing Stability:** Run all tests for the existing `MultiSigWallet` to ensure the current codebase is stable and all tests pass.
   - `npm test` ran successfully, 62 tests passed.
3. [x] **Identify Integration Needs:** Document any potential improvements or refactoring required in the current project to seamlessly integrate the new BRLT token (e.g., shared libraries, testing utilities, deployment script enhancements).
   - Deployment script (`scripts/deploy.js`) refactored to accept `--contract <ContractName>` argument.
   - Current deployment logic for `MultiSigWallet` preserved.
   - Placeholder for `BRLT` deployment added.
   - Testing utilities: Standard Hardhat/Chai. Consider if new tests should be JS or TS.
4. [x] **Define BRLT Token Specifications:**
    - [x] Token Name: BRLT
    - [x] Token Symbol: BRLT
    - [x] Decimals: 6
    - [x] Initial Supply strategy: 0 at deployment; mintable by `MINTER_ROLE` as BRL is collateralized.
    - [x] Access Control model: OpenZeppelin `AccessControl` with:
        - `MINTER_ROLE` for minting.
        - `BURNER_ROLE` for burning (e.g., custom `burnFrom` controlled by this role).
        - `PAUSER_ROLE` for pausing/unpausing token transfers.
        - `DEFAULT_ADMIN_ROLE` to manage roles.
    - [x] Core Features: Standard ERC20, Mintable, Burnable (custom), Pausable.

## Future Tasks

**Phase 2: BRLT ERC-20 Contract Implementation**
5. [x] **Setup BRLT Contract:** Create `BRLT.sol` within the `solidity/` directory.
   - File created with initial structure, roles, and basic functions.
6. [x] **Import OpenZeppelin:** Add necessary OpenZeppelin contracts (e.g., `ERC20.sol`, `AccessControlEnumerable.sol`, `Pausable.sol`).
   - Imports included in `BRLT.sol`.
7. [x] **Implement Constructor:** Write the constructor for `BRLT.sol` to set the name, symbol, decimals, and handle initial supply/owner based on defined specifications.
   - Constructor implemented, sets name/symbol, grants roles to initial admin.
   - `decimals()` function overridden to 18.
   - Basic `mint`, `burnFrom`, `pause`, `unpause` functions added with role checks.
8. [x] **Implement Custom Features (if any):** Add any specific logic required for BRLT beyond standard OpenZeppelin ERC20 functionality.
   - Integrated Blacklisting capability with `BLACKLISTER_ROLE`.
   - Integrated EIP-2612 Permit functionality (`ERC20Permit`).
   - Updated `_beforeTokenTransfer` to handle Pausable and Blacklist checks.
9. [x] **Add NatSpec Documentation:** Write comprehensive NatSpec comments for `BRLT.sol` covering all functions, state variables, events, and modifiers.
   - NatSpec documentation added for contract, roles, events, constructor, and all functions.

**Phase 3: BRLT Token Testing**
10. [x] **Create Test File:** Set up `BRLT.test.js` (or `.ts`) in the `test/` directory.
    - `test/BRLT.test.js` created with initial structure, deployment fixture, and basic deployment tests.
11. [x] **Write Core ERC20 Tests:**
    - Test deployment: verify name, symbol, decimals, initial supply, owner/admin roles.
    - Test standard functions: `balanceOf`, `transfer`, `transferFrom`, `approve`, `allowance`.
    - Test event emissions for transfers and approvals.
12. [x] **Write Feature-Specific Tests:**
    - [x] Test minting/burning functions (if applicable).
    - [x] Test pausable functions (if applicable).
    - [x] Test access control for administrative functions.
    - [x] Test any other custom logic.
13. [x] **Test Edge Cases:** Include tests for zero values, insufficient balances, transferring to zero address (if not blocked by OZ), etc.
14. [x] **Achieve High Test Coverage:** Ensure tests cover at least 90-95% of `BRLT.sol`.

**Phase 4: Multi-Chain Deployment Setup**
15. [x] **Review/Update Hardhat Config:** Ensure `hardhat.config.js` includes network configurations (RPC URLs, chain IDs, private keys/mnemonics via `.env`) for all target mainnets and testnets.
16. [x] **Develop BRLT Deployment Script:**
    - Create a new script in `scripts/` (e.g., `deployBRLT.js`) or a new Ignition module in `ignition/modules/` (e.g., `BRLTModule.js`).
    - The script should handle constructor arguments for BRLT.
    - The script must be configurable for different networks.
17. [x] **Implement Contract Verification:** Integrate automated Etherscan (and other explorers like Basescan, Polygonscan) verification into the deployment script for BRLT.
18. [x] **Test Deployment on Testnets:** Perform trial deployments of BRLT to all specified testnets, verifying contract functionality and explorer verification. *(User action required)*
19. [x] **Document Deployment Process:** Create clear documentation for deploying BRLT to each mainnet and testnet, including prerequisites, steps, and verification checks.
20. [ ] **Prepare for Mainnet Deployment:** Conduct final reviews and checklists before any mainnet deployment. *(User action/review required)*
    - [ ] **Define and Implement `initialAdmin` Strategy:** Decide on the strategy for the `initialAdmin` address (e.g., Gnosis Safe multi-sig). Set up the chosen address/mechanism before mainnet deployment. (Strongly recommended: Gnosis Safe).
    - [ ] **Conduct Final Code & Security Review:** Perform a thorough internal review of `BRLT.sol` and consider a formal third-party audit.
    - [ ] **Verify Test Coverage & Scenarios:** Ensure all critical paths and edge cases are covered by tests.
    - [ ] **Finalize Deployment Scripts & Configuration:** Double-check mainnet configurations, private key management, and gas strategies.
    - [ ] **Update All Documentation:** Ensure `README.md`, `DEPLOYMENT.MD`, and other relevant docs are accurate for mainnet.
    - [ ] **Establish Operational Procedures:** Plan for monitoring, emergency responses, and ongoing role management.
    - [ ] **Obtain Stakeholder Go/No-Go:** Secure final approval before proceeding.

**Phase 5: Advanced Features & Maintenance (New Phase)**
21. [ ] **Implement UUPS Upgradability for BRLT:**
    - [x] Decision: Implement UUPS (Universal Upgradeable Proxy Standard).
    - [x] Refactor `BRLT.sol` to use OpenZeppelin Upgradeable contracts (e.g., `UUPSUpgradeable`, `ERC20Upgradeable`, `AccessControlUpgradeable`, `PausableUpgradeable`, `ERC20PermitUpgradeable`).
    - [x] Implement `initialize` function in `BRLT.sol` to replace constructor logic.
    - [x] Implement `_authorizeUpgrade(address newImplementation)` function in `BRLT.sol`, restricting it to an `UPGRADER_ROLE`.
    - [x] Add `UPGRADER_ROLE` constant and grant it to the initial admin in the `initialize` function.
    - [x] Update `BRLT.test.js` to use the new initializer pattern for deployment fixtures.
    - [x] Add tests to `BRLT.test.js` for the upgrade mechanism (e.g., deploying a V2, upgrading, and checking state/functionality).
    - [x] Create/Update deployment scripts to handle deploying the UUPS proxy (`ERC1967Proxy`) and the `BRLT` implementation contract.
    - [x] Document the BRLT upgrade process in `DEPLOYMENT.md`.

**Phase 6: Integration Testing & Local Network Interaction (New Phase)**
22. [x] **Setup Local Hardhat Node for Interaction:**
    - [x] Confirm `localhost` network configuration in `hardhat.config.js`.
    - [x] Document procedure for starting a standalone `npx hardhat node`.
    - [x] Document deploying `BRLT` to the local node using `deploy-contract --network localhost`.
23. [x] **Develop Interaction Scripts (Optional but Recommended):**
    - [x] Create an initial interaction script (`scripts/interactions/queryBRLT.js`) to query basic contract info, balance, and roles.
    - [x] Create example scripts in `scripts/interactions/` to demonstrate common BRLT operations (mint, transfer, approve, check balance, get roles, etc.) against a local or testnet deployment.
24. [x] **Write Scenario-Based Integration Tests:**
    - [x] Add new test suites to `test/BRLT.test.js` or create a new `test/BRLT.integration.test.js`.
    - [x] Implement tests for full user lifecycle scenarios (mint -> transfer -> approve -> transferFrom -> permit -> burn).
    - [x] Implement tests for pause/unpause scenarios, verifying operational restrictions.
    - [x] Implement tests for blacklisting/unblacklisting scenarios, verifying transfer restrictions.
    - [x] Implement tests for comprehensive access control management (granting/revoking multiple roles by different admins).
    - [x] Implement tests for more complex upgrade scenarios, ensuring state preservation and functionality of new/modified features after an upgrade.
25. [x] **Document Local Interaction and Testing:**
    - [x] Update `README.md` or a development guide on how to use the Hardhat console and interaction scripts for local testing.

## Implementation Plan

*   **Initial Phase:** Focus on analyzing the existing project structure and defining the precise specifications for the BRLT token. This involves understanding any current limitations or necessary improvements to the Hardhat environment.
*   **Development Phase:** Implement the `BRLT.sol` contract leveraging OpenZeppelin's robust and audited components. Write thorough unit tests to ensure correctness and security.
*   **Deployment Phase:** Configure Hardhat for multi-chain deployments, develop and test deployment scripts (including verification), and document the process for testnets and mainnets.

## Relevant Files

*(This section will be updated as files are created/modified)*
- `solidity/BRLT.sol` - (To be created) The BRLT ERC-20 smart contract.
- `test/BRLT.test.js` - (To be created) Unit tests for the BRLT contract.
- `scripts/deployBRLT.js` or `ignition/modules/BRLTModule.js` - (To be created) Deployment script/module for BRLT.
- `hardhat.config.js` - (To be updated) For network configurations.
- `.env` - (To be updated) For sensitive deployment information like private keys and RPC URLs. 