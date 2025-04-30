package transaction

import (
	"context"
	"sort"
	"sync"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// TransactionTransformer defines an interface for modules to modify or enrich transactions.
type TransactionTransformer interface {
	// TransformTransaction applies modifications or adds data to the transaction.
	// It should return an error if the transformation fails.
	TransformTransaction(ctx context.Context, tx *types.Transaction) error
}

// Metadata keys for built-in transformers
const (
	// WalletIDMetadaKey is the key for the transformer that extracts wallet ID from metadata
	WalletIDMetadaKey = "wallet_id"

	// VaultIDMetadaKey is the key for the transformer that extracts vault ID from metadata
	VaultIDMetadaKey = "vault_id"
)

type transformerService struct {
	log              logger.Logger
	transformers     map[string]TransactionTransformer // Registered transformers
	transformerMutex sync.RWMutex                      // Mutex for transformers map
}

// NewTransformerService creates a new transformer service
func NewTransformerService(log logger.Logger) TransformerService {
	return &transformerService{
		log:              log,
		transformers:     make(map[string]TransactionTransformer),
		transformerMutex: sync.RWMutex{},
	}
}

type TransformerService interface {
	// RegisterTransformer adds a transaction transformer with a unique key.
	RegisterTransformer(key string, transformer TransactionTransformer) error

	// UnregisterTransformer removes a transaction transformer by its key.
	UnregisterTransformer(key string) error

	// TransformTransaction applies all registered transformers to a given transaction.
	TransformTransaction(ctx context.Context, tx *types.Transaction) *types.Transaction
}

// RegisterTransformer adds a transaction transformer with a unique key.
func (s *transformerService) RegisterTransformer(key string, transformer TransactionTransformer) error {
	if key == "" {
		return errors.NewInvalidInputError("Transformer key cannot be empty", "key", key)
	}
	if transformer == nil {
		return errors.NewInvalidInputError("Transformer cannot be nil", "transformer", nil)
	}

	s.transformerMutex.Lock()
	defer s.transformerMutex.Unlock()

	if _, exists := s.transformers[key]; exists {
		return errors.NewTransformerAlreadyRegisteredError(key)
	}

	s.transformers[key] = transformer
	s.log.Info("Registered transaction transformer", logger.String("key", key))
	return nil
}

// UnregisterTransformer removes a transaction transformer by its key.
func (s *transformerService) UnregisterTransformer(key string) error {
	if key == "" {
		return errors.NewInvalidInputError("Transformer key cannot be empty", "key", key)
	}

	s.transformerMutex.Lock()
	defer s.transformerMutex.Unlock()

	if _, exists := s.transformers[key]; !exists {
		return errors.NewTransformerNotFoundError(key)
	}

	delete(s.transformers, key)
	s.log.Info("Unregistered transaction transformer", logger.String("key", key))
	return nil
}

// TransformTransaction applies all registered transformers to a given transaction.
// Transformers are applied in the order determined by their sorted registration keys.
// Errors during transformation are logged, but processing continues with the next transformer.
func (s *transformerService) TransformTransaction(ctx context.Context, tx *types.Transaction) *types.Transaction {
	if tx == nil {
		s.log.Warn("TransformTransaction called with nil transaction")
		return nil
	}

	// Get keys from the map
	keys := make([]string, 0, len(s.transformers))
	for k := range s.transformers {
		keys = append(keys, k)
	}

	// Sort the keys alphabetically
	sort.Strings(keys)

	for _, k := range keys {
		transformer := s.transformers[k]
		err := transformer.TransformTransaction(ctx, tx)
		if err != nil {
			// Continue processing with the next transformer despite the error
			s.log.Error("Error applying transaction transformer",
				logger.String("transformer_key", k),
				logger.String("tx_hash", tx.Hash),
				logger.Error(err),
			)
		}
	}

	return tx
}
