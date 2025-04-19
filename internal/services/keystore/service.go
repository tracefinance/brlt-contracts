package keystore

import (
	"context"
	"crypto/elliptic"

	"vault0/internal/core/keystore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Service defines the keystore service interface
type Service interface {
	// CreateKey generates a new cryptographic key
	CreateKey(ctx context.Context, name string, keyType types.KeyType, curveName string, tags map[string]string) (*keystore.Key, error)

	// ImportKey stores an existing key in the keystore
	ImportKey(ctx context.Context, name string, keyType types.KeyType, curveName string, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error)

	// SignData performs a cryptographic signing operation using the specified key
	SignData(ctx context.Context, id string, data []byte, rawData bool) ([]byte, error)

	// GetKey retrieves a key by ID
	GetKey(ctx context.Context, id string) (*keystore.Key, error)

	// ListKeys retrieves keys with optional filtering
	ListKeys(ctx context.Context, filter KeyFilter) ([]*keystore.Key, error)

	// UpdateKey modifies a key's metadata
	UpdateKey(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error)

	// DeleteKey removes a key from the keystore
	DeleteKey(ctx context.Context, id string) error
}

// service implements the Service interface
type service struct {
	keyStore keystore.KeyStore
	log      logger.Logger
}

// NewService creates a new keystore service instance
func NewService(keyStore keystore.KeyStore, log logger.Logger) Service {
	return &service{
		keyStore: keyStore,
		log:      log,
	}
}

// getCurve converts a curve name string to an elliptic.Curve
func getCurve(curveName string) (elliptic.Curve, error) {
	switch curveName {
	case "P224":
		return elliptic.P224(), nil
	case "P256":
		return elliptic.P256(), nil
	case "P384":
		return elliptic.P384(), nil
	case "P521":
		return elliptic.P521(), nil
	case "":
		return nil, nil // Allow nil curve for non-ECDSA keys
	default:
		return nil, errors.NewInvalidParameterError("curve", "must be one of: P224, P256, P384, P521")
	}
}

// CreateKey implements the Service interface
func (s *service) CreateKey(ctx context.Context, name string, keyType types.KeyType, curveName string, tags map[string]string) (*keystore.Key, error) {
	// Convert curve name to elliptic.Curve
	curve, err := getCurve(curveName)
	if err != nil {
		s.log.Error("Invalid curve specified",
			logger.Error(err),
			logger.String("curve", curveName))
		return nil, err
	}

	// Check key type compatibility with curve
	if keyType == types.KeyTypeECDSA && curve == nil {
		s.log.Error("Curve is required for ECDSA keys")
		return nil, errors.NewInvalidParameterError("curve", "is required for ECDSA keys")
	}

	// Create the key
	key, err := s.keyStore.Create(ctx, name, keyType, curve, tags)
	if err != nil {
		s.log.Error("Failed to create key",
			logger.Error(err),
			logger.String("name", name),
			logger.String("key_type", string(keyType)))
		return nil, err
	}

	s.log.Info("Key created successfully",
		logger.String("id", key.ID),
		logger.String("name", key.Name),
		logger.String("key_type", string(keyType)))

	return key, nil
}

// ImportKey implements the Service interface
func (s *service) ImportKey(ctx context.Context, name string, keyType types.KeyType, curveName string, privateKey, publicKey []byte, tags map[string]string) (*keystore.Key, error) {
	// Convert curve name to elliptic.Curve
	curve, err := getCurve(curveName)
	if err != nil {
		s.log.Error("Invalid curve specified",
			logger.Error(err),
			logger.String("curve", curveName))
		return nil, err
	}

	// Check key type compatibility with curve
	if keyType == types.KeyTypeECDSA && curve == nil {
		s.log.Error("Curve is required for ECDSA keys")
		return nil, errors.NewInvalidParameterError("curve", "is required for ECDSA keys")
	}

	// Validate private key
	if len(privateKey) == 0 {
		s.log.Error("Private key is required for import")
		return nil, errors.NewMissingParameterError("private_key")
	}

	// For asymmetric keys, public key is required
	if (keyType == types.KeyTypeECDSA || keyType == types.KeyTypeRSA || keyType == types.KeyTypeEd25519) && len(publicKey) == 0 {
		s.log.Error("Public key is required for asymmetric keys")
		return nil, errors.NewMissingParameterError("public_key")
	}

	// Import the key
	key, err := s.keyStore.Import(ctx, name, keyType, curve, privateKey, publicKey, tags)
	if err != nil {
		s.log.Error("Failed to import key",
			logger.Error(err),
			logger.String("name", name),
			logger.String("key_type", string(keyType)))
		return nil, err
	}

	s.log.Info("Key imported successfully",
		logger.String("id", key.ID),
		logger.String("name", key.Name),
		logger.String("key_type", string(keyType)))

	return key, nil
}

// SignData implements the Service interface
func (s *service) SignData(ctx context.Context, id string, data []byte, rawData bool) ([]byte, error) {
	dataType := keystore.DataTypeDigest
	if rawData {
		dataType = keystore.DataTypeRaw
	}

	signature, err := s.keyStore.Sign(ctx, id, data, dataType)
	if err != nil {
		s.log.Error("Failed to sign data",
			logger.Error(err),
			logger.String("key_id", id))
		return nil, err
	}

	s.log.Info("Data signed successfully",
		logger.String("key_id", id),
		logger.String("data_type", string(dataType)))

	return signature, nil
}

// GetKey implements the Service interface
func (s *service) GetKey(ctx context.Context, id string) (*keystore.Key, error) {
	key, err := s.keyStore.GetPublicKey(ctx, id)
	if err != nil {
		s.log.Error("Failed to get key",
			logger.Error(err),
			logger.String("key_id", id))
		return nil, err
	}

	return key, nil
}

// ListKeys implements the Service interface
func (s *service) ListKeys(ctx context.Context, filter KeyFilter) ([]*keystore.Key, error) {
	// Get all keys
	keys, err := s.keyStore.List(ctx)
	if err != nil {
		s.log.Error("Failed to list keys", logger.Error(err))
		return nil, err
	}

	// Apply filters if specified
	if filter.KeyType != nil || len(filter.Tags) > 0 {
		filtered := make([]*keystore.Key, 0)

		for _, key := range keys {
			// Apply key type filter
			if filter.KeyType != nil && key.Type != *filter.KeyType {
				continue
			}

			// Apply tags filter (all specified tags must match)
			matchesTags := true
			for tagKey, tagValue := range filter.Tags {
				if val, exists := key.Tags[tagKey]; !exists || val != tagValue {
					matchesTags = false
					break
				}
			}

			if !matchesTags {
				continue
			}

			filtered = append(filtered, key)
		}

		return filtered, nil
	}

	return keys, nil
}

// UpdateKey implements the Service interface
func (s *service) UpdateKey(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error) {
	// Update the key
	key, err := s.keyStore.Update(ctx, id, name, tags)
	if err != nil {
		s.log.Error("Failed to update key",
			logger.Error(err),
			logger.String("key_id", id))
		return nil, err
	}

	s.log.Info("Key updated successfully",
		logger.String("id", key.ID),
		logger.String("name", key.Name))

	return key, nil
}

// DeleteKey implements the Service interface
func (s *service) DeleteKey(ctx context.Context, id string) error {
	// First, verify the key exists
	_, err := s.keyStore.GetPublicKey(ctx, id)
	if err != nil {
		s.log.Error("Failed to get key for deletion",
			logger.Error(err),
			logger.String("key_id", id))
		return err
	}

	// Delete the key
	err = s.keyStore.Delete(ctx, id)
	if err != nil {
		s.log.Error("Failed to delete key",
			logger.Error(err),
			logger.String("key_id", id))
		return err
	}

	s.log.Info("Key deleted successfully", logger.String("key_id", id))
	return nil
}
