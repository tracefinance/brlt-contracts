package wallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	coreAbi "vault0/internal/core/abi"
	coreCrypto "vault0/internal/core/crypto"
	"vault0/internal/core/keystore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// ecdsaSignature is used for marshalling ECDSA signatures in ASN.1 DER format
type ecdsaSignature struct {
	R, S *big.Int
}

// MockABIUtils implements the ABIUtils interface for testing
type MockABIUtils struct {
	UnpackFunc             func(contractABI string, methodName string, inputData []byte) (map[string]any, error)
	PackFunc               func(contractABI string, methodName string, args ...any) ([]byte, error)
	ExtractMethodIDFunc    func(data []byte) []byte
	GetAddressFromArgsFunc func(args map[string]any, key string) (types.Address, error)
	GetBytes32FromArgsFunc func(args map[string]any, key string) ([32]byte, error)
	GetBigIntFromArgsFunc  func(args map[string]any, key string) (*types.BigInt, error)
	GetUint64FromArgsFunc  func(args map[string]any, key string) (uint64, error)
}

func (m *MockABIUtils) Unpack(contractABI string, methodName string, inputData []byte) (map[string]any, error) {
	if m.UnpackFunc != nil {
		return m.UnpackFunc(contractABI, methodName, inputData)
	}
	return map[string]any{}, nil
}

func (m *MockABIUtils) Pack(contractABI string, methodName string, args ...any) ([]byte, error) {
	if m.PackFunc != nil {
		return m.PackFunc(contractABI, methodName, args...)
	}
	// Default implementation: for ERC20 transfer, create a simple encoded data format
	if methodName == string(types.ERC20TransferMethod) && len(args) == 2 {
		// For testing we just need a plausible result, not an accurate encoding
		addr, ok := args[0].(common.Address)
		if ok {
			// Create method ID + placeholder data with some address bytes
			// Method ID (0xa9059cbb = keccak256("transfer(address,uint256)")[:4])
			result := []byte{0xa9, 0x05, 0x9c, 0xbb}
			// Padding before address
			padding := make([]byte, 12)
			result = append(result, padding...)
			// Append address bytes
			result = append(result, addr.Bytes()...)
			// Padding for amount parameter
			amountPadding := make([]byte, 32)
			// Just set the last byte to 1 for test purposes
			amountPadding[31] = 0x01
			result = append(result, amountPadding...)
			return result, nil
		}
	}
	return []byte{0x12, 0x34, 0x56, 0x78}, nil // Default mocked result
}

func (m *MockABIUtils) ExtractMethodID(data []byte) []byte {
	if m.ExtractMethodIDFunc != nil {
		return m.ExtractMethodIDFunc(data)
	}
	if len(data) >= 4 {
		return data[:4]
	}
	return nil
}

func (m *MockABIUtils) GetAddressFromArgs(args map[string]any, key string) (types.Address, error) {
	if m.GetAddressFromArgsFunc != nil {
		return m.GetAddressFromArgsFunc(args, key)
	}
	return types.Address{}, nil
}

func (m *MockABIUtils) GetBytes32FromArgs(args map[string]any, key string) ([32]byte, error) {
	if m.GetBytes32FromArgsFunc != nil {
		return m.GetBytes32FromArgsFunc(args, key)
	}
	return [32]byte{}, nil
}

func (m *MockABIUtils) GetBigIntFromArgs(args map[string]any, key string) (*types.BigInt, error) {
	if m.GetBigIntFromArgsFunc != nil {
		return m.GetBigIntFromArgsFunc(args, key)
	}
	return &types.BigInt{}, nil
}

func (m *MockABIUtils) GetUint64FromArgs(args map[string]any, key string) (uint64, error) {
	if m.GetUint64FromArgsFunc != nil {
		return m.GetUint64FromArgsFunc(args, key)
	}
	return 0, nil
}

// MockABILoader implements the ABILoader interface for testing
type MockABILoader struct {
	LoadABIByTypeFunc    func(ctx context.Context, abiType coreAbi.SupportedABIType) (string, error)
	LoadABIByAddressFunc func(ctx context.Context, contractAddress types.Address) (string, error)
}

func (m *MockABILoader) LoadABIByType(ctx context.Context, abiType coreAbi.SupportedABIType) (string, error) {
	if m.LoadABIByTypeFunc != nil {
		return m.LoadABIByTypeFunc(ctx, abiType)
	}
	// Default implementation for ERC20
	if abiType == coreAbi.ABITypeERC20 {
		return `[{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},{"constant":false,"inputs":[{"name":"_to","type":"address"},{"name":"_value","type":"uint256"}],"name":"transfer","outputs":[{"name":"success","type":"bool"}],"payable":false,"stateMutability":"nonpayable","type":"function"}]`, nil
	}
	return "{}", nil
}

func (m *MockABILoader) LoadABIByAddress(ctx context.Context, contractAddress types.Address) (string, error) {
	if m.LoadABIByAddressFunc != nil {
		return m.LoadABIByAddressFunc(ctx, contractAddress)
	}
	return "{}", nil
}

// testChain is a test chain configuration
var testChain = types.Chain{
	ID:              1,
	Type:            types.ChainTypeEthereum,
	Name:            "Ethereum",
	Symbol:          "ETH",
	RPCUrl:          "https://mainnet.infura.io",
	ExplorerUrl:     "https://etherscan.io",
	KeyType:         types.KeyTypeECDSA,
	Curve:           coreCrypto.Secp256k1Curve,
	DefaultGasLimit: 21000,
	DefaultGasPrice: 20000000000, // 20 Gwei
}

// setupTest creates a test wallet with mock dependencies
func setupTest(t *testing.T) (*EVMWallet, *MockKeyStore) {
	ks := &MockKeyStore{}
	log := logger.NewNopLogger()
	abiUtils := &MockABIUtils{}
	abiLoader := &MockABILoader{}

	wallet, err := NewEVMWallet("test", testChain, ks, abiUtils, abiLoader, log)
	require.NoError(t, err)
	return wallet, ks
}

// TestChain tests the Chain method
func TestChain(t *testing.T) {
	wallet, _ := setupTest(t)
	chain := wallet.Chain()
	assert.Equal(t, types.ChainTypeEthereum, chain.Type, "Chain.Type should be Ethereum")
	assert.Equal(t, int64(1), chain.ID, "Chain.ID should be 1")
	assert.Equal(t, "Ethereum", chain.Name, "Chain.Name should be Ethereum")
}

// TestDeriveAddress tests the DeriveAddress method
func TestDeriveAddress(t *testing.T) {
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Generate a test key pair using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      types.KeyTypeECDSA,
				Curve:     coreCrypto.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, errors.NewKeyNotFoundError(id)
	}

	// Derive address
	address, err := wallet.DeriveAddress(ctx)
	require.NoError(t, err)
	expectedAddress := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	assert.Equal(t, expectedAddress, address, "Derived address should match expected address")

	t.Run("key not found", func(t *testing.T) {
		// Setup
		mockKeyStore := keystore.NewMockKeyStore()
		log := logger.NewNopLogger()
		abiUtils := &MockABIUtils{}
		abiLoader := &MockABILoader{}
		wallet, err := NewEVMWallet("non-existent-key", testChain, mockKeyStore, abiUtils, abiLoader, log)
		require.NoError(t, err)

		// Execute
		address, err := wallet.DeriveAddress(context.Background())

		// Assert
		assert.Error(t, err)
		assert.Empty(t, address)
		assert.ErrorContains(t, err, "Key not found: non-existent-key")
	})
}

// TestCreateNativeTransaction tests the CreateNativeTransaction method
func TestCreateNativeTransaction(t *testing.T) {
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Setup key using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := coreCrypto.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", types.KeyTypeECDSA, coreCrypto.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      types.KeyTypeECDSA,
				Curve:     coreCrypto.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, errors.NewKeyNotFoundError(id)
	}

	toAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	amount := big.NewInt(1000000000000000000) // 1 ETH

	// Test regular native transaction
	tx, err := wallet.CreateNativeTransaction(ctx, toAddress, amount, types.TransactionOptions{})
	require.NoError(t, err)
	assert.Equal(t, types.ChainTypeEthereum, tx.ChainType, "Transaction chain should be Ethereum")
	assert.Equal(t, crypto.PubkeyToAddress(privKey.PublicKey).Hex(), tx.From, "From address should match wallet address")
	assert.Equal(t, toAddress, tx.To, "To address should match input")
	assert.Equal(t, amount, tx.Value, "Transaction value should match input")
	assert.Equal(t, types.TransactionTypeNative, tx.Type, "Transaction type should be native")

	// Test contract deployment (zero address)
	deployData := []byte{0x60, 0x80, 0x60, 0x40} // Some dummy contract bytecode
	tx, err = wallet.CreateNativeTransaction(ctx, types.ZeroAddress, big.NewInt(0), types.TransactionOptions{
		Data: deployData,
	})
	require.NoError(t, err)
	assert.Equal(t, types.ZeroAddress, tx.To, "To address should be zero address for contract deployment")
	assert.Equal(t, deployData, tx.Data, "Transaction data should contain contract bytecode")
	assert.Equal(t, types.TransactionTypeNative, tx.Type, "Transaction type should be native")

	// Test failure case: invalid toAddress
	_, err = wallet.CreateNativeTransaction(ctx, "invalid-address", amount, types.TransactionOptions{})
	assert.Error(t, err, "CreateNativeTransaction should fail with invalid address")

	t.Run("key not found", func(t *testing.T) {
		// Setup
		mockKeyStore := keystore.NewMockKeyStore()
		log := logger.NewNopLogger()
		abiUtils := &MockABIUtils{}
		abiLoader := &MockABILoader{}
		wallet, err := NewEVMWallet("non-existent-key", testChain, mockKeyStore, abiUtils, abiLoader, log)
		require.NoError(t, err)

		// Execute
		tx, err := wallet.CreateNativeTransaction(context.Background(), "0x742d35Cc6634C0532925a3b844Bc454e4438f44e", big.NewInt(1), types.TransactionOptions{})

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tx)
		assert.ErrorContains(t, err, "Key not found: non-existent-key")
	})
}

// TestCreateTokenTransaction tests the CreateTokenTransaction method
func TestCreateTokenTransaction(t *testing.T) {
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Setup key using secp256k1 curve
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Get public key bytes using secp256k1 format
	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	// Marshal private key using secp256k1 format
	privKeyBytes, err := coreCrypto.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", types.KeyTypeECDSA, coreCrypto.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock to return our key
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		if id == "test" {
			return &keystore.Key{
				ID:        "test",
				Name:      "test",
				Type:      types.KeyTypeECDSA,
				Curve:     coreCrypto.Secp256k1Curve,
				PublicKey: pubKeyBytes,
			}, nil
		}
		return nil, errors.NewKeyNotFoundError(id)
	}

	tokenAddress := "0xdAC17F958D2ee523a2206206994597C13D831ec7" // USDT address
	toAddress := "0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
	amount := big.NewInt(1000000) // 1 USDT with 6 decimals

	tx, err := wallet.CreateTokenTransaction(ctx, tokenAddress, toAddress, amount, types.TransactionOptions{})
	require.NoError(t, err)
	assert.Equal(t, types.ChainTypeEthereum, tx.ChainType, "Transaction chain should be Ethereum")
	assert.Equal(t, crypto.PubkeyToAddress(privKey.PublicKey).Hex(), tx.From, "From address should match wallet address")
	assert.Equal(t, tokenAddress, tx.To, "To address should be token contract address")
	assert.Equal(t, big.NewInt(0), tx.Value, "Value should be 0 for token transactions")
	assert.NotEmpty(t, tx.Data, "Transaction data should contain ERC20 transfer ABI")
	assert.Equal(t, types.TransactionTypeContractCall, tx.Type, "Transaction type should be ContractCall")

	// Test failure case: invalid toAddress
	_, err = wallet.CreateTokenTransaction(ctx, tokenAddress, "invalid-address", amount, types.TransactionOptions{})
	assert.Error(t, err, "CreateTokenTransaction should fail with invalid toAddress")
}

// TestSignTransaction tests the SignTransaction method with DER-encoded keys
func TestSignTransaction(t *testing.T) {
	wallet, ks := setupTest(t)
	ctx := context.Background()

	// Generate a test key pair
	privKey, err := ecdsa.GenerateKey(coreCrypto.Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	pubKeyBytes, err := coreCrypto.MarshalPublicKey(&privKey.PublicKey)
	require.NoError(t, err)

	privKeyBytes, err := coreCrypto.MarshalPrivateKey(privKey)
	require.NoError(t, err)

	// Import the key into the mock keystore
	_, err = ks.Import(ctx, "test", types.KeyTypeECDSA, coreCrypto.Secp256k1Curve, privKeyBytes, pubKeyBytes, nil)
	require.NoError(t, err)

	// Set up the mock
	address := crypto.PubkeyToAddress(privKey.PublicKey).Hex()
	ks.GetPublicKeyFunc = func(ctx context.Context, id string) (*keystore.Key, error) {
		return &keystore.Key{
			ID:        "test",
			Type:      types.KeyTypeECDSA,
			Curve:     coreCrypto.Secp256k1Curve,
			PublicKey: pubKeyBytes,
		}, nil
	}

	// Create a test transaction
	tx := &types.Transaction{
		BaseTransaction: types.BaseTransaction{
			ChainType: types.ChainTypeEthereum,
			From:      address,
			To:        "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
			Value:     big.NewInt(1000000000000000000), // 1 ETH
			GasPrice:  big.NewInt(20000000000),         // 20 Gwei
			GasLimit:  21000,
			Nonce:     0,
			Type:      types.TransactionTypeNative, // Assuming this test case is for a native tx
		},
	}

	// Set up the mock to sign with the private key
	ks.SignFunc = func(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
		if dataType != keystore.DataTypeDigest {
			return nil, fmt.Errorf("unexpected data type: %v", dataType)
		}

		// Sign the data using the private key
		r, s, err := ecdsa.Sign(rand.Reader, privKey, data)
		if err != nil {
			return nil, err
		}

		// Marshal the signature in ASN.1 DER format
		sig := ecdsaSignature{R: r, S: s}
		return asn1.Marshal(sig)
	}

	// Sign the transaction
	signedTx, err := wallet.SignTransaction(ctx, tx)
	require.NoError(t, err)
	assert.NotEmpty(t, signedTx, "Signed transaction should not be empty")

	t.Run("key not found", func(t *testing.T) {
		// Setup
		mockKeyStore := keystore.NewMockKeyStore()
		log := logger.NewNopLogger()
		abiUtils := &MockABIUtils{}
		abiLoader := &MockABILoader{}
		wallet, err := NewEVMWallet("non-existent-key", testChain, mockKeyStore, abiUtils, abiLoader, log)
		require.NoError(t, err)

		tx := &types.Transaction{
			BaseTransaction: types.BaseTransaction{
				ChainType: types.ChainTypeEthereum,
				From:      "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				To:        "0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
				Value:     big.NewInt(1),
				GasPrice:  big.NewInt(1),
				GasLimit:  21000,
				Type:      types.TransactionTypeNative, // Assuming native type for this test
			},
		}

		// Execute
		signature, err := wallet.SignTransaction(context.Background(), tx)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, signature)
		assert.ErrorContains(t, err, "Key not found: non-existent-key")
	})
}

// TestNewEVMWalletValidation tests the validation in NewEVMWallet
func TestNewEVMWalletValidation(t *testing.T) {
	ks := &MockKeyStore{}
	log := logger.NewNopLogger()
	abiUtils := &MockABIUtils{}
	abiLoader := &MockABILoader{}

	// Test with nil keystore
	_, err := NewEVMWallet("test", types.Chain{}, nil, abiUtils, abiLoader, log)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Invalid wallet configuration: keystore cannot be nil")

	// Test with empty keyID
	_, err = NewEVMWallet("", types.Chain{}, ks, abiUtils, abiLoader, log)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Invalid wallet configuration: keyID cannot be empty")

	// Test with nil abiUtils
	_, err = NewEVMWallet("test", types.Chain{}, ks, nil, abiLoader, log)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Invalid wallet configuration: abiUtils cannot be nil")

	// Test with nil abiLoader
	_, err = NewEVMWallet("test", types.Chain{}, ks, abiUtils, nil, log)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Invalid wallet configuration: abiLoader cannot be nil")

	// Test with invalid key type
	_, err = NewEVMWallet("test", types.Chain{
		KeyType: types.KeyTypeRSA,
		Curve:   coreCrypto.Secp256k1Curve,
	}, ks, abiUtils, abiLoader, log)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Invalid key type: expected ecdsa, got rsa")

	// Test with invalid curve
	_, err = NewEVMWallet("test", types.Chain{
		KeyType: types.KeyTypeECDSA,
		Curve:   elliptic.P256(),
	}, ks, abiUtils, abiLoader, log)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Invalid curve: expected secp256k1, got P-256")
}
