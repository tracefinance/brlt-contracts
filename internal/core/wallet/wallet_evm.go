package wallet

import (
	"context"
	"encoding/asn1"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	coreAbi "vault0/internal/core/abi"
	"vault0/internal/core/keystore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

type EVMWallet struct {
	chain     types.Chain
	keyID     string
	keyStore  keystore.KeyStore
	log       logger.Logger
	abiUtils  coreAbi.ABIUtils
	abiLoader coreAbi.ABILoader
}

func NewEVMWallet(
	keyID string,
	chain types.Chain,
	keyStore keystore.KeyStore,
	abiUtils coreAbi.ABIUtils,
	abiLoader coreAbi.ABILoader,
	log logger.Logger,
) (*EVMWallet, error) {
	return &EVMWallet{
		chain:     chain,
		keyID:     keyID,
		keyStore:  keyStore,
		log:       log,
		abiUtils:  abiUtils,
		abiLoader: abiLoader,
	}, nil
}

func (w *EVMWallet) Chain() types.Chain {
	return w.chain
}

func (w *EVMWallet) DeriveAddress(ctx context.Context) (string, error) {
	key, err := w.keyStore.GetPublicKey(ctx, w.keyID)
	if err != nil {
		return "", err // Don't wrap keystore errors
	}

	publicKey := key.PublicKey
	if len(publicKey) == 0 {
		return "", errors.NewInvalidKeyError("empty public key", nil)
	}

	pubKey, err := crypto.UnmarshalPubkey(publicKey)
	if err != nil {
		return "", errors.NewInvalidKeyError("failed to unmarshal public key", err)
	}

	address := crypto.PubkeyToAddress(*pubKey)
	return address.Hex(), nil
}

func (w *EVMWallet) CreateNativeTransaction(ctx context.Context, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, err // Don't wrap errors from DeriveAddress
	}

	// Allow zero address for contract creation, otherwise validate address
	if toAddress != types.ZeroAddress && !common.IsHexAddress(toAddress) {
		return nil, errors.NewInvalidAddressError(toAddress)
	}

	// For contract deployment (zero address), allow zero amount
	if amount == nil || (toAddress != types.ZeroAddress && amount.Cmp(big.NewInt(0)) <= 0) {
		return nil, errors.NewInvalidAmountError(amount.String())
	}

	gasPrice := options.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice = big.NewInt(int64(w.chain.DefaultGasPrice))
	}

	gasLimit := options.GasLimit
	if gasLimit == 0 {
		gasLimit = w.chain.DefaultGasLimit
	}

	tx := &types.Transaction{
		BaseTransaction: types.BaseTransaction{
			ChainType: w.chain.Type,
			From:      fromAddress,
			To:        toAddress,
			Value:     amount,
			Data:      options.Data,
			Nonce:     options.Nonce,
			GasPrice:  gasPrice,
			GasLimit:  gasLimit,
			Type:      types.TransactionTypeNative,
		},
	}

	return tx, nil
}

func (w *EVMWallet) CreateTokenTransaction(ctx context.Context, tokenAddress, toAddress string, amount *big.Int, options types.TransactionOptions) (*types.Transaction, error) {
	// Derive wallet address
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, err
	}
	// Validate toAddress and tokenAddress
	_, err = types.NewAddress(w.chain.Type, toAddress)
	if err != nil {
		return nil, err
	}
	_, err = types.NewAddress(w.chain.Type, tokenAddress)
	if err != nil {
		return nil, err
	}
	// Validate amount
	if amount == nil || amount.Cmp(big.NewInt(0)) <= 0 {
		return nil, errors.NewInvalidAmountError(amount.String())
	}

	// Get ERC20 transfer ABI by loading standard ERC20 ABI
	erc20ABI, err := w.abiLoader.LoadABIByType(ctx, coreAbi.ABITypeERC20)
	if err != nil {
		return nil, err
	}

	// Pack transfer method data using abiUtils
	data, err := w.abiUtils.Pack(erc20ABI, string(types.ERC20TransferMethod), common.HexToAddress(toAddress), amount)
	if err != nil {
		return nil, err
	}

	gasPrice := options.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice = big.NewInt(int64(w.chain.DefaultGasPrice))
	}

	gasLimit := options.GasLimit
	if gasLimit == 0 {
		gasLimit = 65000
	}

	tx := &types.Transaction{
		BaseTransaction: types.BaseTransaction{
			ChainType: w.chain.Type,
			From:      fromAddress,
			To:        tokenAddress,
			Value:     big.NewInt(0),
			Data:      data,
			Nonce:     options.Nonce,
			GasPrice:  gasPrice,
			GasLimit:  gasLimit,
			Type:      types.TransactionTypeContractCall,
		},
	}

	return tx, nil
}

// SignTransaction signs a transaction with the wallet's key
func (w *EVMWallet) SignTransaction(ctx context.Context, tx *types.Transaction) ([]byte, error) {
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, err // Don't wrap errors from DeriveAddress
	}

	if !strings.EqualFold(fromAddress, tx.From) {
		return nil, errors.NewAddressMismatchError(fromAddress, tx.From)
	}

	toAddress := common.HexToAddress(tx.To)
	ethTx := ethTypes.NewTx(&ethTypes.LegacyTx{
		Nonce:    tx.Nonce,
		GasPrice: tx.GasPrice,
		Gas:      tx.GasLimit,
		To:       &toAddress,
		Value:    tx.Value,
		Data:     tx.Data,
	})

	return w.signEVMTransaction(ctx, ethTx)
}

func (w *EVMWallet) signEVMTransaction(ctx context.Context, tx *ethTypes.Transaction) ([]byte, error) {
	// Create an EIP-155 signer with the chain ID from the wallet config
	signer := ethTypes.NewEIP155Signer(big.NewInt(w.chain.ID))

	// Compute the transaction hash that needs to be signed
	hash := signer.Hash(tx)

	// Sign the hash using the keystore
	signature, err := w.keyStore.Sign(ctx, w.keyID, hash.Bytes(), keystore.DataTypeDigest)
	if err != nil {
		return nil, err // Don't wrap keystore errors
	}

	// Parse the DER-encoded signature into R and S components
	type ecdsaSignature struct {
		R, S *big.Int
	}
	var sigStruct ecdsaSignature
	if _, err := asn1.Unmarshal(signature, &sigStruct); err != nil {
		return nil, errors.NewInvalidSignatureError(err)
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
		return nil, err // Don't wrap keystore errors
	}
	publicKey, err := crypto.UnmarshalPubkey(key.PublicKey)
	if err != nil {
		return nil, errors.NewInvalidKeyError("failed to unmarshal public key", err)
	}

	// Test recovery ID {0, 1} to find the correct v
	for recoveryID := 0; recoveryID <= 1; recoveryID++ {
		testSig := append(rBytes, sBytes...)
		testSig = append(testSig, byte(recoveryID))

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
		if recoveredPubKey.Equal(publicKey) {
			// Found the correct v value, now create the final signature in the format expected by go-ethereum
			signature := make([]byte, 65)
			copy(signature[:32], rBytes)
			copy(signature[32:64], sBytes)
			signature[64] = byte(recoveryID)

			// Create the signed transaction with the signature
			signedTx, err := tx.WithSignature(signer, signature)
			if err != nil {
				return nil, errors.NewSignatureRecoveryError(err)
			}

			// Return the RLP-encoded transaction, ready for broadcasting
			return signedTx.MarshalBinary()
		}
	}

	return nil, errors.NewSignatureRecoveryError(nil)
}

// CreateContractCallTransaction implements the Wallet interface method.
func (w *EVMWallet) CreateContractCallTransaction(ctx context.Context, contractAddress string, value *big.Int, abiString string, method string, args []any, options types.TransactionOptions) (*types.Transaction, error) {
	fromAddress, err := w.DeriveAddress(ctx)
	if err != nil {
		return nil, err // Don't wrap errors from DeriveAddress
	}

	if !common.IsHexAddress(contractAddress) {
		return nil, errors.NewInvalidAddressError(contractAddress)
	}

	// Validate the value
	if value == nil {
		value = big.NewInt(0) // Default to 0 if nil
	} else if value.Cmp(big.NewInt(0)) < 0 {
		return nil, errors.NewInvalidAmountError("value cannot be negative")
	}

	// Use ABIUtils to pack the method call data
	data, err := w.abiUtils.Pack(abiString, method, args...)
	if err != nil {
		// Error handling is already done in the ABIUtils implementation
		return nil, err
	}

	// Set gas price and limit, using defaults if not provided
	gasPrice := options.GasPrice
	if gasPrice == nil || gasPrice.Cmp(big.NewInt(0)) == 0 {
		gasPrice = big.NewInt(int64(w.chain.DefaultGasPrice))
	}

	gasLimit := options.GasLimit
	if gasLimit == 0 {
		// Use a reasonable default for contract calls, perhaps higher than native
		gasLimit = w.chain.DefaultGasLimit * 2 // Example: double the native default
		if gasLimit == 0 {                     // Ensure it's not zero if default is zero
			gasLimit = 100000 // Fallback default
		}
	}

	// Create the transaction struct
	tx := &types.Transaction{
		BaseTransaction: types.BaseTransaction{
			ChainType: w.chain.Type,
			From:      fromAddress,
			To:        contractAddress,
			Value:     value,
			Data:      data,
			Nonce:     options.Nonce,
			GasPrice:  gasPrice,
			GasLimit:  gasLimit,
			Type:      types.TransactionTypeContractCall,
		},
		// Execution fields are zero/nil here as the transaction is not yet executed
	}

	return tx, nil
}
