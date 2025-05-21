// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "../BRLT.sol"; // Import the upgradeable BRLT
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/**
 * @title BRLTv2 Mock Upgradeable Contract
 * @author Vault0 Team
 * @notice Mock contract for testing UUPS upgradeability from BRLT.
 * Adds a new state variable `version2Field` and an initializer `initializeV2`.
 * Overrides `symbol()` to return "BRLTV2".
 * @dev This contract is intended for testing purposes only.
 */
contract BRLTv2 is BRLT {
    uint256 public version2Field;

    /**
     * @dev Emitted when V2 is initialized.
     * @param initializer The address that called the V2 initializer.
     * @param versionField The value set for the version2Field.
     */
    event V2Initialized(address indexed initializer, uint256 versionField);

    /**
     * @notice Initializes BRLTv2 if deployed as the first implementation.
     * @dev Calls the initializers of parent contracts.
     */
    function initialize(address initialAdmin) public virtual override initializer {
        super.initialize(initialAdmin); // Call BRLT's initialize function
        // BRLTv2 specific standalone initialization (if any) would go here.
        // initializeV2 is intended for post-upgrade V2 state setup.
    }

    /**
     * @notice Initializes V2 specific features post-upgrade or for standalone V2.
     * @dev Sets the version2Field.
     * Uses reinitializer(2) to allow being called after the BRLT initialization.
     * @param initialV2Value The value to set for the version2Field.
     */
    function initializeV2(uint256 initialV2Value) public virtual reinitializer(2) {
        version2Field = initialV2Value;
        emit V2Initialized(_msgSender(), initialV2Value);
    }

    /**
     * @notice Returns the value of version2Field.
     * @return The value of version2Field.
     */
    function getVersion2Field() external view returns (uint256) {
        return version2Field;
    }
    
    /**
     * @notice Returns the token symbol "BRLTV2".
     * @return The token symbol.
     */
    function symbol() public pure virtual override returns (string memory) {
        return "BRLTV2";
    }

    // _authorizeUpgrade is inherited from BRLT and allows UPGRADER_ROLE to upgrade.
    // No need to override if the logic remains the same.
} 