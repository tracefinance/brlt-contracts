package transaction

import (
	"context"

	"vault0/internal/core/abiutils"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// evmDecoder implements the Decoder interface for EVM-based transactions.
type evmDecoder struct {
	tokenStore tokenstore.TokenStore
	logger     logger.Logger
	abiUtils   abiutils.ABIUtils
}

// NewEvmDecoder creates a new instance of the EVM transaction decoder.
func NewEvmDecoder(tokenStore tokenstore.TokenStore, log logger.Logger, abiUtils abiutils.ABIUtils) Decoder {
	return &evmDecoder{
		tokenStore: tokenStore,
		logger:     log.With(logger.String("component", "transaction_decoder")),
		abiUtils:   abiUtils,
	}
}

// decodeTransactionFromMetadata handles decoding transactions with a specific type
// and returns the appropriate CoreTransaction
func decodeTransactionFromMetadata(tx *types.Transaction) (types.CoreTransaction, error) {
	switch tx.Type {
	case types.TransactionTypeNative:
		return tx, nil
	case types.TransactionTypeDeploy:
		return tx, nil
	case types.TransactionTypeERC20Transfer:
		return decodeERC20Transfer(tx)
	case types.TransactionTypeMultiSigWithdrawalRequest:
		return decodeMultiSigWithdrawalRequest(tx)
	case types.TransactionTypeMultiSigSignWithdrawal:
		return decodeMultiSigSignWithdrawal(tx)
	case types.TransactionTypeMultiSigExecuteWithdrawal:
		return decodeMultiSigExecuteWithdrawal(tx)
	case types.TransactionTypeMultiSigAddSupportedToken:
		return decodeMultiSigAddSupportedToken(tx)
	case types.TransactionTypeMultiSigRemoveSupportedToken:
		return decodeMultiSigRemoveSupportedToken(tx)
	case types.TransactionTypeMultiSigRecoveryRequest:
		return decodeMultiSigRecoveryRequest(tx)
	case types.TransactionTypeMultiSigCancelRecovery:
		return decodeMultiSigCancelRecovery(tx)
	case types.TransactionTypeMultiSigExecuteRecovery:
		return decodeMultiSigExecuteRecovery(tx)
	case types.TransactionTypeMultiSigProposeRecoveryAddressChange:
		return decodeMultiSigProposeRecoveryAddressChange(tx)
	case types.TransactionTypeMultiSigSignRecoveryAddressChange:
		return decodeMultiSigSignRecoveryAddressChange(tx)
	default:
		return tx, nil
	}
}

// DecodeTransaction implements Mapper. It attempts to map a generic transaction
func (m *evmDecoder) DecodeTransaction(ctx context.Context, tx *types.Transaction) (types.CoreTransaction, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != "" && tx.Type != types.TransactionTypeContractCall {
		m.logger.Debug("Transaction already has a specific type, returning as is",
			logger.String("tx_hash", tx.Hash),
			logger.String("type", string(tx.Type)))

		result, err := decodeTransactionFromMetadata(tx)

		if tx.Type != result.GetType() {
			m.logger.Warn("Transaction has unrecognized specific type",
				logger.String("tx_hash", tx.Hash),
				logger.String("type", string(tx.Type)))
		}

		return result, err
	}

	m.logger.Debug("Transaction type is generic, attempting to parse data",
		logger.String("tx_hash", tx.Hash),
		logger.String("type", string(tx.Type)))

	if tx.BaseTransaction.To == "" || len(tx.Data) < 4 {
		m.logger.Debug("Transaction not eligible for parsing (missing To address or short data)",
			logger.String("tx_hash", tx.Hash))
		return tx, nil
	}

	txCopy := tx.Copy()

	parsedAsERC20, errERC20 := parseAndPopulateERC20Metadata(ctx, txCopy, m.abiUtils, m.tokenStore)
	if errERC20 != nil {
		m.logger.Error("Error attempting to parse as ERC20 transfer, proceeding to check MultiSig",
			logger.String("tx_hash", txCopy.Hash),
			logger.Error(errERC20),
		)
	} else if parsedAsERC20 {
		m.logger.Debug("Successfully parsed as ERC20Transfer", logger.String("tx_hash", txCopy.Hash))
		return decodeERC20Transfer(txCopy)
	}

	parsedAsMultiSig, errMultiSig := parseAndPopulateMultiSigMetadata(ctx, txCopy, m.logger, m.abiUtils, m.tokenStore)
	if errMultiSig != nil {
		m.logger.Error("Error attempting to parse as MultiSig",
			logger.String("tx_hash", txCopy.Hash),
			logger.Error(errMultiSig),
		)
		return tx, nil
	} else if parsedAsMultiSig {
		m.logger.Debug("Successfully parsed as a MultiSig transaction", logger.String("tx_hash", txCopy.Hash), logger.String("new_type", string(txCopy.Type)))

		result, err := decodeTransactionFromMetadata(txCopy)

		if txCopy.Type != result.GetType() {
			m.logger.Error("MultiSig parsing succeeded but resulted in an unexpected type",
				logger.String("tx_hash", txCopy.Hash),
				logger.String("type", string(txCopy.Type)),
			)
			return tx, nil
		}

		return result, err
	}

	m.logger.Debug("Transaction data did not match known ERC20 or MultiSig patterns",
		logger.String("tx_hash", tx.Hash))

	return tx, nil
}
