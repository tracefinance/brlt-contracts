package keygen

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
)

// MarshalPrivateKey marshals a SECP256K1 private key to bytes
func MarshalPrivateKey(key *ecdsa.PrivateKey) ([]byte, error) {
	if key == nil {
		return nil, errors.New("nil private key")
	}

	// Verify the key is using secp256k1 curve
	if key.Curve != Secp256k1Curve {
		return nil, fmt.Errorf("invalid curve: expected secp256k1")
	}

	// Verify the private key scalar is within valid range
	if key.D.Cmp(Secp256k1Curve.Params().N) >= 0 {
		return nil, fmt.Errorf("private key scalar is too large")
	}

	// Return the private key scalar as bytes
	return key.D.Bytes(), nil
}

// UnmarshalPrivateKey unmarshals a SECP256K1 private key from bytes
func UnmarshalPrivateKey(data []byte) (*ecdsa.PrivateKey, error) {
	if len(data) == 0 {
		return nil, errors.New("empty private key data")
	}

	// Convert bytes to scalar
	d := new(big.Int).SetBytes(data)

	// Verify the scalar is within valid range
	if d.Cmp(Secp256k1Curve.Params().N) >= 0 {
		return nil, fmt.Errorf("private key scalar is too large")
	}

	// Create private key
	priv := new(ecdsa.PrivateKey)
	priv.Curve = Secp256k1Curve
	priv.D = d

	// Calculate public key
	priv.PublicKey.X, priv.PublicKey.Y = Secp256k1Curve.ScalarBaseMult(data)

	return priv, nil
}

// MarshalPublicKey serializes an ECDSA public key into Ethereum's uncompressed format.
// Format: 0x04 || 32-byte X coordinate || 32-byte Y coordinate (65 bytes total).
func MarshalPublicKey(pub *ecdsa.PublicKey) ([]byte, error) {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil, errors.New("invalid public key")
	}

	// Get X and Y coordinates as byte slices
	xBytes := pub.X.Bytes()
	yBytes := pub.Y.Bytes()

	// Ensure coordinates are not larger than 32 bytes
	if len(xBytes) > 32 || len(yBytes) > 32 {
		return nil, errors.New("public key coordinates exceed 32 bytes")
	}

	// Pad coordinates to 32 bytes (left-pad with zeros if necessary)
	xPadded := make([]byte, 32)
	yPadded := make([]byte, 32)
	copy(xPadded[32-len(xBytes):], xBytes)
	copy(yPadded[32-len(yBytes):], yBytes)

	// Construct uncompressed format: 0x04 || X || Y
	return append([]byte{0x04}, append(xPadded, yPadded...)...), nil
}

// UnmarshalPublicKey unmarshals a SECP256K1 public key from bytes
func UnmarshalPublicKey(data []byte) (*ecdsa.PublicKey, error) {
	// Validate input length and prefix
	if len(data) != 65 || data[0] != 0x04 {
		return nil, errors.New("invalid uncompressed public key format")
	}

	// Check length
	byteLen := (Secp256k1Curve.Params().BitSize + 7) / 8
	if len(data) != 1+2*byteLen {
		return nil, fmt.Errorf("invalid public key length")
	}
	// Extract X and Y coordinates
	x := new(big.Int).SetBytes(data[1:33])  // First 32 bytes after 0x04
	y := new(big.Int).SetBytes(data[33:65]) // Next 32 bytes

	// Verify point is on curve
	if !Secp256k1Curve.IsOnCurve(x, y) {
		return nil, fmt.Errorf("point is not on secp256k1 curve")
	}

	// Construct the public key
	return &ecdsa.PublicKey{
		Curve: Secp256k1Curve,
		X:     x,
		Y:     y,
	}, nil
}
