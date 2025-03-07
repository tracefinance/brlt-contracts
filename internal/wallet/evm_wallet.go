package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"vault0/internal/config"
	"vault0/internal/keystore"
	"vault0/internal/types"
)

// EVMConfig contains EVM chain specific configuration
type EVMConfig struct {
	// ChainID is the EVM chain ID
	ChainID *big.Int
	// GasLimit is the default gas limit for transactions
	DefaultGasLimit uint64
	// DefaultGasPrice is the default gas price for transactions
	DefaultGasPrice *big.Int
}

// AppConfig is the interface for application configuration
type AppConfig interface {
	GetBlockchainConfig(chainType string) *config.BlockchainConfig
}

// NewEVMConfig returns configuration for EVM based on chain type and app config
func NewEVMConfig(chainType types.ChainType, appConfig AppConfig) (*EVMConfig, error) {
	// Ensure appConfig is never nil
	if appConfig == nil {
		panic("appConfig must not be nil")
	}

	config := &EVMConfig{}

	// Get the blockchain config using the helper method
	blockchainConfig := appConfig.GetBlockchainConfig(string(chainType))
	if blockchainConfig == nil {
		// Return an error instead of creating a default configuration
		return nil, fmt.Errorf("blockchain configuration for %s not found: %w", chainType, types.ErrUnsupportedChain)
	}

	// Set chain ID from config if available
	if blockchainConfig.ChainID != 0 {
		config.ChainID = big.NewInt(blockchainConfig.ChainID)
	} else {
		return nil, fmt.Errorf("chain ID is required for %s", chainType)
	}

	// Set gas limit from config if available
	if blockchainConfig.DefaultGasLimit != 0 {
		config.DefaultGasLimit = blockchainConfig.DefaultGasLimit
	} else {
		config.DefaultGasLimit = 21000 // Default gas limit for simple transfers
	}

	// Set gas price from config if available
	if blockchainConfig.DefaultGasPrice != 0 {
		config.DefaultGasPrice = big.NewInt(blockchainConfig.DefaultGasPrice)
	} else {
		config.DefaultGasPrice = big.NewInt(20000000000) // 20 Gwei default
	}

	return config, nil
}

// EVMWallet implements the Wallet interface for EVM-compatible chains
type EVMWallet struct {
	keyStore  keystore.KeyStore
	chainType types.ChainType
	config    *EVMConfig
	keyID     string
}

// NewEVMWallet creates a new EVMWallet instance
func NewEVMWallet(keyStore keystore.KeyStore, chainType types.ChainType, keyID string, appConfig AppConfig) (*EVMWallet, error) {
	if keyStore == nil {
		return nil, fmt.Errorf("keystore cannot be nil")
	}

	if keyID == "" {
		return nil, fmt.Errorf("keyID cannot be empty")
	}

	config, err := NewEVMConfig(chainType, appConfig)
	if err != nil {
		return nil, err
	}

	return &EVMWallet{
		keyStore:  keyStore,
		chainType: chainType,
		config:    config,
		keyID:     keyID,
	}, nil
}

// ChainType returns the wallet's blockchain type
func (w *EVMWallet) ChainType() types.ChainType {
	return w.chainType
}

// DeriveAddress derives a wallet address using the wallet's keyID
func (w *EVMWallet) DeriveAddress(ctx context.Context) (string, error) {
	// Get the public key from the keystore
	key, err := w.keyStore.GetPublicKey(ctx, w.keyID)
	if err != nil {
		return "", fmt.Errorf("evm: failed to get public key for key ID %s: %w", w.keyID, err)
	}

	publicKey := key.PublicKey

	// For EVM chains, we need to convert the public key to an address
	if len(publicKey) == 0 {
		return "", fmt.Errorf("evm: empty public key: %w", types.ErrInvalidAddress)
	}

	// Check if the public key already has the 0x04 prefix
	// This prefix indicates an uncompressed public key which is what EVM expects
	var pubKey *ecdsa.PublicKey

	if len(publicKey) == 65 && publicKey[0] == 0x04 {
		// Public key already has the right format
		pubKey, err = crypto.UnmarshalPubkey(publicKey)
	} else if len(publicKey) == 64 {
		// Public key might be the raw 64 bytes without the prefix
		// Add the 0x04 prefix for uncompressed public key
		prefixedKey := append([]byte{0x04}, publicKey...)
		pubKey, err = crypto.UnmarshalPubkey(prefixedKey)
	} else if len(publicKey) == 33 && (publicKey[0] == 0x02 || publicKey[0] == 0x03) {
		// Compressed public key, we need to decompress it
		return "", fmt.Errorf("evm: compressed public keys not supported: %w", types.ErrInvalidAddress)
	} else {
		// Unknown format
		return "", fmt.Errorf("evm: invalid public key format: %w", types.ErrInvalidAddress)
	}

	if err != nil {
		return "", fmt.Errorf("evm: failed to parse public key: %w", err)
	}

	address := crypto.PubkeyToAddress(*pubKey)
	return address.Hex(), nil
}

// CreateNativeTransaction creates a native currency transaction without broadcasting
func (w *EVMWallet) CreateNativeTransaction(ctx context.Context, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	// Derive the from address from the wallet's keyID
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to derive from address: %w", err)
	}

	if !common.IsHexAddress(toAddress) {
		return nil, fmt.Errorf("%w: %s", types.ErrInvalidAddress, toAddress)
	}

	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, types.ErrInvalidAmount
	}

	// Set default values if not provided
	gasPrice := options.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice = w.config.DefaultGasPrice
	}

	gasLimit := options.GasLimit
	if gasLimit == 0 {
		gasLimit = w.config.DefaultGasLimit
	}

	// Create the transaction
	tx := &types.Transaction{
		Chain:    w.chainType,
		From:     fromAddress,
		To:       toAddress,
		Value:    amount,
		Data:     options.Data,
		Nonce:    options.Nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Type:     types.TransactionTypeNative,
	}

	return tx, nil
}

// CreateTokenTransaction creates an ERC20 token transaction without broadcasting
func (w *EVMWallet) CreateTokenTransaction(ctx context.Context, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	// Derive the from address from the wallet's keyID
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to derive from address: %w", err)
	}

	if !common.IsHexAddress(toAddress) || !common.IsHexAddress(tokenAddress) {
		return nil, fmt.Errorf("%w: invalid address format", types.ErrInvalidAddress)
	}

	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, types.ErrInvalidAmount
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

	// Set default values if not provided
	gasPrice := options.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice = w.config.DefaultGasPrice
	}

	gasLimit := options.GasLimit
	if gasLimit == 0 {
		gasLimit = 65000 // Default for ERC20 transfers
	}

	// Create the transaction
	tx := &types.Transaction{
		Chain:        w.chainType,
		From:         fromAddress,
		To:           tokenAddress,
		Value:        big.NewInt(0), // 0 ETH for token transfers
		Data:         data,
		Nonce:        options.Nonce,
		GasPrice:     gasPrice,
		GasLimit:     gasLimit,
		Type:         types.TransactionTypeERC20,
		TokenAddress: tokenAddress,
	}

	return tx, nil
}

// SignTransaction signs a transaction
func (w *EVMWallet) SignTransaction(ctx context.Context, tx *types.Transaction) ([]byte, error) {
	// Derive address from public key using the wallet's keyID
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	// Verify that the from address matches the derived address
	if !strings.EqualFold(fromAddress, tx.From) {
		return nil, fmt.Errorf("%w: transaction from address does not match key", types.ErrInvalidAddress)
	}

	// Create Ethereum transaction
	toAddress := common.HexToAddress(tx.To)

	// Create the appropriate transaction based on type
	ethTx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		Gas:      tx.GasLimit,
		To:       &toAddress,
		Value:    tx.Value,
		Data:     tx.Data,
	})

	// Sign the transaction
	return w.signEVMTransaction(ctx, ethTx)
}

// signEVMTransaction signs an EVM transaction using the key store
func (w *EVMWallet) signEVMTransaction(ctx context.Context, tx *ethtypes.Transaction) ([]byte, error) {
	// Hash the transaction for signing
	signer := ethtypes.NewEIP155Signer(w.config.ChainID)
	hash := signer.Hash(tx)

	// Sign the hash with the keystore
	signature, err := w.keyStore.Sign(ctx, w.keyID, hash.Bytes(), keystore.DataTypeDigest)
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
