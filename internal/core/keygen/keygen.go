// Package keygen provides cryptographic key generation functionality for various
// key types and algorithms.
//
// The keygen package is part of the Core/Infrastructure Layer and provides
// secure key generation capabilities for different cryptographic algorithms:
//   - ECDSA (P-256, P-384, P-521, SECP256K1)
//   - RSA (2048-bit)
//   - Ed25519
//   - Symmetric (256-bit)
//
// Key Generation Features:
//   - Secure random number generation using crypto/rand
//   - Support for multiple key types and curves
//   - Standard-compliant key formatting (PKCS#8, PKIX)
//   - Custom handling for blockchain-specific curves (SECP256K1)
package keygen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io"

	"vault0/internal/core/crypto"
	"vault0/internal/types"
)

// KeyGenerator defines an interface for generating cryptographic keys.
// Implementations must ensure secure random number generation and proper
// key formatting according to cryptographic standards.
type KeyGenerator interface {
	// GenerateKeyPair generates a new key pair of the specified type.
	//
	// The function supports multiple key types:
	//   - ECDSA: Supports P-256 (default), P-384, P-521, and SECP256K1 curves
	//   - RSA: Generates 2048-bit keys
	//   - Ed25519: Generates Ed25519 key pairs
	//   - Symmetric: Generates 256-bit symmetric keys
	//
	// Parameters:
	//   - keyType: The type of key to generate (ECDSA, RSA, Ed25519, Symmetric)
	//   - curve: For ECDSA keys, specifies the curve to use. If nil, P-256 is used
	//           For non-ECDSA keys, this parameter is ignored
	//
	// Returns:
	//   - privateKey: DER-encoded private key (PKCS#8 for Ed25519, PKCS#1 for RSA)
	//   - publicKey: DER-encoded public key (PKIX format)
	//   - error: Any error that occurred during key generation
	GenerateKeyPair(keyType types.KeyType, curve elliptic.Curve) (privateKey, publicKey []byte, err error)
}

// defaultKeyGenerator implements the KeyGenerator interface using Go's crypto packages.
// It provides secure key generation using crypto/rand as the random number source.
type defaultKeyGenerator struct{}

// NewKeyGenerator creates a new DefaultKeyGenerator instance.
// This generator uses crypto/rand for secure random number generation.
func NewKeyGenerator() KeyGenerator {
	return &defaultKeyGenerator{}
}

// GenerateKeyPair implements the KeyGenerator interface.
// See the interface documentation for details about supported key types and parameters.
func (kg *defaultKeyGenerator) GenerateKeyPair(keyType types.KeyType, curve elliptic.Curve) (privateKey, publicKey []byte, err error) {
	switch keyType {
	case types.KeyTypeECDSA:
		return kg.generateECDSAKeyPair(curve)
	case types.KeyTypeRSA:
		return kg.generateRSAKeyPair()
	case types.KeyTypeEd25519:
		return kg.generateEd25519KeyPair()
	case types.KeyTypeSymmetric:
		return kg.generateSymmetricKey()
	default:
		return nil, nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// generateECDSAKeyPair generates an ECDSA key pair with the specified curve.
//
// Supported curves:
//   - P-256 (default if curve is nil)
//   - P-384
//   - P-521
//   - SECP256K1 (for blockchain operations)
//
// The function handles SECP256K1 curve differently from standard NIST curves,
// using custom marshalling functions from the crypto package.
//
// Returns:
//   - privateKey: DER-encoded private key
//   - publicKey: DER-encoded public key
//   - error: Any error during generation or encoding
func (kg *defaultKeyGenerator) generateECDSAKeyPair(curve elliptic.Curve) (privateKey, publicKey []byte, err error) {
	// Use P-256 as the default curve if none is provided
	if curve == nil {
		curve = elliptic.P256()
	}

	// Generate ECDSA key pair with the specified curve
	privateECDSA, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Special handling for SECP256K1 curve
	if curve == crypto.Secp256k1Curve {
		// Use our custom marshalling for SECP256K1
		privateKey, err = crypto.MarshalPrivateKey(privateECDSA)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal SECP256K1 private key: %w", err)
		}

		// Use custom marshalling for public key too
		publicKey, err = crypto.MarshalPublicKey(&privateECDSA.PublicKey)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal SECP256K1 public key: %w", err)
		}
	} else {
		// For standard curves, use the regular EC marshalling
		privateKey, err = x509.MarshalECPrivateKey(privateECDSA)
		if err != nil {
			return nil, nil, err
		}

		// Public key marshalling for standard curves
		publicKey, err = x509.MarshalPKIXPublicKey(&privateECDSA.PublicKey)
		if err != nil {
			return nil, nil, err
		}
	}

	return privateKey, publicKey, nil
}

// generateRSAKeyPair generates a 2048-bit RSA key pair.
//
// The private key is encoded in PKCS#1 DER format, while the public
// key is encoded in PKIX format. The key size of 2048 bits provides
// a good balance between security and performance for most applications.
//
// Returns:
//   - privateKey: DER-encoded private key (PKCS#1)
//   - publicKey: DER-encoded public key (PKIX)
//   - error: Any error during generation or encoding
func (kg *defaultKeyGenerator) generateRSAKeyPair() (privateKey, publicKey []byte, err error) {
	// Generate RSA 2048-bit key pair
	privateRSA, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// Convert private key to PKCS#1 DER format
	privateKey = x509.MarshalPKCS1PrivateKey(privateRSA)

	// Convert public key to PKIX DER format
	publicKey, err = x509.MarshalPKIXPublicKey(&privateRSA.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	return privateKey, publicKey, nil
}

// generateEd25519KeyPair generates an Ed25519 key pair.
//
// Ed25519 is a modern signature scheme that provides strong security
// and good performance. The private key is encoded in PKCS#8 format,
// which is the standard format for Ed25519 keys.
//
// Returns:
//   - privateKey: DER-encoded private key (PKCS#8)
//   - publicKey: DER-encoded public key (PKIX)
//   - error: Any error during generation or encoding
func (kg *defaultKeyGenerator) generateEd25519KeyPair() (privateKey, publicKey []byte, err error) {
	// Generate Ed25519 key pair
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// For Ed25519, we need to convert to PKCS8 format for PEM encoding
	pkcs8Key, err := x509.MarshalPKCS8PrivateKey(privKey)
	if err != nil {
		return nil, nil, err
	}

	// Use the PKCS8 DER encoding for the private key
	privateKey = pkcs8Key

	// Convert public key to DER format
	pubDER, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, nil, err
	}

	// Return the DER-encoded keys directly instead of PEM encoding them
	publicKey = pubDER

	return privateKey, publicKey, nil
}

// generateSymmetricKey generates a 32-byte (256-bit) symmetric key.
//
// This function generates a cryptographically secure random key suitable
// for use with symmetric encryption algorithms like AES-256. The key is
// generated using crypto/rand for maximum security.
//
// Returns:
//   - privateKey: 32-byte random key
//   - publicKey: nil (symmetric keys don't have public components)
//   - error: Any error during key generation
func (kg *defaultKeyGenerator) generateSymmetricKey() (privateKey, publicKey []byte, err error) {
	// Generate a 32-byte (256-bit) symmetric key
	privateKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privateKey); err != nil {
		return nil, nil, err
	}
	// No public key for symmetric keys
	return privateKey, nil, nil
}
