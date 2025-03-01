// SPDX-License-Identifier: MIT
pragma solidity ^0.8.28;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";

/**
 * @title MultiSigWallet
 * @dev A wallet that requires two signatures (client and manager) to withdraw funds
 * with a recovery mechanism that allows funds to be sent to a recovery address
 * after a 72-hour timelock period
 */
contract MultiSigWallet is ReentrancyGuard {
    using SafeERC20 for IERC20;

    address public immutable manager;
    address public immutable client;
    address public immutable recoveryAddress;
    
    uint256 public constant RECOVERY_DELAY = 72 hours;
    uint256 public constant MAX_BATCH_SIZE = 20;
    uint256 public constant WITHDRAWAL_EXPIRATION = 24 hours;
    uint256 public recoveryRequestTimestamp;
    bool public recoveryExecuted;
    uint256 public withdrawalNonce;
    
    // Array to store supported tokens for recovery
    address[] public supportedTokensList;
    // Mapping to store supported tokens for automatic recovery
    mapping(address => bool) public supportedTokens;
    
    // Mapping to store withdrawal requests
    struct WithdrawalRequest {
        address token;      // Address of token (address(0) for native coin)
        uint256 amount;     // Amount to withdraw
        address to;         // Recipient address
        bool managerSigned; // Whether manager has signed
        bool clientSigned;  // Whether client has signed
        uint256 timestamp;  // When the request was created
        bool executed;      // Whether the request was executed
        uint256 nonce;     // Unique nonce for the request
    }
    
    mapping(bytes32 => WithdrawalRequest) public withdrawalRequests;
    mapping(address => uint256) public recoveryAttempts;
    
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
    
    modifier onlyManager() {
        require(msg.sender == manager, "Only manager can call this function");
        _;
    }
    
    modifier onlyClient() {
        require(msg.sender == client, "Only client can call this function");
        _;
    }
    
    modifier onlyAuthorized() {
        require(msg.sender == manager || msg.sender == client, "Unauthorized");
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
    
    constructor(address _client, address _recoveryAddress) {
        require(_client != address(0), "Invalid client address");
        require(_recoveryAddress != address(0), "Invalid recovery address");
        manager = msg.sender;
        client = _client;
        recoveryAddress = _recoveryAddress;
        
        // Accept native coin (ETH) by default
        supportedTokens[address(0)] = true;
        supportedTokensList.push(address(0));
    }
    
    /**
     * @dev Allows the manager to add a token to the supported tokens list for automatic recovery
     * @param token The token address to support
     */
    function addSupportedToken(address token) external onlyManager {
        require(!supportedTokens[token], "Token already supported");
        supportedTokens[token] = true;
        supportedTokensList.push(token);
        emit TokenSupported(token);
    }
    
    /**
     * @dev Allows the manager to remove a token from the supported tokens list
     * @param token The token address to remove
     */
    function removeSupportedToken(address token) external onlyManager {
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
     * @dev Allows the manager to recover non-supported tokens after recovery is completed
     * @param token The token address to recover
     * @param to The address to send the recovered tokens to
     */
    function recoverNonSupportedToken(address token, address to) external onlyManager recoveryCompleted nonReentrant {
        require(token != address(0), "Cannot recover native coin");
        require(!supportedTokens[token], "Use regular recovery for supported tokens");
        require(to != address(0), "Invalid recipient address");
        
        uint256 balance = IERC20(token).balanceOf(address(this));
        require(balance > 0, "No balance to recover");
        
        IERC20(token).safeTransfer(to, balance);
        emit NonSupportedTokenRecovered(token, balance, to);
    }
    
    /**
     * @dev Returns the list of all supported tokens
     * @return Array of supported token addresses
     */
    function getSupportedTokens() external view returns (address[] memory) {
        return supportedTokensList;
    }
    
    // Recovery functions
    function requestRecovery() external onlyManager {
        require(recoveryRequestTimestamp == 0, "Recovery already requested");
        require(!recoveryExecuted, "Recovery already executed");
        
        recoveryRequestTimestamp = block.timestamp;
        emit RecoveryRequested(block.timestamp);
    }
    
    function cancelRecovery() external onlyClient {
        require(recoveryRequestTimestamp > 0, "No recovery requested");
        require(!recoveryExecuted, "Recovery already executed");
        require(block.timestamp < recoveryRequestTimestamp + RECOVERY_DELAY, "Recovery period expired");
        
        recoveryRequestTimestamp = 0;
        recoveryExecuted = false;
        emit RecoveryCancelled();
    }
    
    // Internal function to check recovery status
    function _checkRecoveryStatus() internal view {
        require(msg.sender == manager, "Only manager can execute recovery");
        require(recoveryRequestTimestamp > 0, "No recovery requested");
        require(!recoveryExecuted, "Recovery already executed");
        require(block.timestamp >= recoveryRequestTimestamp + RECOVERY_DELAY, "Recovery delay not elapsed");
    }
    
    function executeRecovery() external nonReentrant onlyManager {
        _checkRecoveryStatus();
        
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
    
    // Receive function to accept native coin
    receive() external payable notInRecovery {
        emit Deposited(address(0), msg.sender, msg.value);
    }
    
    // Function to deposit ERC20 tokens
    function depositToken(address token, uint256 amount) external notInRecovery {
        require(token != address(0), "Use receive() for native coin");
        require(amount > 0, "Amount must be greater than 0");
        
        // Use SafeERC20 for the transfer
        IERC20(token).safeTransferFrom(msg.sender, address(this), amount);
        emit Deposited(token, msg.sender, amount);
    }
    
    // Create a withdrawal request
    function requestWithdrawal(
        address token,
        uint256 amount,
        address to
    ) external onlyAuthorized notInRecovery returns (bytes32) {
        require(to != address(0), "Invalid recipient");
        require(amount > 0, "Invalid amount");
        
        uint256 nonce = withdrawalNonce++;
        bytes32 requestId = keccak256(
            abi.encodePacked(
                token, 
                amount, 
                to, 
                block.timestamp,
                nonce,
                block.chainid
            )
        );
        
        require(withdrawalRequests[requestId].timestamp == 0, "Request exists");
        
        withdrawalRequests[requestId] = WithdrawalRequest({
            token: token,
            amount: amount,
            to: to,
            managerSigned: msg.sender == manager,
            clientSigned: msg.sender == client,
            timestamp: block.timestamp,
            executed: false,
            nonce: nonce
        });
        
        emit WithdrawalRequested(requestId, token, amount, to, nonce);
        emit WithdrawalSigned(requestId, msg.sender);
        
        return requestId;
    }
    
    // Sign a withdrawal request
    function signWithdrawal(bytes32 requestId) external onlyAuthorized {
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
        
        // If both have signed, execute the withdrawal
        if (request.managerSigned && request.clientSigned) {
            _executeWithdrawal(requestId);
        }
    }
    
    // Internal function to execute withdrawal
    function _executeWithdrawal(bytes32 requestId) internal nonReentrant {
        WithdrawalRequest storage request = withdrawalRequests[requestId];
        require(request.timestamp > 0, "Request not found");
        require(!request.executed, "Already executed");
        require(request.managerSigned && request.clientSigned, "Not fully signed");
        require(request.timestamp + WITHDRAWAL_EXPIRATION >= block.timestamp, "Request expired");
        
        uint256 amount = request.amount;
        address to = request.to;
        address token = request.token;
        
        // Mark as executed before transfer to prevent reentrancy
        request.executed = true;
        
        // Execute transfer
        if (token == address(0)) {
            require(address(this).balance >= amount, "Insufficient balance");
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
}
