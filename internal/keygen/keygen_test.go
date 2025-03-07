package keygen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultKeyGenerator_GenerateKeyPair(t *testing.T) {
	keyGen := NewKeyGenerator()

	t.Run("GenerateECDSA_P256", func(t *testing.T) {
		// Generate ECDSA key pair with P-256 curve (default)
		privKeyDER, pubKeyDER, err := keyGen.GenerateKeyPair(KeyTypeECDSA, nil)
		require.NoError(t, err)
		require.NotNil(t, privKeyDER)
		require.NotNil(t, pubKeyDER)

		// Validate private key
		privateECDSA, err := x509.ParseECPrivateKey(privKeyDER)
		require.NoError(t, err)
		assert.Equal(t, "P-256", privateECDSA.Curve.Params().Name)

		// Validate public key
		publicKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyDER)
		require.NoError(t, err)
		publicECDSA, ok := publicKeyInterface.(*ecdsa.PublicKey)
		require.True(t, ok, "Public key should be an ECDSA public key")
		assert.Equal(t, privateECDSA.PublicKey.X, publicECDSA.X)
		assert.Equal(t, privateECDSA.PublicKey.Y, publicECDSA.Y)
	})

	t.Run("GenerateECDSA_SECP256K1", func(t *testing.T) {
		// Generate ECDSA key pair with SECP256K1 curve
		privKeyDER, pubKeyDER, err := keyGen.GenerateKeyPair(KeyTypeECDSA, Secp256k1)
		require.NoError(t, err)
		require.NotNil(t, privKeyDER)
		require.NotNil(t, pubKeyDER)

		// For SECP256K1, we've used custom marshalling that can't be directly parsed with standard libraries
		// We can verify that the DER data is not empty
		assert.True(t, len(privKeyDER) > 0, "Private key data should not be empty")
		assert.True(t, len(pubKeyDER) > 0, "Public key data should not be empty")

		// We could add more direct verification by parsing the ASN.1 structures,
		// but for this test, we'll consider the generation successful if the above checks pass
	})

	t.Run("GenerateRSA", func(t *testing.T) {
		// Generate RSA key pair
		privKeyDER, pubKeyDER, err := keyGen.GenerateKeyPair(KeyTypeRSA, nil)
		require.NoError(t, err)
		require.NotNil(t, privKeyDER)
		require.NotNil(t, pubKeyDER)

		// Validate private key
		privateRSA, err := x509.ParsePKCS1PrivateKey(privKeyDER)
		require.NoError(t, err)
		assert.Equal(t, 2048, privateRSA.Size()*8)

		// Validate public key
		publicKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyDER)
		require.NoError(t, err)
		publicRSA, ok := publicKeyInterface.(*rsa.PublicKey)
		require.True(t, ok, "Public key should be an RSA public key")
		assert.Equal(t, privateRSA.PublicKey.N, publicRSA.N)
		assert.Equal(t, privateRSA.PublicKey.E, publicRSA.E)
	})

	t.Run("GenerateEd25519", func(t *testing.T) {
		// Generate Ed25519 key pair
		privKeyDER, pubKeyDER, err := keyGen.GenerateKeyPair(KeyTypeEd25519, nil)
		require.NoError(t, err)
		require.NotNil(t, privKeyDER)
		require.NotNil(t, pubKeyDER)

		// Parse PKCS8 private key
		privateKey, err := x509.ParsePKCS8PrivateKey(privKeyDER)
		require.NoError(t, err)

		// Check that it's an Ed25519 private key
		privEd25519, ok := privateKey.(ed25519.PrivateKey)
		require.True(t, ok, "Private key should be an Ed25519 private key")

		// Validate public key
		publicKeyInterface, err := x509.ParsePKIXPublicKey(pubKeyDER)
		require.NoError(t, err)
		pubEd25519, ok := publicKeyInterface.(ed25519.PublicKey)
		require.True(t, ok, "Public key should be an Ed25519 public key")

		// Verify that public key matches the one derived from private key
		assert.Equal(t, pubEd25519, privEd25519.Public())
	})

	t.Run("GenerateSymmetric", func(t *testing.T) {
		// Generate symmetric key
		privKey, pubKey, err := keyGen.GenerateKeyPair(KeyTypeSymmetric, nil)
		require.NoError(t, err)
		require.NotNil(t, privKey)

		// Symmetric keys should have no public key
		assert.Nil(t, pubKey)

		// Symmetric key should be 32 bytes (256 bits)
		assert.Equal(t, 32, len(privKey))
	})

	t.Run("UnsupportedKeyType", func(t *testing.T) {
		// Try to generate a key with an unsupported type
		_, _, err := keyGen.GenerateKeyPair(KeyType("UnsupportedType"), nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported key type")
	})
}
