// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "../MultiSigWallet.sol";

/**
 * @title MockMultiSigWalletTest
 * @dev Extension of MultiSigWallet for testing internal functions
 */
contract MockMultiSigWalletTest is MultiSigWallet {
    constructor(
        address[] memory _signers, 
        uint256 _quorum, 
        address _recoveryAddress, 
        address[] memory _whitelistedTokens
    ) 
        MultiSigWallet(_signers, _quorum, _recoveryAddress, _whitelistedTokens) {}
    
    // Function to sign without executing, to test direct execution later
    function signWithdrawalWithoutExecution(bytes32 requestId) external onlySigner {
        WithdrawalRequest storage request = withdrawalRequests[requestId];
        require(request.timestamp > 0, "Request not found");
        require(!request.executed, "Already executed");
        require(request.timestamp + WITHDRAWAL_EXPIRATION >= block.timestamp, "Request expired");
        require(!withdrawalSignatures[requestId][msg.sender], "Already signed");
        
        withdrawalSignatures[requestId][msg.sender] = true;
        request.signatureCount++;
        
        emit WithdrawalSigned(requestId, msg.sender);
        // Note: We don't execute the withdrawal here, unlike the original signWithdrawal function
    }
    
    // Function to directly execute withdrawal, bypassing normal flow
    function executeWithdrawalDirect(bytes32 requestId) external {
        _executeWithdrawal(requestId);
    }
}
