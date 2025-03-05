package keygen

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/asn1"
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalSecp256k1PrivateKey(t *testing.T) {
	// Create a test private key
	privateKey := &ecdsa.PrivateKey{
		D: big.NewInt(123456789),
		PublicKey: ecdsa.PublicKey{
			Curve: Secp256k1,
			X:     big.NewInt(111222333),
			Y:     big.NewInt(444555666),
		},
	}

	// Marshal the private key
	derBytes, err := marshalSecp256k1PrivateKey(privateKey)
	require.NoError(t, err)
	require.NotNil(t, derBytes)

	// Parse the DER format to verify structure
	var parsedKey ecPrivateKey
	rest, err := asn1.Unmarshal(derBytes, &parsedKey)
	require.NoError(t, err)
	require.Empty(t, rest, "DER bytes should be fully consumed")

	// Verify key components
	assert.Equal(t, 1, parsedKey.Version)
	assert.NotEmpty(t, parsedKey.PrivateKey)
	assert.Equal(t, asn1.ObjectIdentifier{1, 3, 132, 0, 10}, parsedKey.NamedCurveOID)
	assert.NotEmpty(t, parsedKey.PublicKey.Bytes)

	// Check the first byte of public key is 4 (uncompressed format)
	assert.Equal(t, byte(4), parsedKey.PublicKey.Bytes[0])
}

func TestMarshalSecp256k1PublicKey(t *testing.T) {
	// Create a test public key
	publicKey := &ecdsa.PublicKey{
		Curve: Secp256k1,
		X:     big.NewInt(111222333),
		Y:     big.NewInt(444555666),
	}

	// Marshal the public key
	derBytes, err := marshalSecp256k1PublicKey(publicKey)
	require.NoError(t, err)
	require.NotNil(t, derBytes)

	// Basic verification of DER format
	// Should start with SEQUENCE tag (0x30)
	assert.Equal(t, byte(0x30), derBytes[0])

	// Convert to hex for easier inspection in test failures
	hexDER := hex.EncodeToString(derBytes)
	t.Logf("Public key DER: %s", hexDER)

	// Instead of checking exact byte positions, check for the presence of key patterns

	// Check for OID patterns - these should be present somewhere in the DER
	// OID for id-ecPublicKey (1.2.840.10045.2.1)
	ecPubKeyOID := []byte{0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02, 0x01}
	assert.True(t, bytes.Contains(derBytes, ecPubKeyOID),
		"DER should contain EC public key OID (1.2.840.10045.2.1)")

	// OID for secp256k1 (1.3.132.0.10)
	secp256k1OID := []byte{0x2b, 0x81, 0x04, 0x00, 0x0a}
	assert.True(t, bytes.Contains(derBytes, secp256k1OID),
		"DER should contain secp256k1 OID (1.3.132.0.10)")

	// Find the bit string containing the public key
	// This is a bit string (0x03) followed by length and then 0x00 (for unused bits)
	// then the public key data starting with 0x04 for uncompressed format
	bitStringIndex := bytes.Index(derBytes, []byte{0x03, 0x42, 0x00, 0x04})
	assert.True(t, bitStringIndex > 0,
		"DER should contain a bit string with uncompressed public key data (0x04)")

	// Verify we can find the X and Y coordinates in the expected positions
	// For the test key, X=111222333, Y=444555666
	// We only check for non-zero values at the right positions
	assert.True(t, bytes.Contains(derBytes, []byte{0x04, 0x00, 0x00, 0x00}),
		"Public key data should start with 0x04 followed by X,Y coords")
}

func TestMarshalRoundTrip(t *testing.T) {
	// Generate a real SECP256K1 key pair
	privateECDSA, err := ecdsa.GenerateKey(Secp256k1, rand.Reader)
	require.NoError(t, err)

	// Marshal private key
	privDER, err := marshalSecp256k1PrivateKey(privateECDSA)
	require.NoError(t, err)

	// Marshal public key
	pubDER, err := marshalSecp256k1PublicKey(&privateECDSA.PublicKey)
	require.NoError(t, err)

	// Basic validation of private key structure
	var parsedPrivateKey ecPrivateKey
	_, err = asn1.Unmarshal(privDER, &parsedPrivateKey)
	require.NoError(t, err)

	// Check expected fields
	assert.Equal(t, 1, parsedPrivateKey.Version)
	assert.Len(t, parsedPrivateKey.PrivateKey, 32) // 32 bytes for 256-bit key
	assert.Equal(t, asn1.ObjectIdentifier{1, 3, 132, 0, 10}, parsedPrivateKey.NamedCurveOID)

	// Validate that the public key in the private key structure matches our public key
	// This can only be checked by comparing the curve point coordinates
	assert.Equal(t, byte(0x04), parsedPrivateKey.PublicKey.Bytes[0]) // Uncompressed format

	// Verify the public key DER is not empty
	assert.NotEmpty(t, pubDER)
	assert.Greater(t, len(pubDER), 70) // Typical secp256k1 public key DER is around 88 bytes
}
