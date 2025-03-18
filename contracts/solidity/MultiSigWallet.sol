// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";

/**
 * @title MultiSigWallet
 * @dev A wallet that requires multiple signatures (quorum) to withdraw funds
 * with a recovery mechanism that allows funds to be sent to a recovery address
 * after a 72-hour timelock period
 */
contract MultiSigWallet is ReentrancyGuard {
    using SafeERC20 for IERC20;

    // Signer management
    address[] public signers;
    mapping(address => bool) public isSigner;
    uint256 public immutable quorum;
    
    // Recovery configuration
    address public recoveryAddress;
    uint256 public constant RECOVERY_DELAY = 72 hours;
    uint256 public constant WITHDRAWAL_EXPIRATION = 24 hours;
    uint256 public constant MAX_SUPPORTED_TOKENS = 20;
    uint256 public recoveryRequestTimestamp;
    bool public recoveryExecuted;
    uint256 public withdrawalNonce;
    
    // Array to store supported tokens for recovery
    address[] public supportedTokensList;
    // Mapping to store supported tokens for automatic recovery
    mapping(address => bool) public supportedTokens;
    // Mapping to store whitelisted tokens that can be auto-included
    mapping(address => bool) public whitelistedTokens;
    
    // Recovery address change tracking
    struct RecoveryAddressProposal {
        address proposedAddress;
        uint256 timestamp;
        uint256 signatureCount;
    }
    
    mapping(bytes32 => RecoveryAddressProposal) public recoveryAddressProposals;
    mapping(bytes32 => mapping(address => bool)) public recoveryAddressProposalSignatures;
    
    // Withdrawal request structure
    struct WithdrawalRequest {
        address token;          // Address of token (address(0) for native coin)
        uint256 amount;         // Amount to withdraw
        address to;             // Recipient address
        uint256 timestamp;      // When the request was created
        bool executed;          // Whether the request was executed
        uint256 nonce;          // Unique nonce for the request
        uint256 signatureCount; // Number of signatures collected
    }
    
    mapping(bytes32 => WithdrawalRequest) public withdrawalRequests;
    mapping(bytes32 => mapping(address => bool)) public withdrawalSignatures;
    
    // Events
    event Deposited(address indexed token, address indexed from, uint256 amount);
    event WithdrawalRequested(bytes32 indexed requestId, address token, uint256 amount, address to, uint256 nonce);
    event WithdrawalSigned(bytes32 indexed requestId, address indexed signer);
    event WithdrawalExecuted(bytes32 indexed requestId, address token, uint256 amount, address to);
    event RecoveryRequested(uint256 timestamp);
    event RecoveryCancelled();
    event RecoveryExecuted(address indexed token, uint256 amount);
    event RecoveryCompleted();
    event TokenSupported(address indexed token);
    event TokenRemoved(address indexed token);
    event NonSupportedTokenRecovered(address indexed token, uint256 amount, address to);
    event TokenWhitelisted(address indexed token);
    event RecoveryAddressChangeProposed(address indexed proposer, address newRecoveryAddress, bytes32 indexed proposalId);
    event RecoveryAddressChangeSignatureAdded(address indexed signer, bytes32 indexed proposalId);
    event RecoveryAddressChanged(address indexed oldAddress, address indexed newAddress, bytes32 indexed proposalId);
    
    modifier onlySigner() {
        require(isSigner[msg.sender], "Only signer can call this function");
        _;
    }
    
    modifier notInRecovery() {
        require(recoveryRequestTimestamp == 0, "Wallet in recovery mode");
        _;
    }
    
    modifier recoveryCompleted() {
        require(recoveryExecuted, "Recovery not completed");
        _;
    }
    
    constructor(
        address[] memory _signers, 
        uint256 _quorum,
        address _recoveryAddress,
        address[] memory _whitelistedTokens
    ) {
        require(_signers.length >= 2 && _signers.length <= 7, "Must have 2-7 signers");
        require(_quorum >= (_signers.length + 1) / 2 && _quorum >= 2 && _quorum <= _signers.length, "Invalid quorum");
        require(_recoveryAddress != address(0), "Invalid recovery address");
        require(_whitelistedTokens.length < MAX_SUPPORTED_TOKENS, "Too many whitelisted tokens");
        
        // Set signers
        for (uint256 i = 0; i < _signers.length; i++) {
            address signer = _signers[i];
            require(signer != address(0), "Invalid signer address");
            require(!isSigner[signer], "Duplicate signer");
            isSigner[signer] = true;
            signers.push(signer);
        }
        
        quorum = _quorum;
        recoveryAddress = _recoveryAddress;
        
        // Accept native coin (ETH) by default
        supportedTokens[address(0)] = true;
        supportedTokensList.push(address(0));
        
        // Add whitelisted tokens
        for (uint256 i = 0; i < _whitelistedTokens.length; i++) {
            address token = _whitelistedTokens[i];
            require(token != address(0), "Cannot whitelist zero address");
            if (!whitelistedTokens[token]) {
                whitelistedTokens[token] = true;
                emit TokenWhitelisted(token);
            }
        }
    }
    
    /**
     * @dev Allows a signer to add a token to the supported tokens list for automatic recovery
     * @param token The token address to support
     */
    function addSupportedToken(address token) external onlySigner {
        require(token != address(0), "Cannot add zero address");
        require(!supportedTokens[token], "Token already supported");
        require(supportedTokensList.length < MAX_SUPPORTED_TOKENS, "Maximum supported tokens reached");
        supportedTokens[token] = true;
        supportedTokensList.push(token);
        emit TokenSupported(token);
    }
    
    /**
     * @dev Allows a signer to remove a token from the supported tokens list
     * @param token The token address to remove
     */
    function removeSupportedToken(address token) external onlySigner {
        require(supportedTokens[token], "Token not in supported list");
        supportedTokens[token] = false;
        
        // Remove token from supportedTokensList
        for (uint i = 0; i < supportedTokensList.length; i++) {
            if (supportedTokensList[i] == token) {
                supportedTokensList[i] = supportedTokensList[supportedTokensList.length - 1];
                supportedTokensList.pop();
                break;
            }
        }
        
        emit TokenRemoved(token);
    }
    
    /**
     * @dev Allows a signer to recover non-supported tokens after recovery is completed
     * @param token The token address to recover
     * @param to The address to send the recovered tokens to
     */
    function recoverNonSupportedToken(address token, address to) external onlySigner recoveryCompleted nonReentrant {
        require(token != address(0), "Cannot recover native coin");
        require(!supportedTokens[token], "Use regular recovery for supported tokens");
        require(to != address(0), "Invalid recipient address");
        
        uint256 balance = IERC20(token).balanceOf(address(this));
        require(balance > 0, "No balance to recover");
        
        IERC20(token).safeTransfer(to, balance);
        emit NonSupportedTokenRecovered(token, balance, to);
    }
    
    // Function to deposit ERC20 tokens
    function depositToken(address token, uint256 amount) external notInRecovery {
        require(token != address(0), "Use receive() for native coin");
        require(amount > 0, "Amount must be greater than 0");
        
        // Add token to supported tokens if it's whitelisted and not already supported
        if (whitelistedTokens[token] && !supportedTokens[token]) {
            require(supportedTokensList.length < MAX_SUPPORTED_TOKENS, "Maximum supported tokens reached");
            supportedTokens[token] = true;
            supportedTokensList.push(token);
            emit TokenSupported(token);
        }
        
        // Use SafeERC20 for the transfer
        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        emit Deposited(token, msg.sender, amount);
    }
    
    /**
     * @dev Returns the list of all supported tokens
     * @return Array of supported token addresses
     */
    function getSupportedTokens() external view returns (address[] memory) {
        return supportedTokensList;
    }
    
    /**
     * @dev Returns the list of all signers
     * @return Array of signer addresses
     */
    function getSigners() external view returns (address[] memory) {
        return signers;
    }
    
    // Recovery functions
    /**
     * @dev Any signer can request a recovery process
     */
    function requestRecovery() external onlySigner {
        require(recoveryRequestTimestamp == 0, "Recovery already requested");
        require(!recoveryExecuted, "Recovery already executed");
        
        recoveryRequestTimestamp = block.timestamp;
        emit RecoveryRequested(block.timestamp);
    }
    
    /**
     * @dev Any signer can cancel a recovery process before the delay period expires
     */
    function cancelRecovery() external onlySigner {
        require(recoveryRequestTimestamp > 0, "No recovery requested");
        require(!recoveryExecuted, "Recovery already executed");
        require(block.timestamp < recoveryRequestTimestamp + RECOVERY_DELAY, "Recovery period expired");
        
        recoveryRequestTimestamp = 0;
        recoveryExecuted = false;
        emit RecoveryCancelled();
    }
    
    /**
     * @dev Any signer can execute recovery after the delay period
     */
    function executeRecovery() external nonReentrant onlySigner {
        require(recoveryRequestTimestamp > 0, "No recovery requested");
        require(!recoveryExecuted, "Recovery already executed");
        require(block.timestamp >= recoveryRequestTimestamp + RECOVERY_DELAY, "Recovery delay not elapsed");
        
        // Transfer native coin if it's a supported token
        if (supportedTokens[address(0)]) {
            uint256 balance = address(this).balance;
            if (balance > 0) {
                (bool success, ) = recoveryAddress.call{value: balance}("");
                require(success, "Native coin transfer failed");
                emit RecoveryExecuted(address(0), balance);
            }
        }
        
        // Transfer all supported tokens' balances in one go
        for (uint i = 0; i < supportedTokensList.length; i++) {
            address token = supportedTokensList[i];
            if (token != address(0) && supportedTokens[token]) {
                uint256 balance = IERC20(token).balanceOf(address(this));
                if (balance > 0) {
                    IERC20(token).safeTransfer(recoveryAddress, balance);
                    emit RecoveryExecuted(token, balance);
                }
            }
        }
        
        // Complete the recovery process in the same transaction
        recoveryExecuted = true;
        recoveryRequestTimestamp = 0;
        emit RecoveryCompleted();
    }
    
    /**
     * @dev Propose a change to the recovery address
     * @param newRecoveryAddress The new address to use for recovery
     * @return proposalId The unique identifier for this proposal
     */
    function proposeRecoveryAddressChange(address newRecoveryAddress) external onlySigner notInRecovery returns (bytes32) {
        require(newRecoveryAddress != address(0), "Invalid recovery address");
        
        bytes32 proposalId = keccak256(abi.encode(
            "RECOVERY_ADDRESS_CHANGE", 
            newRecoveryAddress, 
            block.chainid, 
            address(this)
        ));
        
        // If this is a new proposal, initialize it
        if (recoveryAddressProposals[proposalId].timestamp == 0) {
            recoveryAddressProposals[proposalId] = RecoveryAddressProposal({
                proposedAddress: newRecoveryAddress,
                timestamp: block.timestamp,
                signatureCount: 0
            });
            
            emit RecoveryAddressChangeProposed(msg.sender, newRecoveryAddress, proposalId);
        }
        
        // Sign the proposal if not already signed
        if (!recoveryAddressProposalSignatures[proposalId][msg.sender]) {
            recoveryAddressProposalSignatures[proposalId][msg.sender] = true;
            recoveryAddressProposals[proposalId].signatureCount++;
            
            emit RecoveryAddressChangeSignatureAdded(msg.sender, proposalId);
            
            // If quorum reached, change the recovery address
            if (recoveryAddressProposals[proposalId].signatureCount >= quorum) {
                address oldRecoveryAddress = recoveryAddress;
                recoveryAddress = newRecoveryAddress;
                emit RecoveryAddressChanged(oldRecoveryAddress, newRecoveryAddress, proposalId);
            }
        }
        
        return proposalId;
    }
    
    /**
     * @dev Sign an existing recovery address change proposal
     * @param proposalId The proposal ID to sign
     */
    function signRecoveryAddressChange(bytes32 proposalId) external onlySigner notInRecovery {
        RecoveryAddressProposal storage proposal = recoveryAddressProposals[proposalId];
        require(proposal.timestamp > 0, "Proposal does not exist");
        require(!recoveryAddressProposalSignatures[proposalId][msg.sender], "Already signed");
        
        recoveryAddressProposalSignatures[proposalId][msg.sender] = true;
        proposal.signatureCount++;
        
        emit RecoveryAddressChangeSignatureAdded(msg.sender, proposalId);
        
        // If quorum reached, change the recovery address
        if (proposal.signatureCount >= quorum) {
            address oldRecoveryAddress = recoveryAddress;
            recoveryAddress = proposal.proposedAddress;
            emit RecoveryAddressChanged(oldRecoveryAddress, recoveryAddress, proposalId);
        }
    }
    
    // Receive function to accept native coin
    receive() external payable notInRecovery {
        emit Deposited(address(0), msg.sender, msg.value);
    }
    
    /**
     * @dev Create a withdrawal request
     * @param token The token to withdraw (address(0) for native coin)
     * @param amount The amount to withdraw
     * @param to The recipient address
     * @return requestId The unique identifier for this withdrawal request
     */
    function requestWithdrawal(
        address token,
        uint256 amount,
        address to
    ) external onlySigner notInRecovery returns (bytes32) {
        require(to != address(0), "Invalid recipient");
        require(amount > 0, "Invalid amount");
        
        uint256 nonce = withdrawalNonce++;
        
        bytes32 requestId = keccak256(
            abi.encode(
                token, 
                amount, 
                to, 
                nonce,
                block.chainid,
                address(this)
            )
        );
        
        require(withdrawalRequests[requestId].timestamp == 0, "Request exists");
        
        withdrawalRequests[requestId] = WithdrawalRequest({
            token: token,
            amount: amount,
            to: to,
            timestamp: block.timestamp,
            executed: false,
            nonce: nonce,
            signatureCount: 1
        });
        
        withdrawalSignatures[requestId][msg.sender] = true;
        
        emit WithdrawalRequested(requestId, token, amount, to, nonce);
        emit WithdrawalSigned(requestId, msg.sender);
        
        // If quorum is 1, execute immediately
        if (quorum == 1) {
            _executeWithdrawal(requestId);
        }
        
        return requestId;
    }
    
    /**
     * @dev Sign an existing withdrawal request
     * @param requestId The withdrawal request ID to sign
     */
    function signWithdrawal(bytes32 requestId) external onlySigner notInRecovery {
        WithdrawalRequest storage request = withdrawalRequests[requestId];
        require(request.timestamp > 0, "Request not found");
        require(!request.executed, "Already executed");
        require(request.timestamp + WITHDRAWAL_EXPIRATION >= block.timestamp, "Request expired");
        require(!withdrawalSignatures[requestId][msg.sender], "Already signed");
        
        withdrawalSignatures[requestId][msg.sender] = true;
        request.signatureCount++;
        
        emit WithdrawalSigned(requestId, msg.sender);
        
        // Execute if quorum reached
        if (request.signatureCount >= quorum) {
            _executeWithdrawal(requestId);
        }
    }
    
    /**
     * @dev Internal function to execute a withdrawal that has reached quorum
     * @param requestId The withdrawal request ID to execute
     */
    function _executeWithdrawal(bytes32 requestId) internal nonReentrant {
        WithdrawalRequest storage request = withdrawalRequests[requestId];
        require(request.timestamp > 0, "Request not found");
        require(!request.executed, "Already executed");
        require(request.signatureCount >= quorum, "Not enough signatures");
        require(request.timestamp + WITHDRAWAL_EXPIRATION >= block.timestamp, "Request expired");
        
        uint256 amount = request.amount;
        address to = request.to;
        address token = request.token;
        
        // Check balances before marking as executed
        if (token == address(0)) {
            require(address(this).balance >= amount, "Insufficient balance");
        } else {
            require(IERC20(token).balanceOf(address(this)) >= amount, "Insufficient token balance");
        }
        
        // Mark as executed before transfer to prevent reentrancy
        request.executed = true;
        
        // Execute transfer
        if (token == address(0)) {
            (bool success, ) = to.call{value: amount}("");
            require(success, "Transfer failed");
        } else {
            IERC20(token).safeTransfer(to, amount);
        }
        
        emit WithdrawalExecuted(requestId, token, amount, to);
    }
    
    // View functions
    function getBalance() external view returns (uint256) {
        return address(this).balance;
    }
    
    function getTokenBalance(address token) external view returns (uint256) {
        require(token != address(0), "Use getBalance() for native coin");
        return IERC20(token).balanceOf(address(this));
    }
    
    /**
     * @dev Check if a withdrawal request has reached quorum
     * @param requestId The withdrawal request ID to check
     * @return True if the request has reached quorum
     */
    function hasReachedQuorum(bytes32 requestId) external view returns (bool) {
        WithdrawalRequest storage request = withdrawalRequests[requestId];
        return request.signatureCount >= quorum;
    }
    
    /**
     * @dev Check if a signer has signed a withdrawal request
     * @param requestId The withdrawal request ID to check
     * @param signer The signer address to check
     * @return True if the signer has signed the request
     */
    function hasSignedWithdrawal(bytes32 requestId, address signer) external view returns (bool) {
        return withdrawalSignatures[requestId][signer];
    }
    
    /**
     * @dev Check if a recovery address proposal has reached quorum
     * @param proposalId The proposal ID to check
     * @return True if the proposal has reached quorum
     */
    function hasRecoveryAddressProposalReachedQuorum(bytes32 proposalId) external view returns (bool) {
        RecoveryAddressProposal storage proposal = recoveryAddressProposals[proposalId];
        return proposal.signatureCount >= quorum;
    }
    
    /**
     * @dev Check if a signer has signed a recovery address change proposal
     * @param proposalId The proposal ID to check
     * @param signer The signer address to check
     * @return True if the signer has signed the proposal
     */
    function hasSignedRecoveryAddressProposal(bytes32 proposalId, address signer) external view returns (bool) {
        return recoveryAddressProposalSignatures[proposalId][signer];
    }
}
