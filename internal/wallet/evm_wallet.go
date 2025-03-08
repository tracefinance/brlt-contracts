package wallet

import (
	"context"
	"encoding/asn1"
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

type EVMConfig struct {
	ChainID         *big.Int
	DefaultGasLimit uint64
	DefaultGasPrice *big.Int
}

type AppConfig interface {
	GetBlockchainConfig(chainType string) *config.BlockchainConfig
}

func NewEVMConfig(chainType types.ChainType, appConfig AppConfig) (*EVMConfig, error) {
	if appConfig == nil {
		panic("appConfig must not be nil")
	}

	config := &EVMConfig{}
	blockchainConfig := appConfig.GetBlockchainConfig(string(chainType))
	if blockchainConfig == nil {
		return nil, fmt.Errorf("blockchain configuration for %s not found: %w", chainType, types.ErrUnsupportedChain)
	}

	if blockchainConfig.ChainID != 0 {
		config.ChainID = big.NewInt(blockchainConfig.ChainID)
	} else {
		return nil, fmt.Errorf("chain ID is required for %s", chainType)
	}

	if blockchainConfig.DefaultGasLimit != 0 {
		config.DefaultGasLimit = blockchainConfig.DefaultGasLimit
	} else {
		config.DefaultGasLimit = 21000
	}

	if blockchainConfig.DefaultGasPrice != 0 {
		config.DefaultGasPrice = big.NewInt(blockchainConfig.DefaultGasPrice)
	} else {
		config.DefaultGasPrice = big.NewInt(20000000000)
	}

	return config, nil
}

type EVMWallet struct {
	keyStore  keystore.KeyStore
	chainType types.ChainType
	config    *EVMConfig
	keyID     string
}

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

func (w *EVMWallet) ChainType() types.ChainType {
	return w.chainType
}

func (w *EVMWallet) DeriveAddress(ctx context.Context) (string, error) {
	key, err := w.keyStore.GetPublicKey(ctx, w.keyID)
	if err != nil {
		return "", fmt.Errorf("evm: failed to get public key for key ID %s: %w", w.keyID, err)
	}

	publicKey := key.PublicKey
	if len(publicKey) == 0 {
		return "", fmt.Errorf("evm: empty public key: %w", types.ErrInvalidAddress)
	}

	pubKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return "", fmt.Errorf("evm: failed to unmarshal public key: %w", err)
	}

	address := crypto.PubkeyToAddress(*pubKey)
	return address.Hex(), nil
}

func (w *EVMWallet) CreateNativeTransaction(ctx context.Context, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
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

	gasPrice := options.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice = w.config.DefaultGasPrice
	}

	gasLimit := options.GasLimit
	if gasLimit == 0 {
		gasLimit = w.config.DefaultGasLimit
	}

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

func (w *EVMWallet) CreateTokenTransaction(ctx context.Context, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
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

	const transferMethodSignature = "transfer(address,uint256)"
	methodID := crypto.Keccak256([]byte(transferMethodSignature))[:4]
	paddedAddress := common.LeftPadBytes(common.HexToAddress(toAddress).Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amount.Bytes(), 32)
	data := append(methodID, append(paddedAddress, paddedAmount...)...)

	gasPrice := options.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice = w.config.DefaultGasPrice
	}

	gasLimit := options.GasLimit
	if gasLimit == 0 {
		gasLimit = 65000
	}

	tx := &types.Transaction{
		Chain:        w.chainType,
		From:         fromAddress,
		To:           tokenAddress,
		Value:        big.NewInt(0),
		Data:         data,
		Nonce:        options.Nonce,
		GasPrice:     gasPrice,
		GasLimit:     gasLimit,
		Type:         types.TransactionTypeERC20,
		TokenAddress: tokenAddress,
	}

	return tx, nil
}

func (w *EVMWallet) SignTransaction(ctx context.Context, tx *types.Transaction) ([]byte, error) {
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to derive address: %w", err)
	}

	if !strings.EqualFold(fromAddress, tx.From) {
		return nil, fmt.Errorf("%w: transaction from address does not match key", types.ErrInvalidAddress)
	}

	toAddress := common.HexToAddress(tx.To)
	ethTx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		Gas:      tx.GasLimit,
		To:       &toAddress,
		Value:    tx.Value,
		Data:     tx.Data,
	})

	return w.signEVMTransaction(ctx, ethTx)
}

func (w *EVMWallet) signEVMTransaction(ctx context.Context, tx *ethtypes.Transaction) ([]byte, error) {
	// Create an EIP-155 signer with the chain ID from the wallet config
	signer := ethtypes.NewEIP155Signer(w.config.ChainID)

	// Compute the transaction hash that needs to be signed
	hash := signer.Hash(tx)

	// Sign the hash using the keystore
	signature, err := w.keyStore.Sign(ctx, w.keyID, hash.Bytes(), keystore.DataTypeDigest)
	if err != nil {
		return nil, fmt.Errorf("keystore signing failed: %w", err)
	}

	// Parse the DER-encoded signature into R and S components
	type ecdsaSignature struct {
		R, S *big.Int
	}
	var sigStruct ecdsaSignature
	if _, err := asn1.Unmarshal(signature, &sigStruct); err != nil {
		return nil, fmt.Errorf("failed to parse DER signature: %w", err)
	}

	// Get the secp256k1 curve order (N) and compute N/2
	N := crypto.S256().Params().N
	halfN := new(big.Int).Rsh(N, 1)

	// Normalize S: if S > N/2, adjust it to N - S
	if sigStruct.S.Cmp(halfN) > 0 {
		sigStruct.S.Sub(N, sigStruct.S)
	}

	// Ensure R and S are 32 bytes long (Ethereum expects 32-byte values)
	rBytes := common.LeftPadBytes(sigStruct.R.Bytes(), 32)
	sBytes := common.LeftPadBytes(sigStruct.S.Bytes(), 32)

	// Retrieve the expected public key from the keystore
	key, err := w.keyStore.GetPublicKey(ctx, w.keyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}
	expectedPubKey, err := crypto.UnmarshalPubkey(key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal public key: %w", err)
	}

	// Test recovery ID (v = 27 or 28) to find the correct one
	for recID := 0; recID <= 1; recID++ {
		v := byte(recID)
		testSig := append(rBytes, sBytes...)
		testSig = append(testSig, v)

		// Attempt to recover the public key
		recoveredPubKeyBytes, err := crypto.Ecrecover(hash.Bytes(), testSig)
		if err != nil {
			continue
		}
		recoveredPubKey, err := crypto.UnmarshalPubkey(recoveredPubKeyBytes)
		if err != nil {
			continue
		}

		// Check if the recovered public key matches the expected one
		if recoveredPubKey.X.Cmp(expectedPubKey.X) == 0 && recoveredPubKey.Y.Cmp(expectedPubKey.Y) == 0 {
			// Adjust v for EIP-155: v = 35 + 2*chainID + recID
			vAdjusted := new(big.Int).Mul(w.config.ChainID, big.NewInt(2))
			vAdjusted.Add(vAdjusted, big.NewInt(35+int64(recID)))
			finalSig := append(rBytes, sBytes...)
			finalSig = append(finalSig, byte(vAdjusted.Uint64()))

			// Apply the signature to the transaction
			signedTx, err := tx.WithSignature(signer, finalSig)
			if err != nil {
				return nil, fmt.Errorf("failed to apply signature: %w", err)
			}

			// Serialize the signed transaction
			txBytes, err := signedTx.MarshalBinary()
			if err != nil {
				return nil, fmt.Errorf("failed to encode transaction: %w", err)
			}
			return txBytes, nil
		}
	}

	return nil, fmt.Errorf("failed to recover correct public key from signature")
}
