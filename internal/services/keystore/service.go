package keystore

import (
	"context"
	"crypto/elliptic"

	"vault0/internal/core/keystore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/wallet"
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

	// GetKeyById retrieves a key by ID
	GetKeyById(ctx context.Context, id string) (*keystore.Key, error)

	// ListKeys retrieves keys with optional filtering and pagination
	// limit specifies the maximum number of keys to return (0 means use default)
	// nextToken is used for pagination (empty string for first page)
	// returns a Page of keys and error if any
	ListKeys(ctx context.Context, filter KeyFilter, limit int, nextToken string) (*types.Page[*keystore.Key], error)

	// UpdateKey modifies a key's metadata
	UpdateKey(ctx context.Context, id string, name string, tags map[string]string) (*keystore.Key, error)

	// DeleteKey removes a key from the keystore
	DeleteKey(ctx context.Context, id string) error
}

// service implements the Service interface
type service struct {
	keyStore      keystore.KeyStore
	log           logger.Logger
	walletService wallet.Service
}

// NewService creates a new keystore service instance
func NewService(keyStore keystore.KeyStore, log logger.Logger, walletSvc wallet.Service) Service {
	return &service{
		keyStore:      keyStore,
		log:           log,
		walletService: walletSvc,
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

// GetKeyById implements the Service interface
func (s *service) GetKeyById(ctx context.Context, id string) (*keystore.Key, error) {
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
func (s *service) ListKeys(ctx context.Context, filter KeyFilter, limit int, nextToken string) (*types.Page[*keystore.Key], error) {
	// Get keys with pagination
	keysPage, err := s.keyStore.List(ctx, limit, nextToken)
	if err != nil {
		s.log.Error("Failed to list keys", logger.Error(err))
		return nil, err
	}

	// If no filtering is required, return the page as-is
	if filter.KeyType == nil && len(filter.Tags) == 0 {
		return keysPage, nil
	}

	// Apply filters if specified
	filtered := make([]*keystore.Key, 0)

	for _, key := range keysPage.Items {
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

	// Create a new page with the filtered items
	// Note: When applying filters, we lose the ability to properly use the NextToken
	// since we're filtering after pagination. In a production system, you might want
	// to implement filtering at the database level for better pagination support.
	filteredPage := &types.Page[*keystore.Key]{
		Items:     filtered,
		Limit:     keysPage.Limit,
		NextToken: keysPage.NextToken,
	}

	return filteredPage, nil
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
	// Check if the key is associated with any wallets using the wallet service
	wallets, err := s.walletService.FindWalletsByKeyID(ctx, id)
	if err != nil {
		s.log.Error("Failed to check key association with wallets via service",
			logger.Error(err),
			logger.String("key_id", id))
		return err
	}

	// If associated (wallets list is not empty), prevent deletion and return the specific error
	if len(wallets) > 0 {
		s.log.Warn("Attempted to delete key associated with one or more wallets",
			logger.String("key_id", id),
			logger.Int("associated_wallets_count", len(wallets)))
		return errors.NewKeyInUseByWalletError(id)
	}

	// If not associated, proceed with deletion
	err = s.keyStore.Delete(ctx, id)
	if err != nil {
		s.log.Error("Failed to delete key from keystore",
			logger.Error(err),
			logger.String("key_id", id))
		return err
	}

	s.log.Info("Key deleted successfully", logger.String("key_id", id))

	return nil
}
