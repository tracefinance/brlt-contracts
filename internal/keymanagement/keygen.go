package keymanagement

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
	privateKey, err = x509.MarshalECPrivateKey(privateECDSA)
	if err != nil {
		return nil, nil, err
	}

	// Convert public key to DER format
	publicKey, err = x509.MarshalPKIXPublicKey(&privateECDSA.PublicKey)
	if err != nil {
		return nil, nil, err
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

	return privKey, pubKey, nil
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
