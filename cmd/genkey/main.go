package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"os"

	"vault0/internal/core/crypto"
	"vault0/internal/core/keygen"
	"vault0/internal/types"
)

func main() {
	// Define command line flags
	keyType := flag.String("type", "encryption", "Type of key to generate (encryption, keypair)")
	keySize := flag.Int("size", 32, "Encryption key size in bytes (16, 24, or 32)")
	outputFormat := flag.String("format", "text", "Output format (text, env)")
	flag.Parse()

	switch *keyType {
	case "encryption":
		generateEncryptionKey(*keySize, *outputFormat)
	case "keypair":
		generateKeypair(*outputFormat)
	default:
		fmt.Printf("Invalid key type: %s. Must be 'encryption' or 'keypair'.\n", *keyType)
		os.Exit(1)
	}
}

// generateEncryptionKey generates a new encryption key for the database
func generateEncryptionKey(keySize int, format string) {
	// Validate key size
	if keySize != 16 && keySize != 24 && keySize != 32 {
		fmt.Printf("Invalid key size: %d. Must be 16, 24, or 32 bytes.\n", keySize)
		os.Exit(1)
	}

	// Generate the key
	key, err := crypto.GenerateEncryptionKeyBase64(keySize)
	if err != nil {
		fmt.Printf("Error generating encryption key: %v\n", err)
		os.Exit(1)
	}

	// Print the key
	if format == "env" {
		fmt.Printf("DB_ENCRYPTION_KEY='%s'\n", key)
	} else {
		fmt.Printf("Generated %d-byte encryption key (base64 encoded):\n\n%s\n\n", keySize, key)
		fmt.Printf("To use this key, set the DB_ENCRYPTION_KEY environment variable:\n\n")
		fmt.Printf("export DB_ENCRYPTION_KEY='%s'\n", key)
	}
}

// generateKeypair generates a new ECDSA keypair for blockchain use
func generateKeypair(format string) {
	kg := keygen.NewKeyGenerator()
	privKey, pubKey, err := kg.GenerateKeyPair(types.KeyTypeECDSA, crypto.Secp256k1Curve)
	if err != nil {
		fmt.Printf("Error generating keypair: %v\n", err)
		os.Exit(1)
	}

	if format == "env" {
		fmt.Printf("PRIVATE_KEY='%s'\n", hex.EncodeToString(privKey))
		fmt.Printf("PUBLIC_KEY='%s'\n", hex.EncodeToString(pubKey))
	} else {
		fmt.Printf("Generated ECDSA keypair:\n\n")
		fmt.Printf("Private key (hex):\n%s\n\n", hex.EncodeToString(privKey))
		fmt.Printf("Public key (hex):\n%s\n", hex.EncodeToString(pubKey))
	}
}
