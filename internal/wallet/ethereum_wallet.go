package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"vault0/internal/config"
	"vault0/internal/keystore"
)

// EthereumConfig contains Ethereum chain specific configuration
type EthereumConfig struct {
	// RPC URL for the Ethereum node
	RPCURL string
	// ChainID is the Ethereum chain ID
	ChainID *big.Int
	// GasLimit is the default gas limit for transactions
	DefaultGasLimit uint64
	// DefaultGasPrice is the default gas price for transactions
	DefaultGasPrice *big.Int
}

// NewEthereumConfig returns configuration for Ethereum based on chain type and app config
func NewEthereumConfig(chainType ChainType, appConfig *config.Config) *EthereumConfig {
	// Ensure appConfig is never nil
	if appConfig == nil {
		panic("appConfig must not be nil")
	}

	config := &EthereumConfig{}

	// Get the blockchain config using the helper method
	blockchainConfig := appConfig.GetBlockchainConfig(string(chainType))
	if blockchainConfig == nil {
		// Create a minimal default configuration if no specific config found
		config.DefaultGasLimit = 21000                   // Standard ETH transfer gas limit
		config.DefaultGasPrice = big.NewInt(20000000000) // 20 Gwei
		return config
	}

	// Set configuration from blockchain config
	config.RPCURL = blockchainConfig.RPCURL
	config.ChainID = big.NewInt(blockchainConfig.ChainID)
	config.DefaultGasLimit = blockchainConfig.DefaultGasLimit
	// Convert Gwei to Wei (1 Gwei = 10^9 Wei)
	config.DefaultGasPrice = big.NewInt(blockchainConfig.DefaultGasPrice * 1e9)

	return config
}

// EthereumWallet implements the Wallet interface for Ethereum and EVM-compatible chains
type EthereumWallet struct {
	keyStore  keystore.KeyStore
	chainType ChainType
	config    *EthereumConfig
	client    *ethclient.Client
}

// NewEthereumWallet creates a new Ethereum wallet
func NewEthereumWallet(keyStore keystore.KeyStore, chainType ChainType, config interface{}, appConfig *config.Config) (*EthereumWallet, error) {
	// Ensure appConfig is never nil
	if appConfig == nil {
		return nil, fmt.Errorf("appConfig must not be nil")
	}

	var ethConfig *EthereumConfig

	if config != nil {
		var ok bool
		ethConfig, ok = config.(*EthereumConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for Ethereum wallet")
		}
	} else {
		ethConfig = NewEthereumConfig(chainType, appConfig)
	}

	if ethConfig.RPCURL == "" {
		return nil, fmt.Errorf("RPC URL is required")
	}
	if ethConfig.ChainID == nil {
		return nil, fmt.Errorf("Chain ID is required")
	}

	client, err := ethclient.Dial(ethConfig.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ethereum node: %w", err)
	}

	return &EthereumWallet{
		keyStore:  keyStore,
		chainType: chainType,
		config:    ethConfig,
		client:    client,
	}, nil
}

// ChainType returns the blockchain type
func (w *EthereumWallet) ChainType() ChainType {
	return w.chainType
}

// DeriveAddress derives a wallet address from a public key
func (w *EthereumWallet) DeriveAddress(ctx context.Context, publicKey []byte) (string, error) {
	if len(publicKey) != 65 && len(publicKey) != 33 {
		return "", fmt.Errorf("%w: invalid public key length", ErrInvalidAddress)
	}

	var pubKey *ecdsa.PublicKey
	var err error

	// Check if the public key is compressed (33 bytes) or uncompressed (65 bytes)
	if len(publicKey) == 33 {
		// Decompress the public key
		pubKey, err = crypto.DecompressPubkey(publicKey)
		if err != nil {
			return "", fmt.Errorf("failed to decompress public key: %w", err)
		}
	} else {
		// Uncompressed public key
		pubKey, err = crypto.UnmarshalPubkey(publicKey)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal public key: %w", err)
		}
	}

	// Derive Ethereum address from public key
	address := crypto.PubkeyToAddress(*pubKey)
	return address.Hex(), nil
}

// GetBalance returns the native currency balance of an address
func (w *EthereumWallet) GetBalance(ctx context.Context, address string) (*big.Int, error) {
	if !common.IsHexAddress(address) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAddress, address)
	}

	ethAddress := common.HexToAddress(address)
	balance, err := w.client.BalanceAt(ctx, ethAddress, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	return balance, nil
}

// GetTokenBalance returns the token balance of an address
func (w *EthereumWallet) GetTokenBalance(ctx context.Context, address, tokenAddress string) (*big.Int, error) {
	if !common.IsHexAddress(address) || !common.IsHexAddress(tokenAddress) {
		return nil, fmt.Errorf("%w: invalid address format", ErrInvalidAddress)
	}

	// ERC20 balanceOf method ABI
	const erc20ABI = `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"type":"function"}]`

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %w", err)
	}

	// Pack the address to get the calldata
	data, err := parsedABI.Pack("balanceOf", common.HexToAddress(address))
	if err != nil {
		return nil, fmt.Errorf("failed to pack data: %w", err)
	}

	// Create a call message
	tokenAddr := common.HexToAddress(tokenAddress)
	msg := ethereum.CallMsg{
		To:   &tokenAddr,
		Data: data,
	}

	// Make the call
	result, err := w.client.CallContract(ctx, msg, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to call contract: %w", err)
	}

	// Unpack the result
	var balance *big.Int
	err = parsedABI.UnpackIntoInterface(&balance, "balanceOf", result)
	if err != nil {
		return nil, fmt.Errorf("failed to unpack result: %w", err)
	}

	return balance, nil
}

// SignTransaction signs a transaction without broadcasting
func (w *EthereumWallet) SignTransaction(ctx context.Context, keyID string, tx *Transaction) ([]byte, error) {
	// Get the public key
	key, err := w.keyStore.GetPublicKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Derive address from public key
	fromAddress, err := w.DeriveAddress(ctx, key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	// Create Ethereum transaction
	var ethTx *types.Transaction

	// Get latest nonce if not provided
	var nonce uint64
	if tx.Nonce == 0 {
		nonce, err = w.client.PendingNonceAt(ctx, common.HexToAddress(fromAddress))
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}
	} else {
		nonce = tx.Nonce
	}

	// Get gas price if not provided
	gasPrice := tx.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice, err = w.client.SuggestGasPrice(ctx)
		if err != nil {
			gasPrice = w.config.DefaultGasPrice
		}
	}

	// Get gas limit if not provided
	gasLimit := tx.GasLimit
	if gasLimit == 0 {
		// Use default gas limit based on transaction type
		switch tx.Type {
		case TransactionTypeNative:
			gasLimit = w.config.DefaultGasLimit
		case TransactionTypeERC20:
			gasLimit = 65000 // Default for ERC20 transfers
		default:
			gasLimit = 100000 // Higher default for contract interactions
		}
	}

	toAddress := common.HexToAddress(tx.To)
	value := tx.Value
	data := tx.Data

	// Create the appropriate transaction based on type
	ethTx = types.NewTx(&types.LegacyTx{
		Nonce:    nonce,
		GasPrice: gasPrice,
		Gas:      gasLimit,
		To:       &toAddress,
		Value:    value,
		Data:     data,
	})

	// Sign the transaction
	txBytes, err := w.signEthereumTransaction(ctx, keyID, ethTx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	return txBytes, nil
}

// signEthereumTransaction signs an Ethereum transaction using the key store
func (w *EthereumWallet) signEthereumTransaction(ctx context.Context, keyID string, tx *types.Transaction) ([]byte, error) {
	// Hash the transaction for signing
	signer := types.NewEIP155Signer(w.config.ChainID)
	hash := signer.Hash(tx)

	// Sign the hash with the keystore
	signature, err := w.keyStore.Sign(ctx, keyID, hash.Bytes())
	if err != nil {
		return nil, fmt.Errorf("keystore signing failed: %w", err)
	}

	// The signature needs to have the recovery ID as the last byte
	// Ethereum expects this in a specific format with v = 27 + recovery_id
	if len(signature) != 65 {
		return nil, fmt.Errorf("invalid signature length: %d", len(signature))
	}

	// Adjust v according to EIP-155
	v := signature[64]
	signature[64] = v + byte(w.config.ChainID.Uint64()*2+35)

	// Create a signed transaction
	r := new(big.Int).SetBytes(signature[:32])
	s := new(big.Int).SetBytes(signature[32:64])
	v_int := new(big.Int).SetBytes([]byte{signature[64]})

	signedTx, err := tx.WithSignature(signer, append(append(r.Bytes(), s.Bytes()...), v_int.Bytes()...))
	if err != nil {
		return nil, fmt.Errorf("failed to add signature to transaction: %w", err)
	}

	// Encode the signed transaction
	txBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return nil, fmt.Errorf("failed to encode transaction: %w", err)
	}

	return txBytes, nil
}

// BroadcastTransaction broadcasts a signed transaction
func (w *EthereumWallet) BroadcastTransaction(ctx context.Context, signedTx []byte) (*Transaction, error) {
	// Decode the transaction
	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(signedTx); err != nil {
		return nil, fmt.Errorf("failed to decode transaction: %w", err)
	}

	// Send the transaction
	err := w.client.SendTransaction(ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTransactionFailed, err)
	}

	// Get the from address
	var from string
	signer := types.NewEIP155Signer(w.config.ChainID)
	if sender, err := types.Sender(signer, tx); err == nil {
		from = sender.Hex()
	}

	// Create a Transaction object
	result := &Transaction{
		Chain:    w.chainType,
		Hash:     tx.Hash().Hex(),
		From:     from,
		To:       tx.To().Hex(),
		Value:    tx.Value(),
		Data:     tx.Data(),
		Nonce:    tx.Nonce(),
		GasPrice: tx.GasPrice(),
		GasLimit: tx.Gas(),
		Status:   "pending",
	}

	return result, nil
}

// GetTransaction retrieves a transaction by hash
func (w *EthereumWallet) GetTransaction(ctx context.Context, hash string) (*Transaction, error) {
	if !strings.HasPrefix(hash, "0x") {
		hash = "0x" + hash
	}

	// Convert the hash string to an Ethereum hash
	txHash := common.HexToHash(hash)

	// Get the transaction
	tx, isPending, err := w.client.TransactionByHash(ctx, txHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Get the transaction receipt for status
	var status string
	var receipt *types.Receipt
	if !isPending {
		receipt, err = w.client.TransactionReceipt(ctx, txHash)
		if err != nil {
			status = "unknown"
		} else if receipt.Status == 1 {
			status = "success"
		} else {
			status = "failed"
		}
	} else {
		status = "pending"
	}

	// Get the sender address
	var from string
	signer := types.NewEIP155Signer(w.config.ChainID)
	if sender, err := types.Sender(signer, tx); err == nil {
		from = sender.Hex()
	}

	// Create a Transaction object
	result := &Transaction{
		Chain:    w.chainType,
		Hash:     tx.Hash().Hex(),
		From:     from,
		To:       tx.To().Hex(),
		Value:    tx.Value(),
		Data:     tx.Data(),
		Nonce:    tx.Nonce(),
		GasPrice: tx.GasPrice(),
		GasLimit: tx.Gas(),
		Status:   status,
	}

	// Get the block timestamp if available
	if !isPending && receipt != nil {
		block, err := w.client.BlockByNumber(ctx, receipt.BlockNumber)
		if err == nil {
			result.Timestamp = int64(block.Time())
		}
	}

	return result, nil
}

// SendNative sends native currency (ETH, MATIC, etc.)
func (w *EthereumWallet) SendNative(ctx context.Context, keyID, toAddress string, amount *big.Int, options *TransactionOptions) (*Transaction, error) {
	if !common.IsHexAddress(toAddress) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAddress, toAddress)
	}

	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get the public key
	key, err := w.keyStore.GetPublicKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Derive address from public key
	fromAddress, err := w.DeriveAddress(ctx, key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	// Get nonce
	var nonce uint64
	if options != nil && options.Nonce != nil {
		nonce = *options.Nonce
	} else {
		nonce, err = w.client.PendingNonceAt(ctx, common.HexToAddress(fromAddress))
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}
	}

	// Get gas price
	var gasPrice *big.Int
	if options != nil && options.GasPrice != nil {
		gasPrice = options.GasPrice
	} else {
		gasPrice, err = w.client.SuggestGasPrice(ctx)
		if err != nil {
			gasPrice = w.config.DefaultGasPrice
		}
	}

	// Get gas limit
	var gasLimit uint64
	if options != nil && options.GasLimit != 0 {
		gasLimit = options.GasLimit
	} else {
		gasLimit = w.config.DefaultGasLimit
	}

	// Check if sender has sufficient balance
	balance, err := w.GetBalance(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance: %w", err)
	}

	// Calculate the transaction fee
	fee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))

	// Check if the balance is sufficient to cover amount + fee
	totalRequired := new(big.Int).Add(amount, fee)
	if balance.Cmp(totalRequired) < 0 {
		return nil, ErrInsufficientBalance
	}

	// Prepare transaction data
	var data []byte
	if options != nil {
		data = options.Data
	}

	// Create the transaction
	tx := &Transaction{
		Chain:    w.chainType,
		From:     fromAddress,
		To:       toAddress,
		Value:    amount,
		Data:     data,
		Nonce:    nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Type:     TransactionTypeNative,
	}

	// Sign the transaction
	signedTx, err := w.SignTransaction(ctx, keyID, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Broadcast the transaction
	return w.BroadcastTransaction(ctx, signedTx)
}

// SendToken sends ERC20 tokens
func (w *EthereumWallet) SendToken(ctx context.Context, keyID, tokenAddress, toAddress string, amount *big.Int, options *TransactionOptions) (*Transaction, error) {
	if !common.IsHexAddress(toAddress) || !common.IsHexAddress(tokenAddress) {
		return nil, fmt.Errorf("%w: invalid address format", ErrInvalidAddress)
	}

	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrInvalidAmount
	}

	// Get the public key
	key, err := w.keyStore.GetPublicKey(ctx, keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get key: %w", err)
	}

	// Derive address from public key
	fromAddress, err := w.DeriveAddress(ctx, key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	// Create the token transfer data
	// ERC20 transfer method ABI
	const transferMethodSignature = "transfer(address,uint256)"
	methodID := crypto.Keccak256([]byte(transferMethodSignature))[:4]

	// Encode the transfer parameters
	paddedAddress := common.LeftPadBytes(common.HexToAddress(toAddress).Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)

	// Combine the data
	data := append(methodID, append(paddedAddress, paddedAmount...)...)

	// Get nonce
	var nonce uint64
	if options != nil && options.Nonce != nil {
		nonce = *options.Nonce
	} else {
		nonce, err = w.client.PendingNonceAt(ctx, common.HexToAddress(fromAddress))
		if err != nil {
			return nil, fmt.Errorf("failed to get nonce: %w", err)
		}
	}

	// Get gas price
	var gasPrice *big.Int
	if options != nil && options.GasPrice != nil {
		gasPrice = options.GasPrice
	} else {
		gasPrice, err = w.client.SuggestGasPrice(ctx)
		if err != nil {
			gasPrice = w.config.DefaultGasPrice
		}
	}

	// Get gas limit for token transfer
	var gasLimit uint64
	if options != nil && options.GasLimit != 0 {
		gasLimit = options.GasLimit
	} else {
		// Estimate gas for token transfer
		fromAddr := common.HexToAddress(fromAddress)
		tokenAddr := common.HexToAddress(tokenAddress)
		msg := ethereum.CallMsg{
			From: fromAddr,
			To:   &tokenAddr,
			Data: data,
		}

		estimatedGas, err := w.client.EstimateGas(ctx, msg)
		if err != nil {
			// Use default if estimation fails
			gasLimit = 65000 // Default for ERC20 transfers
		} else {
			// Add some buffer for safety
			gasLimit = estimatedGas + 10000
		}
	}

	// Check token balance
	tokenBalance, err := w.GetTokenBalance(ctx, fromAddress, tokenAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get token balance: %w", err)
	}

	if tokenBalance.Cmp(amount) < 0 {
		return nil, ErrInsufficientBalance
	}

	// Check if sender has sufficient ETH for gas
	ethBalance, err := w.GetBalance(ctx, fromAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get ETH balance: %w", err)
	}

	// Calculate the transaction fee
	fee := new(big.Int).Mul(gasPrice, big.NewInt(int64(gasLimit)))
	if ethBalance.Cmp(fee) < 0 {
		return nil, fmt.Errorf("%w: insufficient ETH for gas", ErrInsufficientBalance)
	}

	// Create the transaction
	tx := &Transaction{
		Chain:        w.chainType,
		From:         fromAddress,
		To:           tokenAddress,
		Value:        big.NewInt(0), // 0 ETH for token transfers
		Data:         data,
		Nonce:        nonce,
		GasPrice:     gasPrice,
		GasLimit:     gasLimit,
		Type:         TransactionTypeERC20,
		TokenAddress: tokenAddress,
	}

	// Sign the transaction
	signedTx, err := w.SignTransaction(ctx, keyID, tx)
	if err != nil {
		return nil, fmt.Errorf("failed to sign transaction: %w", err)
	}

	// Broadcast the transaction
	return w.BroadcastTransaction(ctx, signedTx)
}
