package keygen

import (
	"crypto/ecdsa"
	"encoding/asn1"
)

// ASN.1 structure for EC private key
type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

// marshalSecp256k1PrivateKey serializes a secp256k1 private key to DER format
func marshalSecp256k1PrivateKey(privateKey *ecdsa.PrivateKey) ([]byte, error) {
	// Secp256k1 OID (1.3.132.0.10)
	// This is the standard OID for secp256k1 curve
	oid := asn1.ObjectIdentifier{1, 3, 132, 0, 10}

	// Convert private key to bytes (big-endian)
	privateKeyBytes := privateKey.D.Bytes()
	// Pad to 32 bytes if necessary
	paddedPrivateKey := make([]byte, 32)
	copy(paddedPrivateKey[32-len(privateKeyBytes):], privateKeyBytes)

	// Manually construct public key bytes
	// For secp256k1, we need to manually create the compressed format
	// since we can't use ecdh with a custom curve
	publicKeyBytes := make([]byte, 65)
	publicKeyBytes[0] = 4 // uncompressed point format
	xBytes := privateKey.PublicKey.X.Bytes()
	yBytes := privateKey.PublicKey.Y.Bytes()
	copy(publicKeyBytes[1+32-len(xBytes):33], xBytes)
	copy(publicKeyBytes[33+32-len(yBytes):], yBytes)

	// Create ASN.1 structure
	key := ecPrivateKey{
		Version:       1,
		PrivateKey:    paddedPrivateKey,
		NamedCurveOID: oid,
		PublicKey:     asn1.BitString{Bytes: publicKeyBytes},
	}

	return asn1.Marshal(key)
}

// marshalSecp256k1PublicKey serializes a secp256k1 public key to DER format
func marshalSecp256k1PublicKey(publicKey *ecdsa.PublicKey) ([]byte, error) {
	// Create SubjectPublicKeyInfo structure
	// OID for id-ecPublicKey is 1.2.840.10045.2.1
	algorithmIdentifier := []byte{
		0x30, 0x10, // SEQUENCE (16 bytes)
		0x06, 0x07, 0x2a, 0x86, 0x48, 0xce, 0x3d, 0x02, 0x01, // OID for id-ecPublicKey
		0x06, 0x05, 0x2b, 0x81, 0x04, 0x00, 0x0a, // OID for secp256k1 (1.3.132.0.10)
	}

	// Manually construct the public key bytes
	// We can't use ecdh for custom curves like secp256k1
	publicKeyBytes := make([]byte, 65)
	publicKeyBytes[0] = 4 // uncompressed point format
	xBytes := publicKey.X.Bytes()
	yBytes := publicKey.Y.Bytes()
	copy(publicKeyBytes[1+32-len(xBytes):33], xBytes)
	copy(publicKeyBytes[33+32-len(yBytes):], yBytes)

	// Build DER sequence
	size := len(algorithmIdentifier) + 2 + len(publicKeyBytes)
	result := make([]byte, size+4)

	// Overall SEQUENCE
	result[0] = 0x30 // SEQUENCE tag
	result[1] = byte(size)

	// Copy algorithm identifier
	copy(result[2:], algorithmIdentifier)

	// BIT STRING for public key
	result[2+len(algorithmIdentifier)] = 0x03 // BIT STRING tag
	result[2+len(algorithmIdentifier)+1] = byte(len(publicKeyBytes) + 1)
	result[2+len(algorithmIdentifier)+2] = 0x00 // No unused bits

	// Copy public key bytes
	copy(result[2+len(algorithmIdentifier)+3:], publicKeyBytes)

	return result, nil
}
