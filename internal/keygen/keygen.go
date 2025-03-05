package keygen

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
	GenerateKeyPair(keyType KeyType) (privateKey, publicKey []byte, err error)
}

// DefaultKeyGenerator implements the KeyGenerator interface
type DefaultKeyGenerator struct{}

// NewKeyGenerator creates a new DefaultKeyGenerator
func NewKeyGenerator() *DefaultKeyGenerator {
	return &DefaultKeyGenerator{}
}

// GenerateKeyPair generates a new key pair of the specified type
func (kg *DefaultKeyGenerator) GenerateKeyPair(keyType KeyType) (privateKey, publicKey []byte, err error) {
	switch keyType {
	case KeyTypeECDSA:
		return kg.generateECDSAKeyPair()
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

// generateECDSAKeyPair generates an ECDSA P-256 key pair
func (kg *DefaultKeyGenerator) generateECDSAKeyPair() (privateKey, publicKey []byte, err error) {
	// Generate ECDSA P-256 key pair
	privateECDSA, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	// Convert private key to DER format
	privDER, err := x509.MarshalECPrivateKey(privateECDSA)
	if err != nil {
		return nil, nil, err
	}

	// Convert public key to DER format
	pubDER, err := x509.MarshalPKIXPublicKey(&privateECDSA.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	// PEM encode the private key
	privateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: privDER,
	})

	// PEM encode the public key
	publicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

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
	privDER := x509.MarshalPKCS1PrivateKey(privateRSA)

	// Convert public key to PKIX DER format
	pubDER, err := x509.MarshalPKIXPublicKey(&privateRSA.PublicKey)
	if err != nil {
		return nil, nil, err
	}

	// PEM encode the private key
	privateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	})

	// PEM encode the public key
	publicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

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

	// PEM encode the private key
	privateKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pkcs8Key,
	})

	// Convert public key to DER format
	pubDER, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, nil, err
	}

	// PEM encode the public key
	publicKey = pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubDER,
	})

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
