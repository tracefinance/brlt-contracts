// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "../BRLT.sol"; // Import the upgradeable BRLT
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/**
 * @title BRLTv2 (for testing UUPS upgrades)
 * @notice A mock V2 implementation of BRLT.
 */
contract BRLTv2 is BRLT {
    uint256 public version2Field;

    /// @custom:oz-upgrades-unsafe-allow constructor
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initializer for BRLTv2. If BRLTv2 were deployed as the first implementation.
     * @dev Calls the initializers of parent contracts.
     */
    function initialize(address initialAdmin) public override initializer {
        super.initialize(initialAdmin); // Call BRLT's initialize function
        // Any BRLTv2 specific initialization logic (not for reinitialization from V1) would go here.
        // For this mock, there isn't any for the main initialize.
    }

    /**
     * @notice Re-initializer for BRLTv2 specific state. Only intended for use during upgrade from V1.
     * @dev Sets the version2Field.
     */
    function initializeV2(uint256 initialV2Value) public reinitializer(2) {
        version2Field = initialV2Value;
    }

    /**
     * @notice Returns the token symbol (overridden to "BRLTV2").
     * @return The token symbol.
     */
    function symbol() public pure override returns (string memory) {
        return "BRLTV2";
    }

    /**
     * @notice Gets the value of version2Field.
     * @return The value of version2Field.
     */
    function getVersion2Field() public view returns (uint256) {
        return version2Field;
    }

    /**
     * @dev This empty reserved space is put in place to allow future versions to add new
     * variables without shifting down storage in the inheritance chain.
     * See https://docs.openzeppelin.com/contracts/4.x/upgradeable#storage_gaps
     */
    uint256[50] private __gap;
} 