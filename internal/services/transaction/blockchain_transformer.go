package transaction

import (
	"context"
	"vault0/internal/core/blockchain"
	"vault0/internal/core/transaction"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// BlockchainTransformer defines methods for enriching transactions with blockchain data
type BlockchainTransformer interface {
	// TransformTransaction enriches a transaction with blockchain data
	TransformTransaction(ctx context.Context, tx *types.Transaction) error
}

// NewBlockchainTransformer creates a new blockchain transformer
func NewBlockchainTransformer(
	log logger.Logger,
	blockchainFactory blockchain.Factory,
	transactionFactory transaction.Factory,
) BlockchainTransformer {
	return &blockchainTransformer{
		log:                log,
		blockchainFactory:  blockchainFactory,
		transactionFactory: transactionFactory,
	}
}

type blockchainTransformer struct {
	log                logger.Logger
	blockchainFactory  blockchain.Factory
	transactionFactory transaction.Factory
}

// TransformTransaction enriches a transaction with blockchain data
func (t *blockchainTransformer) TransformTransaction(ctx context.Context, tx *types.Transaction) error {
	if tx == nil {
		return errors.NewInvalidInputError("Transaction cannot be nil", "transaction", nil)
	}

	// Get blockchain client for this chain
	client, err := t.blockchainFactory.NewClient(tx.ChainType)
	if err != nil {
		return errors.NewOperationFailedError("get blockchain client", err)
	}

	decoder, err := t.transactionFactory.NewDecoder(tx.ChainType)
	if err != nil {
		return errors.NewOperationFailedError("create transaction decoder", err)
	}

	// Get transaction receipt for additional data
	receipt, err := client.GetTransactionReceipt(ctx, tx.Hash)
	if err != nil {
		return errors.NewOperationFailedError("get transaction receipt", err)
	}

	// Enrich transaction with receipt data
	if receipt != nil {
		// Convert uint64 status to TransactionStatus
		if receipt.Status == 1 {
			tx.Status = types.TransactionStatusSuccess
		} else {
			tx.Status = types.TransactionStatusFailed
		}
		tx.GasUsed = receipt.GasUsed
	}

	// Load transaction input data for contract calls with empty data
	if tx.Type == types.TransactionTypeContractCall && len(tx.Data) == 0 {
		t.log.Debug("Loading transaction input data for contract call",
			logger.String("tx_hash", tx.Hash))

		// Get full transaction data from client
		fullTx, err := client.GetTransaction(ctx, tx.Hash)
		if err != nil {
			t.log.Warn("Failed to load transaction input data",
				logger.String("tx_hash", tx.Hash),
				logger.Error(err))
			// Continue with existing data, don't return error
		} else if fullTx != nil && fullTx.Data != nil {
			tx.Data = fullTx.Data
			t.log.Debug("Loaded transaction input data",
				logger.String("tx_hash", tx.Hash),
				logger.Int("data_length", len(tx.Data)))
		}
	}

	typedTx, err := decoder.DecodeTransaction(ctx, tx)
	if err != nil {
		return err
	}

	tx.Type = typedTx.GetType()
	tx.Metadata = typedTx.GetMetadata()

	return nil
}
