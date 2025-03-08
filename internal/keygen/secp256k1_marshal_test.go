package keygen

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestMarshalUnmarshalPublicKey tests the MarshalPublicKey and UnmarshalPublicKey functions.
// It generates a key pair, marshals the public key, and then unmarshals it to verify correctness.
func TestMarshalUnmarshalPublicKey(t *testing.T) {
	// Generate a key pair using secp256k1
	priv, err := ecdsa.GenerateKey(Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Marshal the public key (expecting uncompressed format: 0x04 || X || Y)
	pubBytes, err := MarshalPublicKey(&priv.PublicKey)
	require.NoError(t, err)
	require.Len(t, pubBytes, 65, "Public key should be 65 bytes: 1 prefix + 32 X + 32 Y")
	require.Equal(t, byte(0x04), pubBytes[0], "Public key should start with 0x04 for uncompressed format")

	// Unmarshal the public key and verify it matches the original
	pub, err := UnmarshalPublicKey(pubBytes)
	require.NoError(t, err)
	require.Equal(t, priv.PublicKey.X, pub.X, "X coordinate should match")
	require.Equal(t, priv.PublicKey.Y, pub.Y, "Y coordinate should match")
}

// TestMarshalUnmarshalPrivateKey tests the MarshalPrivateKey and UnmarshalPrivateKey functions.
// It generates a key pair, marshals the private key, and then unmarshals it to verify correctness.
func TestMarshalUnmarshalPrivateKey(t *testing.T) {
	// Generate a key pair using secp256k1
	priv, err := ecdsa.GenerateKey(Secp256k1Curve, rand.Reader)
	require.NoError(t, err)

	// Marshal the private key
	privBytes, err := MarshalPrivateKey(priv)
	require.NoError(t, err)
	require.NotEmpty(t, privBytes, "Private key bytes should not be empty")

	// Unmarshal the private key and verify it matches the original
	priv2, err := UnmarshalPrivateKey(privBytes)
	require.NoError(t, err)
	require.Equal(t, priv.D, priv2.D, "Private key scalar D should match")
	require.Equal(t, priv.PublicKey.X, priv2.PublicKey.X, "Public key X coordinate should match")
	require.Equal(t, priv.PublicKey.Y, priv2.PublicKey.Y, "Public key Y coordinate should match")
}

// TestUnmarshalPublicKeyInvalid tests the UnmarshalPublicKey function with invalid inputs.
func TestUnmarshalPublicKeyInvalid(t *testing.T) {
	// Test with invalid length (too short)
	_, err := UnmarshalPublicKey([]byte{0x04})
	require.Error(t, err, "Unmarshaling a too-short public key should fail")

	// Test with invalid prefix (not 0x04)
	invalidPrefix := append([]byte{0x05}, make([]byte, 64)...)
	_, err = UnmarshalPublicKey(invalidPrefix)
	require.Error(t, err, "Unmarshaling a public key with invalid prefix should fail")
}

// TestUnmarshalPrivateKeyInvalid tests the UnmarshalPrivateKey function with invalid inputs.
func TestUnmarshalPrivateKeyInvalid(t *testing.T) {
	// Test with empty bytes
	_, err := UnmarshalPrivateKey([]byte{})
	require.Error(t, err, "Unmarshaling an empty private key should fail")

	// Test with an invalid scalar (larger than the curve order)
	largeScalar := new(big.Int).Add(Secp256k1Curve.Params().N, big.NewInt(1))
	_, err = UnmarshalPrivateKey(largeScalar.Bytes())
	require.Error(t, err, "Unmarshaling a scalar larger than curve order should fail")
}
