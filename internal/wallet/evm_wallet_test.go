package wallet

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"

	"vault0/internal/keygen"
	"vault0/internal/keystore"
	"vault0/internal/types"
)

// testKeyStore is a mock keystore for testing signEVMTransaction
type testKeyStore struct {
	privateKey *ecdsa.PrivateKey
	publicKey  *ecdsa.PublicKey
}

// Create implements the keystore.Create method
func (ks *testKeyStore) Create(ctx context.Context, id, name string, keyType keygen.KeyType, tags map[string]string) (*keystore.Key, error) {
	return nil, fmt.Errorf("not implemented")
}

// Import implements the keystore.Import method
func (ks *testKeyStore) Import(ctx context.Context, id, name string, keyType keygen.KeyType, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	return nil, fmt.Errorf("not implemented")
}

// Sign implements the keystore.Sign method, signing the digest with the private key
func (ks *testKeyStore) Sign(ctx context.Context, id string, data []byte, dataType keystore.DataType) ([]byte, error) {
	if id != "test-key" {
		return nil, fmt.Errorf("invalid key ID: %s", id)
	}
	if dataType != keystore.DataTypeDigest {
		return nil, fmt.Errorf("expected digest data type, got %s", dataType)
	}
	r, s, err := ecdsa.Sign(rand.Reader, ks.privateKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to sign data: %w", err)
	}
	return asn1.Marshal(struct{ R, S *big.Int }{r, s})
}

// GetPublicKey implements the keystore.GetPublicKey method
func (ks *testKeyStore) GetPublicKey(ctx context.Context, id string) (*keystore.Key, error) {
	if id != "test-key" {
		return nil, fmt.Errorf("invalid key ID: %s", id)
	}
	pubKeyBytes, err := keygen.MarshalPublicKey(ks.publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}
	return &keystore.Key{
		ID:        id,
		PublicKey: pubKeyBytes,
	}, nil
}

// List implements the keystore.List method
func (ks *testKeyStore) List(ctx context.Context) ([]*keystore.Key, error) {
	return nil, fmt.Errorf("not implemented")
}

// Update implements the keystore.Update method
func (ks *testKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error) {
	return nil, fmt.Errorf("not implemented")
}

// Delete implements the keystore.Delete method
func (ks *testKeyStore) Delete(ctx context.Context, id string) error {
	return fmt.Errorf("not implemented")
}

func TestSignEVMTransaction(t *testing.T) {
	// Generate a private key using the secp256k1 curve
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	assert.NoError(t, err, "Failed to generate private key")

	// Derive the public key and Ethereum address
	publicKey := privateKey.Public().(*ecdsa.PublicKey)
	expectedAddress := crypto.PubkeyToAddress(*publicKey)

	// Create a mock keystore
	testKS := &testKeyStore{
		privateKey: privateKey,
		publicKey:  publicKey,
	}

	// Configure EVMWallet with chain ID 1 (Ethereum mainnet)
	evmConfig := &EVMConfig{
		ChainID: big.NewInt(1),
	}

	// Initialize EVMWallet
	wallet := &EVMWallet{
		keyStore:  testKS,
		chainType: types.ChainTypeEthereum,
		config:    evmConfig,
		keyID:     "test-key",
	}

	// Create a sample legacy transaction
	toAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc454e4438f44e")
	tx := ethtypes.NewTx(&ethtypes.LegacyTx{
		Nonce:    0,
		GasPrice: big.NewInt(20_000_000_000), // 20 Gwei
		Gas:      21_000,
		To:       &toAddress,
		Value:    big.NewInt(1_000_000_000_000_000_000), // 1 ETH
		Data:     nil,
	})

	// Sign the transaction
	signedTxBytes, err := wallet.signEVMTransaction(context.Background(), tx)
	assert.NoError(t, err, "Failed to sign transaction")
	assert.NotNil(t, signedTxBytes, "Signed transaction bytes should not be nil")

	// Decode the signed transaction
	var signedTx ethtypes.Transaction
	err = signedTx.UnmarshalBinary(signedTxBytes)
	assert.NoError(t, err, "Failed to decode signed transaction")

	// Create an EIP-155 signer with the same chain ID
	signer := ethtypes.NewEIP155Signer(evmConfig.ChainID)

	// Recover the sender address from the signed transaction
	sender, err := signer.Sender(&signedTx)
	assert.NoError(t, err, "Failed to recover sender address")
	assert.Equal(t, expectedAddress, sender, "Recovered sender address does not match expected address")

	// Verify that transaction fields are preserved
	assert.Equal(t, tx.Nonce(), signedTx.Nonce(), "Nonce mismatch")
	assert.Equal(t, tx.GasPrice(), signedTx.GasPrice(), "Gas price mismatch")
	assert.Equal(t, tx.Gas(), signedTx.Gas(), "Gas limit mismatch")
	assert.Equal(t, tx.To(), signedTx.To(), "Recipient address mismatch")
	assert.Equal(t, tx.Value(), signedTx.Value(), "Value mismatch")
	assert.Equal(t, tx.Data(), signedTx.Data(), "Data mismatch")
}
