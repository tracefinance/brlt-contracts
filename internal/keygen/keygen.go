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
)

// KeyType represents the type of cryptographic key
type KeyType string

const (
	// KeyTypeECDSA represents ECDSA keys
	KeyTypeECDSA KeyType = "ecdsa"
	// KeyTypeRSA represents RSA keys
	KeyTypeRSA KeyType = "rsa"
	// KeyTypeEd25519 represents Ed25519 keys
	KeyTypeEd25519 KeyType = "ed25519"
	// KeyTypeSymmetric represents symmetric keys
	KeyTypeSymmetric KeyType = "symmetric"
)

// KeyGenerator defines an interface for generating cryptographic keys
type KeyGenerator interface {
	// GenerateKeyPair generates a new key pair of the specified type
	// For ECDSA keys, an optional curve can be provided. If nil, P-256 is used.
	GenerateKeyPair(keyType KeyType, curve elliptic.Curve) (privateKey, publicKey []byte, err error)
}

// DefaultKeyGenerator implements the KeyGenerator interface
type DefaultKeyGenerator struct{}

// NewKeyGenerator creates a new DefaultKeyGenerator
func NewKeyGenerator() *DefaultKeyGenerator {
	return &DefaultKeyGenerator{}
}

// GenerateKeyPair generates a new key pair of the specified type
// For ECDSA keys, an optional curve can be provided. If nil, P-256 is used.
func (kg *DefaultKeyGenerator) GenerateKeyPair(keyType KeyType, curve elliptic.Curve) (privateKey, publicKey []byte, err error) {
	switch keyType {
	case KeyTypeECDSA:
		return kg.generateECDSAKeyPair(curve)
	case KeyTypeRSA:
		return kg.generateRSAKeyPair()
	case KeyTypeEd25519:
		return kg.generateEd25519KeyPair()
	case KeyTypeSymmetric:
		return kg.generateSymmetricKey()
	default:
		return nil, nil, fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// generateECDSAKeyPair generates an ECDSA key pair with the specified curve
// If curve is nil, P-256 is used
func (kg *DefaultKeyGenerator) generateECDSAKeyPair(curve elliptic.Curve) (privateKey, publicKey []byte, err error) {
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
	if curve == Secp256k1 {
		// Use our custom marshalling for SECP256K1
		privateKey, err = marshalSecp256k1PrivateKey(privateECDSA)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to marshal SECP256K1 private key: %w", err)
		}

		// Use custom marshalling for public key too
		publicKey, err = marshalSecp256k1PublicKey(&privateECDSA.PublicKey)
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

// generateRSAKeyPair generates an RSA 2048-bit key pair
func (kg *DefaultKeyGenerator) generateRSAKeyPair() (privateKey, publicKey []byte, err error) {
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

// generateEd25519KeyPair generates an Ed25519 key pair
func (kg *DefaultKeyGenerator) generateEd25519KeyPair() (privateKey, publicKey []byte, err error) {
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

// generateSymmetricKey generates a 32-byte (256-bit) symmetric key
func (kg *DefaultKeyGenerator) generateSymmetricKey() (privateKey, publicKey []byte, err error) {
	// Generate a 32-byte (256-bit) symmetric key
	privateKey = make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, privateKey); err != nil {
		return nil, nil, err
	}
	// No public key for symmetric keys
	return privateKey, nil, nil
}
