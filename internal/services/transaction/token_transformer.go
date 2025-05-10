package transaction

import (
	"context"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// TokenTransformer defines methods for enriching transactions with token data
type TokenTransformer interface {
	// TransformTransaction enriches a transaction with token data
	TransformTransaction(ctx context.Context, tx *types.Transaction) error
}

// NewTokenTransformer creates a new token transformer
func NewTokenTransformer(
	log logger.Logger,
	tokenStore tokenstore.TokenStore,
) TokenTransformer {
	return &tokenTransformer{
		log:        log,
		tokenStore: tokenStore,
	}
}

type tokenTransformer struct {
	log        logger.Logger
	tokenStore tokenstore.TokenStore
}

// TransformTransaction enriches a transaction with token data
func (t *tokenTransformer) TransformTransaction(ctx context.Context, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewInvalidInputError("Transaction cannot be nil", "transaction", nil)
	}

	// Check if metadata contains token_address
	tokenAddress, hasTokenAddress := tx.Metadata.GetString(types.ERC20TokenAddressMetadataKey)
	if !hasTokenAddress {
		// Also check for MultiSig token addresses
		tokenAddress, hasTokenAddress = tx.Metadata.GetString(types.MultiSigTokenAddressMetadataKey)
		if !hasTokenAddress {
			// No token address found, nothing to do
			return nil
		}
	}

	// Get token information from token store
	token, err := t.tokenStore.GetToken(ctx, tokenAddress)
	if err != nil {
		t.log.Warn("Failed to retrieve token information",
			logger.String("token_address", tokenAddress),
			logger.String("tx_hash", tx.Hash),
			logger.Error(err))
		// Continue without token info, don't return error
		return nil
	}

	// Update metadata with token information
	if token != nil {
		// Set symbol based on token type
		if tx.Metadata.Contains(types.ERC20TokenAddressMetadataKey) {
			if err := tx.Metadata.Set(types.ERC20TokenSymbolMetadataKey, token.Symbol); err != nil {
				t.log.Warn("Failed to set token symbol in metadata",
					logger.String("token_address", tokenAddress),
					logger.Error(err))
			}
			if err := tx.Metadata.Set(types.ERC20TokenDecimalsMetadataKey, token.Decimals); err != nil {
				t.log.Warn("Failed to set token decimals in metadata",
					logger.String("token_address", tokenAddress),
					logger.Error(err))
			}
		} else if tx.Metadata.Contains(types.MultiSigTokenAddressMetadataKey) {
			if err := tx.Metadata.Set(types.MultiSigTokenSymbolMetadataKey, token.Symbol); err != nil {
				t.log.Warn("Failed to set token symbol in metadata",
					logger.String("token_address", tokenAddress),
					logger.Error(err))
			}
			if err := tx.Metadata.Set(types.MultiSigTokenDecimalsMetadataKey, token.Decimals); err != nil {
				t.log.Warn("Failed to set token decimals in metadata",
					logger.String("token_address", tokenAddress),
					logger.Error(err))
			}
		}

		t.log.Debug("Enriched transaction with token information",
			logger.String("tx_hash", tx.Hash),
			logger.String("token_address", tokenAddress),
			logger.String("token_symbol", token.Symbol))
	}

	return nil
}
