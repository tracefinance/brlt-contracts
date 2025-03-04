package keystore

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultKeyGenerator_GenerateKeyPair(t *testing.T) {
	keyGen := NewKeyGenerator()

	t.Run("GenerateECDSA", func(t *testing.T) {
		// Generate ECDSA key pair
		privKeyPEM, pubKeyPEM, err := keyGen.GenerateKeyPair(KeyTypeECDSA)
		require.NoError(t, err)
		require.NotNil(t, privKeyPEM)
		require.NotNil(t, pubKeyPEM)

		// Decode private PEM to DER
		privBlock, _ := pem.Decode(privKeyPEM)
		require.NotNil(t, privBlock)
		assert.Equal(t, "EC PRIVATE KEY", privBlock.Type)

		// Decode public PEM to DER
		pubBlock, _ := pem.Decode(pubKeyPEM)
		require.NotNil(t, pubBlock)
		assert.Equal(t, "PUBLIC KEY", pubBlock.Type)

		// Validate private key
		privateECDSA, err := x509.ParseECPrivateKey(privBlock.Bytes)
		require.NoError(t, err)
		assert.Equal(t, "P-256", privateECDSA.Curve.Params().Name)

		// Validate public key
		publicKeyInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
		require.NoError(t, err)
		publicECDSA, ok := publicKeyInterface.(*ecdsa.PublicKey)
		require.True(t, ok, "Public key should be an ECDSA public key")
		assert.Equal(t, privateECDSA.PublicKey.X, publicECDSA.X)
		assert.Equal(t, privateECDSA.PublicKey.Y, publicECDSA.Y)
	})

	t.Run("GenerateRSA", func(t *testing.T) {
		// Generate RSA key pair
		privKeyPEM, pubKeyPEM, err := keyGen.GenerateKeyPair(KeyTypeRSA)
		require.NoError(t, err)
		require.NotNil(t, privKeyPEM)
		require.NotNil(t, pubKeyPEM)

		// Decode private PEM to DER
		privBlock, _ := pem.Decode(privKeyPEM)
		require.NotNil(t, privBlock)
		assert.Equal(t, "RSA PRIVATE KEY", privBlock.Type)

		// Decode public PEM to DER
		pubBlock, _ := pem.Decode(pubKeyPEM)
		require.NotNil(t, pubBlock)
		assert.Equal(t, "PUBLIC KEY", pubBlock.Type)

		// Validate private key
		privateRSA, err := x509.ParsePKCS1PrivateKey(privBlock.Bytes)
		require.NoError(t, err)
		assert.Equal(t, 2048, privateRSA.Size()*8)

		// Validate public key
		publicKeyInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
		require.NoError(t, err)
		publicRSA, ok := publicKeyInterface.(*rsa.PublicKey)
		require.True(t, ok, "Public key should be an RSA public key")
		assert.Equal(t, privateRSA.PublicKey.N, publicRSA.N)
		assert.Equal(t, privateRSA.PublicKey.E, publicRSA.E)
	})

	t.Run("GenerateEd25519", func(t *testing.T) {
		// Generate Ed25519 key pair
		privKeyPEM, pubKeyPEM, err := keyGen.GenerateKeyPair(KeyTypeEd25519)
		require.NoError(t, err)
		require.NotNil(t, privKeyPEM)
		require.NotNil(t, pubKeyPEM)

		// Decode private PEM to DER
		privBlock, _ := pem.Decode(privKeyPEM)
		require.NotNil(t, privBlock)
		assert.Equal(t, "PRIVATE KEY", privBlock.Type)

		// Decode public PEM to DER
		pubBlock, _ := pem.Decode(pubKeyPEM)
		require.NotNil(t, pubBlock)
		assert.Equal(t, "PUBLIC KEY", pubBlock.Type)

		// Parse PKCS8 private key
		privateKey, err := x509.ParsePKCS8PrivateKey(privBlock.Bytes)
		require.NoError(t, err)

		// Check that it's an Ed25519 private key
		privEd25519, ok := privateKey.(ed25519.PrivateKey)
		require.True(t, ok, "Private key should be an Ed25519 private key")

		// Validate public key
		publicKeyInterface, err := x509.ParsePKIXPublicKey(pubBlock.Bytes)
		require.NoError(t, err)
		pubEd25519, ok := publicKeyInterface.(ed25519.PublicKey)
		require.True(t, ok, "Public key should be an Ed25519 public key")

		// Verify that public key matches the one derived from private key
		assert.Equal(t, pubEd25519, privEd25519.Public())
	})

	t.Run("GenerateSymmetric", func(t *testing.T) {
		// Generate symmetric key
		privKey, pubKey, err := keyGen.GenerateKeyPair(KeyTypeSymmetric)
		require.NoError(t, err)
		require.NotNil(t, privKey)

		// Symmetric keys should have no public key
		assert.Nil(t, pubKey)

		// Symmetric key should be 32 bytes (256 bits)
		assert.Equal(t, 32, len(privKey))
	})

	t.Run("UnsupportedKeyType", func(t *testing.T) {
		// Try to generate a key with an unsupported type
		_, _, err := keyGen.GenerateKeyPair("UnsupportedType")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported key type")
	})
}
