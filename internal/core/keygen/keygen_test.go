package keygen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"vault0/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewKeyGenerator(t *testing.T) {
	kg := NewKeyGenerator()
	assert.NotNil(t, kg, "KeyGenerator should not be nil")
}

func TestGenerateKeyPair_ECDSA(t *testing.T) {
	kg := NewKeyGenerator()

	tests := []struct {
		name  string
		curve elliptic.Curve
	}{
		{"P256", elliptic.P256()},
		{"P384", elliptic.P384()},
		{"P521", elliptic.P521()},
		{"Default (nil) curve", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privKey, pubKey, err := kg.GenerateKeyPair(types.KeyTypeECDSA, tt.curve)
			require.NoError(t, err)
			require.NotNil(t, privKey)
			require.NotNil(t, pubKey)

			// Parse private key
			parsedPrivKey, err := x509.ParseECPrivateKey(privKey)
			require.NoError(t, err)
			assert.NotNil(t, parsedPrivKey)

			// Parse public key
			parsedPubKeyInterface, err := x509.ParsePKIXPublicKey(pubKey)
			require.NoError(t, err)
			parsedPubKey, ok := parsedPubKeyInterface.(*ecdsa.PublicKey)
			require.True(t, ok)
			assert.NotNil(t, parsedPubKey)

			// Verify that the public key matches the private key
			assert.Equal(t, parsedPrivKey.PublicKey, *parsedPubKey)
		})
	}
}

func TestGenerateKeyPair_RSA(t *testing.T) {
	kg := NewKeyGenerator()

	privKey, pubKey, err := kg.GenerateKeyPair(types.KeyTypeRSA, nil)
	require.NoError(t, err)
	require.NotNil(t, privKey)
	require.NotNil(t, pubKey)

	// Parse private key
	parsedPrivKey, err := x509.ParsePKCS1PrivateKey(privKey)
	require.NoError(t, err)
	assert.NotNil(t, parsedPrivKey)
	assert.Equal(t, 2048, parsedPrivKey.Size()*8) // Verify key size

	// Parse public key
	parsedPubKeyInterface, err := x509.ParsePKIXPublicKey(pubKey)
	require.NoError(t, err)
	parsedPubKey, ok := parsedPubKeyInterface.(*rsa.PublicKey)
	require.True(t, ok)
	assert.NotNil(t, parsedPubKey)

	// Verify that the public key matches the private key
	assert.Equal(t, parsedPrivKey.PublicKey, *parsedPubKey)
}

func TestGenerateKeyPair_Ed25519(t *testing.T) {
	kg := NewKeyGenerator()

	privKey, pubKey, err := kg.GenerateKeyPair(types.KeyTypeEd25519, nil)
	require.NoError(t, err)
	require.NotNil(t, privKey)
	require.NotNil(t, pubKey)

	// Parse private key
	parsedPrivKeyInterface, err := x509.ParsePKCS8PrivateKey(privKey)
	require.NoError(t, err)
	parsedPrivKey, ok := parsedPrivKeyInterface.(ed25519.PrivateKey)
	require.True(t, ok)
	assert.NotNil(t, parsedPrivKey)

	// Parse public key
	parsedPubKeyInterface, err := x509.ParsePKIXPublicKey(pubKey)
	require.NoError(t, err)
	parsedPubKey, ok := parsedPubKeyInterface.(ed25519.PublicKey)
	require.True(t, ok)
	assert.NotNil(t, parsedPubKey)

	// Verify that the public key matches the private key's public key
	assert.Equal(t, parsedPrivKey.Public(), parsedPubKey)
}

func TestGenerateKeyPair_Symmetric(t *testing.T) {
	kg := NewKeyGenerator()

	privKey, pubKey, err := kg.GenerateKeyPair(types.KeyTypeSymmetric, nil)
	require.NoError(t, err)
	require.NotNil(t, privKey)
	require.Nil(t, pubKey) // Symmetric keys don't have a public key

	// Verify key length
	assert.Equal(t, 32, len(privKey))
}

func TestGenerateKeyPair_InvalidType(t *testing.T) {
	kg := NewKeyGenerator()

	privKey, pubKey, err := kg.GenerateKeyPair("invalid", nil)
	assert.Error(t, err)
	assert.ErrorContains(t, err, "Invalid key type")
	assert.Nil(t, privKey)
	assert.Nil(t, pubKey)
}

func TestDefaultKeyGenerator_GenerateKeyPair(t *testing.T) {
	kg := NewKeyGenerator()

	tests := []struct {
		name      string
		keyType   types.KeyType
		curve     elliptic.Curve
		wantErr   bool
		validator func(t *testing.T, privKey, pubKey []byte)
	}{
		{
			name:    "ECDSA with P256 curve",
			keyType: types.KeyTypeECDSA,
			curve:   elliptic.P256(),
			validator: func(t *testing.T, privKey, pubKey []byte) {
				// Parse private key
				privateKey, err := x509.ParseECPrivateKey(privKey)
				require.NoError(t, err)
				assert.Equal(t, elliptic.P256(), privateKey.Curve)

				// Parse public key
				publicKeyIface, err := x509.ParsePKIXPublicKey(pubKey)
				require.NoError(t, err)
				publicKey, ok := publicKeyIface.(*ecdsa.PublicKey)
				require.True(t, ok)
				assert.Equal(t, elliptic.P256(), publicKey.Curve)
			},
		},
		{
			name:    "ECDSA with P384 curve",
			keyType: types.KeyTypeECDSA,
			curve:   elliptic.P384(),
			validator: func(t *testing.T, privKey, pubKey []byte) {
				// Parse private key
				privateKey, err := x509.ParseECPrivateKey(privKey)
				require.NoError(t, err)
				assert.Equal(t, elliptic.P384(), privateKey.Curve)

				// Parse public key
				publicKeyIface, err := x509.ParsePKIXPublicKey(pubKey)
				require.NoError(t, err)
				publicKey, ok := publicKeyIface.(*ecdsa.PublicKey)
				require.True(t, ok)
				assert.Equal(t, elliptic.P384(), publicKey.Curve)
			},
		},
		{
			name:    "RSA key pair",
			keyType: types.KeyTypeRSA,
			validator: func(t *testing.T, privKey, pubKey []byte) {
				// Parse private key
				privateKey, err := x509.ParsePKCS1PrivateKey(privKey)
				require.NoError(t, err)
				assert.Equal(t, 2048, privateKey.Size()*8) // Check key size is 2048 bits

				// Parse public key
				publicKeyIface, err := x509.ParsePKIXPublicKey(pubKey)
				require.NoError(t, err)
				publicKey, ok := publicKeyIface.(*rsa.PublicKey)
				require.True(t, ok)
				assert.Equal(t, 2048, publicKey.Size()*8)
			},
		},
		{
			name:    "Ed25519 key pair",
			keyType: types.KeyTypeEd25519,
			validator: func(t *testing.T, privKey, pubKey []byte) {
				// Parse private key
				privateKeyIface, err := x509.ParsePKCS8PrivateKey(privKey)
				require.NoError(t, err)
				_, ok := privateKeyIface.(ed25519.PrivateKey)
				require.True(t, ok)

				// Parse public key
				publicKeyIface, err := x509.ParsePKIXPublicKey(pubKey)
				require.NoError(t, err)
				_, ok = publicKeyIface.(ed25519.PublicKey)
				require.True(t, ok)
			},
		},
		{
			name:    "Symmetric key",
			keyType: types.KeyTypeSymmetric,
			validator: func(t *testing.T, privKey, pubKey []byte) {
				assert.Len(t, privKey, 32) // Check symmetric key length is 32 bytes (256 bits)
				assert.Nil(t, pubKey)      // Symmetric keys don't have a public key
			},
		},
		{
			name:    "Invalid key type",
			keyType: "invalid",
			wantErr: true,
			validator: func(t *testing.T, privKey, pubKey []byte) {
				assert.Nil(t, privKey)
				assert.Nil(t, pubKey)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			privKey, pubKey, err := kg.GenerateKeyPair(tt.keyType, tt.curve)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "Invalid key type")
				return
			}

			require.NoError(t, err)
			tt.validator(t, privKey, pubKey)
		})
	}
}

func TestDefaultKeyGenerator_GenerateKeyPair_MultipleCallsConsistency(t *testing.T) {
	kg := NewKeyGenerator()

	// Test that multiple calls with the same parameters generate different keys
	keyTypes := []types.KeyType{types.KeyTypeECDSA, types.KeyTypeRSA, types.KeyTypeEd25519, types.KeyTypeSymmetric}

	for _, keyType := range keyTypes {
		t.Run(string(keyType), func(t *testing.T) {
			// Generate first pair
			priv1, pub1, err := kg.GenerateKeyPair(keyType, nil)
			require.NoError(t, err)

			// Generate second pair
			priv2, pub2, err := kg.GenerateKeyPair(keyType, nil)
			require.NoError(t, err)

			// Keys should be different
			assert.NotEqual(t, priv1, priv2, "private keys should be different")
			if keyType != types.KeyTypeSymmetric {
				assert.NotEqual(t, pub1, pub2, "public keys should be different")
			}
		})
	}
}
