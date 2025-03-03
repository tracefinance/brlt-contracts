package keymanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"vault0/internal/config"
)

// DBKeyManager implements the KeyManager interface using a local database
type DBKeyManager struct {
	db           *sql.DB
	encryptor    Encryptor
	keyGenerator KeyGenerator
}

// NewDBKeyManager creates a new DBKeyManager instance
func NewDBKeyManager(db *sql.DB, cfg *config.Config) (*DBKeyManager, error) {
	if cfg.DBEncryptionKey == "" {
		return nil, errors.New("DB_ENCRYPTION_KEY environment variable is required")
	}

	// Create the encryptor
	encryptor, err := NewAESEncryptorFromBase64(cfg.DBEncryptionKey)
	if err != nil {
		return nil, err
	}

	return &DBKeyManager{
		db:           db,
		encryptor:    encryptor,
		keyGenerator: NewKeyGenerator(),
	}, nil
}

// Create creates a new key with the given ID, name, and type
func (km *DBKeyManager) Create(ctx context.Context, id, name string, keyType KeyType, tags map[string]string) (*Key, error) {
	// Check if key with the same ID already exists
	var count int
	err := km.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
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
	privateKey, publicKey, err := km.keyGenerator.GenerateKeyPair(keyType)
	if err != nil {
		return nil, err
	}

	// Encrypt the private key before storing
	encryptedPrivateKey, err := km.encryptor.Encrypt(privateKey)
	if err != nil {
		return nil, err
	}

	// Set the key material
	key.PrivateKey = encryptedPrivateKey
	key.PublicKey = publicKey

	// Insert the key into the database
	_, err = km.db.ExecContext(
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
func (km *DBKeyManager) Import(ctx context.Context, id, name string, keyType KeyType, privateKey, publicKey []byte, tags map[string]string) (*Key, error) {
	// Check if key with the same ID already exists
	var count int
	err := km.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
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
	encryptedPrivateKey, err := km.encryptor.Encrypt(privateKey)
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
	_, err = km.db.ExecContext(
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

// Get retrieves a key by its ID
func (km *DBKeyManager) Get(ctx context.Context, id string) (*Key, error) {
	var (
		key      Key
		keyType  string
		tagsJSON string
	)

	// Query the database
	err := km.db.QueryRowContext(
		ctx,
		"SELECT id, name, key_type, tags, created_at, private_key, public_key FROM keys WHERE id = ?",
		id,
	).Scan(
		&key.ID,
		&key.Name,
		&keyType,
		&tagsJSON,
		&key.CreatedAt,
		&key.PrivateKey,
		&key.PublicKey,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	// Convert key type
	key.Type = KeyType(keyType)

	// Parse tags JSON
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &key.Tags); err != nil {
			return nil, err
		}
	} else {
		key.Tags = make(map[string]string)
	}

	// Decrypt the private key if present
	if len(key.PrivateKey) > 0 {
		decryptedPrivateKey, err := km.encryptor.Decrypt(key.PrivateKey)
		if err != nil {
			return nil, err
		}
		key.PrivateKey = decryptedPrivateKey
	}

	return &key, nil
}

// GetPublicOnly retrieves only the public part of a key by its ID
func (km *DBKeyManager) GetPublicOnly(ctx context.Context, id string) (*Key, error) {
	var (
		key      Key
		keyType  string
		tagsJSON string
	)

	// Query the database
	err := km.db.QueryRowContext(
		ctx,
		"SELECT id, name, key_type, tags, created_at, NULL, public_key FROM keys WHERE id = ?",
		id,
	).Scan(
		&key.ID,
		&key.Name,
		&keyType,
		&tagsJSON,
		&key.CreatedAt,
		&key.PrivateKey, // Will be NULL
		&key.PublicKey,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}

	// Convert key type
	key.Type = KeyType(keyType)

	// Parse tags JSON
	if tagsJSON != "" {
		if err := json.Unmarshal([]byte(tagsJSON), &key.Tags); err != nil {
			return nil, err
		}
	} else {
		key.Tags = make(map[string]string)
	}

	return &key, nil
}

// List lists all keys
func (km *DBKeyManager) List(ctx context.Context) ([]*Key, error) {
	// Query the database
	rows, err := km.db.QueryContext(
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
		key.Type = KeyType(keyType)

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
func (km *DBKeyManager) Update(ctx context.Context, id string, name string, tags map[string]string) (*Key, error) {
	// Check if key exists
	var count int
	err := km.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
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
	_, err = km.db.ExecContext(
		ctx,
		"UPDATE keys SET name = ?, tags = ? WHERE id = ?",
		name,
		string(tagsJSON),
		id,
	)
	if err != nil {
		return nil, err
	}

	// Get the updated key
	return km.Get(ctx, id)
}

// Delete deletes a key by its ID
func (km *DBKeyManager) Delete(ctx context.Context, id string) error {
	// Check if key exists
	var count int
	err := km.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM keys WHERE id = ?", id).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrKeyNotFound
	}

	// Delete the key
	_, err = km.db.ExecContext(ctx, "DELETE FROM keys WHERE id = ?", id)
	return err
}

// Close releases any resources used by the key manager
func (km *DBKeyManager) Close() error {
	// No resources to release
	return nil
}
