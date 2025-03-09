package keygen

import (
	"crypto/ecdsa"
	"encoding/asn1"
	"errors"
	"fmt"
	"math/big"
)

// secp256k1OID is the Object Identifier for the SECP256K1 curve
var secp256k1OID = asn1.ObjectIdentifier{1, 3, 132, 0, 10}

// ecPrivateKey represents the ASN.1 structure for an EC private key as per RFC 5915
type ecPrivateKey struct {
	Version       int
	PrivateKey    []byte
	NamedCurveOID asn1.ObjectIdentifier `asn1:"optional,explicit,tag:0"`
	PublicKey     asn1.BitString        `asn1:"optional,explicit,tag:1"`
}

// MarshalPrivateKey marshals an SECP256K1 private key into DER format
func MarshalPrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, errors.New("nil private key")
	}

	// Verify the curve is SECP256K1
	if key.Curve != Secp256k1Curve {
		return nil, fmt.Errorf("invalid curve: expected SECP256K1")
	}

	// Ensure the private key scalar is within the curve order
	if key.D.Cmp(Secp256k1Curve.Params().N) >= 0 {
		return nil, fmt.Errorf("private key scalar is too large")
	}

	// Convert the private key scalar (D) to 32 bytes, padding if necessary
	privateKeyBytes := key.D.Bytes()
	if len(privateKeyBytes) > 32 {
		return nil, fmt.Errorf("private key scalar exceeds 32 bytes")
	}
	privateKeyBytes = padLeft(privateKeyBytes, 32)

	// Marshal the public key (uncompressed format: 65 bytes)
	publicKeyBytes, err := MarshalPublicKey(&key.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	// Construct the ECPrivateKey structure
	ecPrivKey := ecPrivateKey{
		Version:       1,
		PrivateKey:    privateKeyBytes,
		NamedCurveOID: secp256k1OID,
		PublicKey:     asn1.BitString{Bytes: publicKeyBytes, BitLength: len(publicKeyBytes) * 8},
	}

	// Encode to DER format
	derBytes, err := asn1.Marshal(ecPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key to DER: %w", err)
	}

	return derBytes, nil
}

// UnmarshalPrivateKey unmarshals a DER-encoded SECP256K1 private key
func UnmarshalPrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	var ecPrivKey ecPrivateKey
	_, err := asn1.Unmarshal(data, &ecPrivKey)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal DER private key: %w", err)
	}

	// Check version
	if ecPrivKey.Version != 1 {
		return nil, fmt.Errorf("unsupported EC private key version: %d", ecPrivKey.Version)
	}

	// Verify the curve OID
	if !ecPrivKey.NamedCurveOID.Equal(secp256k1OID) {
		return nil, fmt.Errorf("unsupported curve OID: %v", ecPrivKey.NamedCurveOID)
	}

	// Validate private key length (must be 32 bytes for SECP256K1)
	if len(ecPrivKey.PrivateKey) != 32 {
		return nil, fmt.Errorf("invalid private key length: expected 32 bytes, got %d", len(ecPrivKey.PrivateKey))
	}

	// Convert private key bytes to big.Int
	d := new(big.Int).SetBytes(ecPrivKey.PrivateKey)
	if d.Cmp(Secp256k1Curve.Params().N) >= 0 {
		return nil, fmt.Errorf("private key scalar is too large")
	}

	// Compute the public key from the private key scalar
	x, y := Secp256k1Curve.ScalarBaseMult(ecPrivKey.PrivateKey)

	// If public key is included, verify it matches the computed values
	if len(ecPrivKey.PublicKey.Bytes) > 0 {
		pubKey, err := UnmarshalPublicKey(ecPrivKey.PublicKey.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal public key from DER: %w", err)
		}
		if pubKey.X.Cmp(x) != 0 || pubKey.Y.Cmp(y) != 0 {
			return nil, errors.New("public key in DER does not match computed public key")
		}
	}

	// Construct and return the private key
	return &ecdsa.PrivateKey{
		D: d,
		PublicKey: ecdsa.PublicKey{
			Curve: Secp256k1Curve,
			X:     x,
			Y:     y,
		},
	}, nil
}

// MarshalPublicKey serializes an ECDSA public key into the uncompressed format.
// Format: 0x04 || 32-byte X coordinate || 32-byte Y coordinate (65 bytes total).
func MarshalPublicKey(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil, errors.New("invalid public key: nil pointer")
	}

	// Verify the curve is SECP256K1
	if pub.Curve != Secp256k1Curve {
		return nil, fmt.Errorf("invalid curve: expected SECP256K1")
	}

	// Format: 0x04 || X || Y
	xBytes := padLeft(pub.X.Bytes(), 32)
	yBytes := padLeft(pub.Y.Bytes(), 32)

	result := make([]byte, 65)
	result[0] = 0x04
	copy(result[1:33], xBytes)
	copy(result[33:], yBytes)
	return result, nil
}

// UnmarshalPublicKey deserializes a SECP256K1 public key from its uncompressed byte representation.
func UnmarshalPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	// Validate the input length and prefix
	if len(data) != 65 {
		return nil, fmt.Errorf("invalid public key length: expected 65 bytes, got %d", len(data))
	}
	if data[0] != 0x04 {
		return nil, fmt.Errorf("invalid public key prefix: expected 0x04, got 0x%02x", data[0])
	}

	// Extract and convert X and Y coordinates
	x := new(big.Int).SetBytes(data[1:33])
	y := new(big.Int).SetBytes(data[33:65])

	// Verify the point is on the curve
	if !Secp256k1Curve.IsOnCurve(x, y) {
		return nil, errors.New("public key coordinates are not on the SECP256K1 curve")
	}

	// Return the constructed public key
	return &ecdsa.PublicKey{
		Curve: Secp256k1Curve,
		X:     x,
		Y:     y,
	}, nil
}

// padLeft pads a byte slice with leading zeros to reach the specified length
func padLeft(b []byte, n int) []byte {
	if len(b) >= n {
		return b
	}
	padded := make([]byte, n)
	copy(padded[n-len(b):], b)
	return padded
}
