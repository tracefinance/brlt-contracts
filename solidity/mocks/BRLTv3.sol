// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "./BRLTv2.sol"; // Inherits from BRLTv2, which inherits from BRLT's core logic

/**
 * @title BRLTv3 Mock Upgradeable Contract
 * @author Vault0 Team
 * @notice Mock contract for testing UUPS upgradeability from BRLTv2.
 * Adds a new role, a new state variable, and a new view function.
 * Changes the symbol to BRLTV3.
 * @dev This contract is intended for testing purposes only.
 */
contract BRLTv3 is BRLTv2 {
    /// @dev Role identifier for addresses authorized to report data.
    bytes32 public constant REPORTER_ROLE = keccak256("REPORTER_ROLE");

    /// @dev Stores the block number of the last report.
    uint256 public lastReportedBlock;

    /// @dev Emitted when V3 is initialized.
    event V3Initialized(address indexed initializer);
    /// @dev Emitted when a report is made.
    event ReportMade(address indexed reporter, uint256 blockNumber);

    /**
     * @notice Initializes BRLTv3 if deployed as the first implementation.
     * @dev Calls the initializers of parent contracts.
     */
    function initialize(address initialAdmin) public virtual override initializer {
        super.initialize(initialAdmin); // Call BRLTv2's initialize function (which calls BRLT's)
        // BRLTv3 specific standalone initialization (if any) would go here.
        // initializeV3 is intended for post-upgrade V3 state setup.
    }

    /**
     * @notice Initializes V3 specific features.
     * @dev This function can be called after upgrading to V3.
     * Typically, an initializer should be protected (e.g., `initializer` modifier or access control),
     * but for this mock, it's left open for simplicity in testing direct calls.
     * Grants REPORTER_ROLE to the caller for convenience in testing.
     */
    function initializeV3() external virtual {
        _grantRole(REPORTER_ROLE, _msgSender()); // Grant role to initializer for easy testing
        emit V3Initialized(_msgSender());
    }

    /**
     * @notice Allows an account with REPORTER_ROLE to set the last reported block.
     * @param blockNum The block number to set.
     */
    function makeReport(uint256 blockNum) external virtual onlyRole(REPORTER_ROLE) {
        lastReportedBlock = blockNum;
        emit ReportMade(_msgSender(), blockNum);
    }

    /**
     * @notice Returns the version 2 field (from BRLTv2) and the last reported block.
     * @return v2Field The value of version2Field.
     * @return reportedBlock The value of lastReportedBlock.
     */
    function getReportData() external view virtual returns (uint256 v2Field, uint256 reportedBlock) {
        v2Field = version2Field; // Accesses state from BRLTv2
        reportedBlock = lastReportedBlock;
    }
    
    /**
     * @notice Returns the token symbol "BRLTV3".
     * @return The token symbol.
     */
    function symbol() public pure virtual override returns (string memory) {
        return "BRLTV3";
    }

    // _authorizeUpgrade is inherited from BRLT (via BRLTv2) and allows UPGRADER_ROLE to upgrade.
    // No need to override if the logic remains the same.
} 