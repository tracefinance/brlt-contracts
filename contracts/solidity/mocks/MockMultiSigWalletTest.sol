// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "../MultiSigWallet.sol";

/**
 * @title MockMultiSigWalletTest
 * @dev A mock contract that extends MultiSigWallet for testing purposes
 */
contract MockMultiSigWalletTest is MultiSigWallet {
    
    constructor(
        address[] memory _signers, 
        uint256 _quorum,
        address _recoveryAddress,
        address[] memory _whitelistedTokens
    ) MultiSigWallet(_signers, _quorum, _recoveryAddress, _whitelistedTokens) {}
    
    /**
     * @dev Allows testing of signWithdrawal without execution (even if quorum is reached)
     */
    function signWithdrawalWithoutExecution(bytes32 requestId) external onlySigner notInRecovery {
        WithdrawalRequest storage request = withdrawalRequests[requestId];
        require(request.timestamp > 0, "Request not found");
        require(!request.executed, "Already executed");
        require(request.timestamp + WITHDRAWAL_EXPIRATION >= block.timestamp, "Request expired");
        require(!withdrawalSignatures[requestId][msg.sender], "Already signed");
        
        withdrawalSignatures[requestId][msg.sender] = true;
        request.signatureCount++;
        
        emit WithdrawalSigned(requestId, msg.sender);
    }
    
    /**
     * @dev Exposes the internal _executeWithdrawal function for testing
     */
    function executeWithdrawalDirect(bytes32 requestId) external {
        _executeWithdrawal(requestId);
    }
}
