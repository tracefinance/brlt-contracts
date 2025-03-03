package main

import (
	"flag"
	"fmt"
	"os"

	"vault0/internal/keymanagement"
)

func main() {
	// Parse command line arguments
	keySize := flag.Int("size", 32, "Key size in bytes (16, 24, or 32)")
	flag.Parse()

	// Validate key size
	if *keySize != 16 && *keySize != 24 && *keySize != 32 {
		fmt.Printf("Invalid key size: %d. Must be 16, 24, or 32 bytes.\n", *keySize)
		os.Exit(1)
	}

	// Generate the key
	key, err := keymanagement.GenerateEncryptionKeyBase64(*keySize)
	if err != nil {
		fmt.Printf("Error generating key: %v\n", err)
		os.Exit(1)
	}

	// Print the key
	fmt.Printf("Generated %d-byte encryption key (base64 encoded):\n\n%s\n\n", *keySize, key)
	fmt.Printf("To use this key, set the DB_ENCRYPTION_KEY environment variable:\n\n")
	fmt.Printf("export DB_ENCRYPTION_KEY='%s'\n", key)
}
