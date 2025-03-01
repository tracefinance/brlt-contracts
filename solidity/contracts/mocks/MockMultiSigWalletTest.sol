// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "../MultiSigWallet.sol";

/**
 * @title MockMultiSigWalletTest
 * @dev Extension of MultiSigWallet for testing internal functions
 */
contract MockMultiSigWalletTest is MultiSigWallet {
    constructor(address _client, address _recoveryAddress) 
        MultiSigWallet(_client, _recoveryAddress) {}
    
    // Function to sign without executing, to test direct execution later
    function signWithdrawalWithoutExecution(bytes32 requestId) external onlyAuthorized {
        WithdrawalRequest storage request = withdrawalRequests[requestId];
        require(request.timestamp > 0, "Request not found");
        require(!request.executed, "Already executed");
        require(request.timestamp + WITHDRAWAL_EXPIRATION >= block.timestamp, "Request expired");
        
        if (msg.sender == manager) {
            require(!request.managerSigned, "Already signed");
            request.managerSigned = true;
        } else {
            require(!request.clientSigned, "Already signed");
            request.clientSigned = true;
        }
        
        emit WithdrawalSigned(requestId, msg.sender);
        // Note: We don't call _executeWithdrawal here unlike the original
    }
    
    // Function to directly execute withdrawal, bypassing normal flow
    function executeWithdrawalDirect(bytes32 requestId) external {
        _executeWithdrawal(requestId);
    }
}
