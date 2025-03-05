package wallet

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"vault0/internal/config"
	"vault0/internal/keystore"
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

// NewEVMConfig returns configuration for EVM based on chain type and app config
func NewEVMConfig(chainType ChainType, appConfig *config.Config) *EVMConfig {
	// Ensure appConfig is never nil
	if appConfig == nil {
		panic("appConfig must not be nil")
	}

	config := &EVMConfig{}

	// Get the blockchain config using the helper method
	blockchainConfig := appConfig.GetBlockchainConfig(string(chainType))
	if blockchainConfig == nil {
		// Create a minimal default configuration if no specific config found
		config.DefaultGasLimit = 21000                   // Standard ETH transfer gas limit
		config.DefaultGasPrice = big.NewInt(20000000000) // 20 Gwei
		return config
	}

	// Set configuration from blockchain config
	config.ChainID = big.NewInt(blockchainConfig.ChainID)
	config.DefaultGasLimit = blockchainConfig.DefaultGasLimit
	// Convert Gwei to Wei (1 Gwei = 10^9 Wei)
	config.DefaultGasPrice = big.NewInt(blockchainConfig.DefaultGasPrice * 1e9)

	return config
}

// EVMWallet implements the Wallet interface for EVM-compatible chains
type EVMWallet struct {
	keyStore  keystore.KeyStore
	chainType ChainType
	config    *EVMConfig
}

// NewEVMWallet creates a new EVM wallet
func NewEVMWallet(keyStore keystore.KeyStore, chainType ChainType, config interface{}, appConfig *config.Config) (*EVMWallet, error) {
	// Ensure appConfig is never nil
	if appConfig == nil {
		return nil, fmt.Errorf("appConfig must not be nil")
	}

	var evmConfig *EVMConfig

	if config != nil {
		var ok bool
		evmConfig, ok = config.(*EVMConfig)
		if !ok {
			return nil, fmt.Errorf("invalid config type for EVM wallet")
		}
	} else {
		evmConfig = NewEVMConfig(chainType, appConfig)
	}

	if evmConfig.ChainID == nil {
		return nil, fmt.Errorf("Chain ID is required")
	}

	return &EVMWallet{
		keyStore:  keyStore,
		chainType: chainType,
		config:    evmConfig,
	}, nil
}

// ChainType returns the blockchain type
func (w *EVMWallet) ChainType() ChainType {
	return w.chainType
}

// DeriveAddress derives a wallet address from a public key
func (w *EVMWallet) DeriveAddress(ctx context.Context, publicKey []byte) (string, error) {
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

// CreateNativeTransaction creates a native currency transaction without broadcasting
func (w *EVMWallet) CreateNativeTransaction(ctx context.Context, fromAddress, toAddress string, amount *big.Int, options TransactionOptions) (*Transaction, error) {
	if !common.IsHexAddress(toAddress) {
		return nil, fmt.Errorf("%w: %s", ErrInvalidAddress, toAddress)
	}

	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrInvalidAmount
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
	tx := &Transaction{
		Chain:    w.chainType,
		From:     fromAddress,
		To:       toAddress,
		Value:    amount,
		Data:     options.Data,
		Nonce:    options.Nonce,
		GasPrice: gasPrice,
		GasLimit: gasLimit,
		Type:     TransactionTypeNative,
	}

	return tx, nil
}

// CreateTokenTransaction creates an ERC20 token transaction without broadcasting
func (w *EVMWallet) CreateTokenTransaction(ctx context.Context, fromAddress, tokenAddress, toAddress string, amount *big.Int, options TransactionOptions) (*Transaction, error) {
	if !common.IsHexAddress(toAddress) || !common.IsHexAddress(tokenAddress) {
		return nil, fmt.Errorf("%w: invalid address format", ErrInvalidAddress)
	}

	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, ErrInvalidAmount
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
	tx := &Transaction{
		Chain:        w.chainType,
		From:         fromAddress,
		To:           tokenAddress,
		Value:        big.NewInt(0), // 0 ETH for token transfers
		Data:         data,
		Nonce:        options.Nonce,
		GasPrice:     gasPrice,
		GasLimit:     gasLimit,
		Type:         TransactionTypeERC20,
		TokenAddress: tokenAddress,
	}

	return tx, nil
}

// SignTransaction signs a transaction without broadcasting
func (w *EVMWallet) SignTransaction(ctx context.Context, keyID string, tx *Transaction) ([]byte, error) {
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

	// Verify that the from address matches the derived address
	if !strings.EqualFold(fromAddress, tx.From) {
		return nil, fmt.Errorf("%w: transaction from address does not match key", ErrInvalidAddress)
	}

	// Create Ethereum transaction
	toAddress := common.HexToAddress(tx.To)

	// Create the appropriate transaction based on type
	ethTx := types.NewTx(&types.LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		Gas:      tx.GasLimit,
		To:       &toAddress,
		Value:    tx.Value,
		Data:     tx.Data,
	})

	// Sign the transaction
	return w.signEVMTransaction(ctx, keyID, ethTx)
}

// signEVMTransaction signs an EVM transaction using the key store
func (w *EVMWallet) signEVMTransaction(ctx context.Context, keyID string, tx *types.Transaction) ([]byte, error) {
	// Hash the transaction for signing
	signer := types.NewEIP155Signer(w.config.ChainID)
	hash := signer.Hash(tx)

	// Sign the hash with the keystore
	signature, err := w.keyStore.Sign(ctx, keyID, hash.Bytes(), keystore.DataTypeDigest)
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
