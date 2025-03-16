package types

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

const (
	CurveNameP256      = "P-256"
	CurveNameSecp256k1 = "secp256k1"
)
