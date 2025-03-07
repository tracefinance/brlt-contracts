package keystore

import (
	"context"
	"database/sql"
	"encoding/asn1"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"time"

	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"vault0/internal/config"
	"vault0/internal/keygen"
)

// DBKeyStore implements the KeyStore interface using a local database
type DBKeyStore struct {
	db           *sql.DB
	encryptor    keygen.Encryptor
	keyGenerator keygen.KeyGenerator
}

// NewDBKeyStore creates a new DBKeyStore instance
func NewDBKeyStore(db *sql.DB, cfg *config.Config) (*DBKeyStore, error) {
	if cfg.DBEncryptionKey == "" {
		return nil, errors.New("DB_ENCRYPTION_KEY environment variable is required")
	}

	// Create the encryptor
	encryptor, err := keygen.NewAESEncryptorFromBase64(cfg.DBEncryptionKey)
	if err != nil {
		return nil, err
	}

	return &DBKeyStore{
		db:           db,
		encryptor:    encryptor,
		keyGenerator: keygen.NewKeyGenerator(),
	}, nil
}

// Create creates a new key with the given ID, name, and type
func (ks *DBKeyStore) Create(ctx context.Context, id, name string, keyType keygen.KeyType, tags map[string]string) (*Key, error) {
	// Check if key with the same ID already exists
	var count int
	err := ks.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrKeyAlreadyExists
	}

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}

	// Create the key
	key := &Key{
		ID:        id,
		Name:      name,
		Type:      keyType,
		Tags:      tags,
		CreatedAt: time.Now().Unix(),
	}

	// Generate cryptographic key material based on key type
	privateKey, publicKey, err := ks.keyGenerator.GenerateKeyPair(keyType, nil)
	if err != nil {
		return nil, err
	}

	// Encrypt the private key before storing
	encryptedPrivateKey, err := ks.encryptor.Encrypt(privateKey)
	if err != nil {
		return nil, err
	}

	// Set the key material
	key.PrivateKey = encryptedPrivateKey
	key.PublicKey = publicKey

	// Insert the key into the database
	_, err = ks.db.ExecContext(
		ctx,
		"INSERT INTO keys (id, name, key_type, tags, created_at, private_key, public_key) VALUES (?, ?, ?, ?, ?, ?, ?)",
		key.ID,
		key.Name,
		string(key.Type),
		string(tagsJSON),
		key.CreatedAt,
		key.PrivateKey,
		key.PublicKey,
	)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// Import imports an existing key
func (ks *DBKeyStore) Import(ctx context.Context, id, name string, keyType keygen.KeyType, privateKey, publicKey []byte, tags map[string]string) (*Key, error) {
	// Check if key with the same ID already exists
	var count int
	err := ks.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrKeyAlreadyExists
	}

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}

	// Encrypt the private key
	encryptedPrivateKey, err := ks.encryptor.Encrypt(privateKey)
	if err != nil {
		return nil, err
	}

	// Create the key
	key := &Key{
		ID:         id,
		Name:       name,
		Type:       keyType,
		Tags:       tags,
		CreatedAt:  time.Now().Unix(),
		PrivateKey: encryptedPrivateKey,
		PublicKey:  publicKey,
	}

	// Insert the key into the database
	_, err = ks.db.ExecContext(
		ctx,
		"INSERT INTO keys (id, name, key_type, tags, created_at, private_key, public_key) VALUES (?, ?, ?, ?, ?, ?, ?)",
		key.ID,
		key.Name,
		string(key.Type),
		string(tagsJSON),
		key.CreatedAt,
		key.PrivateKey,
		key.PublicKey,
	)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// GetPublicKey retrieves only the public part of a key by its ID
func (ks *DBKeyStore) GetPublicKey(ctx context.Context, id string) (*Key, error) {
	var (
		key      Key
		keyType  string
		tagsJSON string
	)

	// Query the database
	err := ks.db.QueryRowContext(
		ctx,
		"SELECT id, name, key_type, tags, created_at, public_key FROM keys WHERE id = ?",
		id,
	).Scan(
		&key.ID,
		&key.Name,
		&keyType,
		&tagsJSON,
		&key.CreatedAt,
		&key.PublicKey,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	// Convert key type
	key.Type = keygen.KeyType(keyType)

	// Parse tags JSON
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &key.Tags); err != nil {
			return nil, err
		}
	} else {
		key.Tags = make(map[string]string)
	}

	// Set private key to nil to ensure it's never exposed
	key.PrivateKey = nil

	return &key, nil
}

// List lists all keys
func (ks *DBKeyStore) List(ctx context.Context) ([]*Key, error) {
	// Query the database
	rows, err := ks.db.QueryContext(
		ctx,
		"SELECT id, name, key_type, tags, created_at, NULL, public_key FROM keys",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*Key
	for rows.Next() {
		var (
			key      Key
			keyType  string
			tagsJSON string
		)

		err := rows.Scan(
			&key.ID,
			&key.Name,
			&keyType,
			&tagsJSON,
			&key.CreatedAt,
			&key.PrivateKey, // Will be NULL
			&key.PublicKey,
		)
		if err != nil {
			return nil, err
		}

		// Convert key type
		key.Type = keygen.KeyType(keyType)

		// Parse tags JSON
		if tagsJSON != "" {
			if err := json.Unmarshal([]byte(tagsJSON), &key.Tags); err != nil {
				return nil, err
			}
		} else {
			key.Tags = make(map[string]string)
		}

		keys = append(keys, &key)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return keys, nil
}

// Update updates a key's metadata
func (ks *DBKeyStore) Update(ctx context.Context, id string, name string, tags map[string]string) (*Key, error) {
	// Check if key exists
	var count int
	err := ks.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, ErrKeyNotFound
	}

	// Convert tags to JSON
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, err
	}

	// Update the key
	_, err = ks.db.ExecContext(
		ctx,
		"UPDATE keys SET name = ?, tags = ? WHERE id = ?",
		name,
		string(tagsJSON),
		id,
	)
	if err != nil {
		return nil, err
	}

	// Get the updated key (using public only to ensure security)
	return ks.GetPublicKey(ctx, id)
}

// Delete deletes a key by its ID
func (ks *DBKeyStore) Delete(ctx context.Context, id string) error {
	// Check if key exists
	var count int
	err := ks.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrKeyNotFound
	}

	// Delete the key
	_, err = ks.db.ExecContext(ctx, "DELETE FROM keys WHERE id = ?", id)
	return err
}

// Sign signs the provided data using the key identified by id
// This method never returns the private key material, only the signature
func (ks *DBKeyStore) Sign(ctx context.Context, id string, data []byte, dataType DataType) ([]byte, error) {
	var (
		privateKeyBytes []byte
		keyType         string
	)

	// Query the database for the private key and key type
	err := ks.db.QueryRowContext(
		ctx,
		"SELECT private_key, key_type FROM keys WHERE id = ?",
		id,
	).Scan(&privateKeyBytes, &keyType)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("key not found")
		}
		return nil, err
	}

	// Decrypt the private key
	decryptedPrivateKey, err := ks.encryptor.Decrypt(privateKeyBytes)
	if err != nil {
		return nil, err
	}

	// Sign the data based on the key type
	signature, err := ks.signData(keygen.KeyType(keyType), decryptedPrivateKey, data, dataType)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// signData signs data using the appropriate algorithm based on the key type
func (ks *DBKeyStore) signData(keyType keygen.KeyType, privateKeyBytes, data []byte, dataType DataType) ([]byte, error) {
	switch keyType {
	case keygen.KeyTypeECDSA:
		return ks.signWithECDSA(privateKeyBytes, data, dataType)
	case keygen.KeyTypeRSA:
		return ks.signWithRSA(privateKeyBytes, data, dataType)
	case keygen.KeyTypeEd25519:
		return ks.signWithEd25519(privateKeyBytes, data)
	case keygen.KeyTypeSymmetric:
		// For symmetric keys, we use HMAC instead of digital signatures
		return ks.signWithHMAC(privateKeyBytes, data)
	default:
		return nil, errors.New("unsupported key type for signing")
	}
}

// signWithECDSA signs data using an ECDSA private key
func (ks *DBKeyStore) signWithECDSA(privateKeyBytes, data []byte, dataType DataType) ([]byte, error) {
	// Parse the DER encoded private key directly (no PEM decoding needed)
	privateKey, err := x509.ParseECPrivateKey(privateKeyBytes)
	if err != nil {
		return nil, errors.New("failed to parse ECDSA private key: " + err.Error())
	}

	var digest []byte
	if dataType == DataTypeRaw {
		// Create a hash of the data
		hash := sha256.Sum256(data)
		digest = hash[:]
	} else {
		// Use the data as-is (it's already hashed)
		digest = data
	}

	// Sign the hash
	r, s, err := ecdsa.Sign(rand.Reader, privateKey, digest)
	if err != nil {
		return nil, err
	}

	// ASN.1 DER encoding for ECDSA signatures
	type ECDSASignature struct {
		R, S *big.Int
	}
	signature, err := asn1.Marshal(ECDSASignature{R: r, S: s})
	if err != nil {
		return nil, errors.New("failed to marshal ECDSA signature: " + err.Error())
	}

	return signature, nil
}

// signWithRSA signs data using an RSA private key
func (ks *DBKeyStore) signWithRSA(privateKeyBytes, data []byte, dataType DataType) ([]byte, error) {
	// Parse the DER encoded private key directly (no PEM decoding needed)
	privateKey, err := x509.ParsePKCS1PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, errors.New("failed to parse RSA private key: " + err.Error())
	}

	var digest []byte
	var hashAlgo crypto.Hash = crypto.SHA256

	if dataType == DataTypeRaw {
		// Create a hash of the data
		hash := sha256.Sum256(data)
		digest = hash[:]
	} else {
		// Use the data as-is (it's already hashed)
		// For RSA with pre-hashed data, the data must be exactly 32 bytes
		// for SHA-256 hash algorithm
		if len(data) != sha256.Size {
			return nil, fmt.Errorf("pre-hashed data must be %d bytes for SHA-256", sha256.Size)
		}
		digest = data
	}

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, hashAlgo, digest)
	if err != nil {
		return nil, err
	}

	return signature, nil
}

// signWithEd25519 signs data using an Ed25519 private key
func (ks *DBKeyStore) signWithEd25519(privateKeyBytes, data []byte) ([]byte, error) {
	// Parse the DER encoded private key directly (no PEM decoding needed)
	privateKey, err := x509.ParsePKCS8PrivateKey(privateKeyBytes)
	if err != nil {
		return nil, errors.New("failed to parse Ed25519 private key: " + err.Error())
	}

	// Convert to the correct type
	edKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not an Ed25519 key")
	}

	// Ed25519 has special handling:
	// - Ed25519 always uses the data as-is, as the algorithm has its own internal hashing
	// - Regardless of dataType, we don't pre-hash for Ed25519

	// Sign the data directly
	signature := ed25519.Sign(edKey, data)
	return signature, nil
}

// signWithHMAC creates an HMAC for symmetric keys
func (ks *DBKeyStore) signWithHMAC(keyBytes, data []byte) ([]byte, error) {
	// Create a new HMAC instance using SHA-256
	h := hmac.New(sha256.New, keyBytes)

	// For HMAC, we always use the raw data regardless of dataType
	// since HMAC already incorporates hashing
	_, err := h.Write(data)
	if err != nil {
		return nil, err
	}

	// Compute the HMAC
	return h.Sum(nil), nil
}
