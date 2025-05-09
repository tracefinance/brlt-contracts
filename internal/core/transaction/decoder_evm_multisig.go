package transaction

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/crypto"

	"vault0/internal/core/abiutils"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Precomputed MultiSig method IDs
var (
	multiSigRequestWithdrawalMethodID            []byte
	multiSigSignWithdrawalMethodID               []byte
	multiSigExecuteWithdrawalMethodID            []byte
	multiSigAddSupportedTokenMethodID            []byte
	multiSigRemoveSupportedTokenMethodID         []byte
	multiSigRequestRecoveryMethodID              []byte
	multiSigCancelRecoveryMethodID               []byte
	multiSigExecuteRecoveryMethodID              []byte
	multiSigProposeRecoveryAddressChangeMethodID []byte
	multiSigSignRecoveryAddressChangeMethodID    []byte
)

func init() {
	multiSigRequestWithdrawalMethodID = crypto.Keccak256([]byte(types.MultiSigRequestWithdrawalMethod))[:4]
	multiSigSignWithdrawalMethodID = crypto.Keccak256([]byte(types.MultiSigSignWithdrawalMethod))[:4]
	multiSigExecuteWithdrawalMethodID = crypto.Keccak256([]byte(types.MultiSigExecuteWithdrawalMethod))[:4]
	multiSigAddSupportedTokenMethodID = crypto.Keccak256([]byte(types.MultiSigAddSupportedTokenMethod))[:4]
	multiSigRemoveSupportedTokenMethodID = crypto.Keccak256([]byte(types.MultiSigRemoveSupportedTokenMethod))[:4]
	multiSigRequestRecoveryMethodID = crypto.Keccak256([]byte(types.MultiSigRequestRecoveryMethod))[:4]
	multiSigCancelRecoveryMethodID = crypto.Keccak256([]byte(types.MultiSigCancelRecoveryMethod))[:4]
	multiSigExecuteRecoveryMethodID = crypto.Keccak256([]byte(types.MultiSigExecuteRecoveryMethod))[:4]
	multiSigProposeRecoveryAddressChangeMethodID = crypto.Keccak256([]byte(types.MultiSigProposeRecoveryAddressChangeMethod))[:4]
	multiSigSignRecoveryAddressChangeMethodID = crypto.Keccak256([]byte(types.MultiSigSignRecoveryAddressChangeMethod))[:4]
}

// getMethodNameFromSignature extracts the method name (e.g., "transfer")
// from a full signature string (e.g., "transfer(address,uint256)").
func getMethodNameFromSignature(fullSignature string) string {
	if idx := strings.Index(fullSignature, "("); idx != -1 {
		return fullSignature[:idx]
	}
	// Fallback, though valid Solidity signatures should contain '('.
	return fullSignature
}

func createMultiSigWithdrawalRequestFromMetadata(tx *types.Transaction) (*types.MultiSigWithdrawalRequest, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	token, ok := tx.Metadata.GetString(types.MultiSigTokenAddressMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.MultiSigTokenAddressMetadataKey)
	}
	recipient, ok := tx.Metadata.GetString(types.MultiSigRecipientMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.MultiSigRecipientMetadataKey)
	}
	amount, ok := tx.Metadata.GetBigInt(types.MultiSigAmountMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing or invalid metadata: "+types.MultiSigAmountMetadataKey)
	}
	withdrawalNonce, ok := tx.Metadata.GetUint64(types.MultiSigWithdrawalNonceMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing or invalid metadata: "+types.MultiSigWithdrawalNonceMetadataKey)
	}

	tokenSymbol, _ := tx.Metadata.GetString(types.MultiSigTokenSymbolMetadataKey)
	tokenDecimals, _ := tx.Metadata.GetUint8(types.MultiSigTokenDecimalsMetadataKey)

	msTx := &types.MultiSigWithdrawalRequest{
		Transaction:     *tx.Copy(),
		TokenAddress:    token,
		TokenSymbol:     tokenSymbol,
		TokenDecimals:   tokenDecimals,
		Amount:          amount.ToBigInt(),
		Recipient:       recipient,
		WithdrawalNonce: withdrawalNonce,
	}
	msTx.Type = types.TransactionTypeMultiSigWithdrawalRequest

	return msTx, nil
}

func createMultiSigSignWithdrawalFromMetadata(tx *types.Transaction) (*types.MultiSigSignWithdrawal, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	requestID, ok := tx.Metadata.GetBytes32(types.MultiSigRequestIDMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing or invalid metadata: "+types.MultiSigRequestIDMetadataKey)
	}

	msTx := &types.MultiSigSignWithdrawal{
		Transaction: *tx.Copy(),
		RequestID:   requestID,
	}
	msTx.Type = types.TransactionTypeMultiSigSignWithdrawal

	return msTx, nil
}

func createMultiSigExecuteWithdrawalFromMetadata(tx *types.Transaction) (*types.MultiSigExecuteWithdrawal, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	requestID, ok := tx.Metadata.GetBytes32(types.MultiSigRequestIDMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing or invalid metadata: "+types.MultiSigRequestIDMetadataKey)
	}

	msTx := &types.MultiSigExecuteWithdrawal{
		Transaction: *tx.Copy(),
		RequestID:   requestID,
	}
	msTx.Type = types.TransactionTypeMultiSigExecuteWithdrawal

	return msTx, nil
}

func createMultiSigAddSupportedTokenFromMetadata(tx *types.Transaction) (*types.MultiSigAddSupportedToken, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	token, ok := tx.Metadata.GetString(types.MultiSigTokenAddressMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.MultiSigTokenAddressMetadataKey)
	}

	msTx := &types.MultiSigAddSupportedToken{
		Transaction:  *tx.Copy(),
		TokenAddress: token,
	}
	msTx.Type = types.TransactionTypeMultiSigAddSupportedToken

	return msTx, nil
}

func createMultiSigRemoveSupportedTokenFromMetadata(tx *types.Transaction) (*types.MultiSigRemoveSupportedToken, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	token, ok := tx.Metadata.GetString(types.MultiSigTokenAddressMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.MultiSigTokenAddressMetadataKey)
	}

	msTx := &types.MultiSigRemoveSupportedToken{
		Transaction:  *tx.Copy(),
		TokenAddress: token,
	}
	msTx.Type = types.TransactionTypeMultiSigRemoveSupportedToken

	return msTx, nil
}

func createMultiSigRecoveryRequestFromMetadata(tx *types.Transaction) (*types.MultiSigRecoveryRequest, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	msTx := &types.MultiSigRecoveryRequest{
		Transaction: *tx.Copy(),
	}
	msTx.Type = types.TransactionTypeMultiSigRecoveryRequest

	return msTx, nil
}

func createMultiSigCancelRecoveryFromMetadata(tx *types.Transaction) (*types.MultiSigCancelRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	msTx := &types.MultiSigCancelRecovery{
		Transaction: *tx.Copy(),
	}
	msTx.Type = types.TransactionTypeMultiSigCancelRecovery

	return msTx, nil
}

func createMultiSigExecuteRecoveryFromMetadata(tx *types.Transaction) (*types.MultiSigExecuteRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	msTx := &types.MultiSigExecuteRecovery{
		Transaction: *tx.Copy(),
	}
	msTx.Type = types.TransactionTypeMultiSigExecuteRecovery

	return msTx, nil
}

func createMultiSigProposeRecoveryAddressChangeFromMetadata(tx *types.Transaction) (*types.MultiSigProposeRecoveryAddressChange, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	newAddr, ok := tx.Metadata.GetString(types.MultiSigNewRecoveryAddressMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.MultiSigNewRecoveryAddressMetadataKey)
	}

	msTx := &types.MultiSigProposeRecoveryAddressChange{
		Transaction:        *tx.Copy(),
		NewRecoveryAddress: newAddr,
	}
	msTx.Type = types.TransactionTypeMultiSigProposeRecoveryAddressChange

	return msTx, nil
}

func createMultiSigSignRecoveryAddressChangeFromMetadata(tx *types.Transaction) (*types.MultiSigSignRecoveryAddressChange, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	proposalID, ok := tx.Metadata.GetBytes32(types.MultiSigProposalIDMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing or invalid metadata: "+types.MultiSigProposalIDMetadataKey)
	}

	msTx := &types.MultiSigSignRecoveryAddressChange{
		Transaction: *tx.Copy(),
		ProposalID:  proposalID,
	}
	msTx.Type = types.TransactionTypeMultiSigSignRecoveryAddressChange

	return msTx, nil
}

// parseAndPopulateMultiSigMetadata attempts to parse tx data as a known MultiSig interaction
func parseAndPopulateMultiSigMetadata(ctx context.Context, tx *types.Transaction, log logger.Logger, abiTools abiutils.ABIUtils, ts tokenstore.TokenStore) (bool, error) {
	if tx == nil {
		return false, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}
	if tx.BaseTransaction.To == "" {
		log.Debug("Skipping MultiSig parse: tx.To is empty", logger.String("tx_hash", tx.Hash))
		return false, nil
	}
	if len(tx.Data) < 4 {
		log.Debug("Skipping MultiSig parse: tx.Data too short for method ID",
			logger.String("tx_hash", tx.Hash),
			logger.Int("data_len", len(tx.Data)))
		return false, nil
	}

	multiSigABIString, err := abiTools.LoadABIByName(ctx, abiutils.ABITypeMultiSig)
	if err != nil {
		log.Error("Failed to load MultiSig ABI for parsing",
			logger.String("tx_hash", tx.Hash),
			logger.Error(err))
		return false, fmt.Errorf("failed to load MultiSig ABI: %w", err)
	}

	methodID := abiTools.ExtractMethodID(tx.Data)

	var parsedArgs map[string]any
	var currentMethodNameForParsing string
	var specificTxType types.TransactionType
	metadataToSet := make(map[string]any)
	var parsingErr error

	switch {
	case bytes.Equal(methodID, multiSigRequestWithdrawalMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigRequestWithdrawalMethod))
		parsedArgs, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument names ("token", "amount", "to") must match the MultiSig contract's ABI definition.
		tokenAddr, errExtract := abiTools.GetAddressFromArgs(parsedArgs, "token")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}
		amountBigInt, errExtract := abiTools.GetBigIntFromArgs(parsedArgs, "amount")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}
		recipientAddr, errExtract := abiTools.GetAddressFromArgs(parsedArgs, "to")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		// Fetch token details from tokenStore
		tokenInfo, tokenErr := ts.GetToken(ctx, tokenAddr.String())
		if tokenErr != nil {
			log.Warn("Failed to resolve token details for MultiSig withdrawal request",
				logger.String("tx_hash", tx.Hash),
				logger.String("token_address", tokenAddr.String()),
				logger.Error(tokenErr),
			)
			// Use fallback details if token is not registered or error occurs
			tokenInfo = &types.Token{Symbol: "UNKNOWN", Decimals: 0} // Default to 0 for decimals as it's safer
		}

		metadataToSet[types.MultiSigTokenAddressMetadataKey] = tokenAddr.String()
		metadataToSet[types.MultiSigAmountMetadataKey] = amountBigInt
		metadataToSet[types.MultiSigRecipientMetadataKey] = recipientAddr.String()
		// Add symbol and decimals to metadata
		metadataToSet[types.MultiSigTokenSymbolMetadataKey] = tokenInfo.Symbol
		metadataToSet[types.MultiSigTokenDecimalsMetadataKey] = tokenInfo.Decimals
		specificTxType = types.TransactionTypeMultiSigWithdrawalRequest

	case bytes.Equal(methodID, multiSigSignWithdrawalMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigSignWithdrawalMethod))
		parsedArgs, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "requestId" must match the ABI.
		requestIDBytes, errExtract := abiTools.GetBytes32FromArgs(parsedArgs, "requestId")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigRequestIDMetadataKey] = requestIDBytes
		specificTxType = types.TransactionTypeMultiSigSignWithdrawal

	case bytes.Equal(methodID, multiSigExecuteWithdrawalMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigExecuteWithdrawalMethod))
		parsedArgs, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "requestId" must match the ABI.
		requestIDBytes, errExtract := abiTools.GetBytes32FromArgs(parsedArgs, "requestId")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigRequestIDMetadataKey] = requestIDBytes
		specificTxType = types.TransactionTypeMultiSigExecuteWithdrawal

	case bytes.Equal(methodID, multiSigAddSupportedTokenMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigAddSupportedTokenMethod))
		parsedArgs, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "token" must match the ABI.
		tokenAddr, errExtract := abiTools.GetAddressFromArgs(parsedArgs, "token")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigTokenAddressMetadataKey] = tokenAddr.String()
		specificTxType = types.TransactionTypeMultiSigAddSupportedToken

	case bytes.Equal(methodID, multiSigRemoveSupportedTokenMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigRemoveSupportedTokenMethod))
		parsedArgs, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "token" must match the ABI.
		tokenAddr, errExtract := abiTools.GetAddressFromArgs(parsedArgs, "token")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigTokenAddressMetadataKey] = tokenAddr.String()
		specificTxType = types.TransactionTypeMultiSigRemoveSupportedToken

	case bytes.Equal(methodID, multiSigRequestRecoveryMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigRequestRecoveryMethod))
		_, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data) // No args to extract
		if parsingErr != nil {
			break
		}
		specificTxType = types.TransactionTypeMultiSigRecoveryRequest

	case bytes.Equal(methodID, multiSigCancelRecoveryMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigCancelRecoveryMethod))
		_, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data) // No args
		if parsingErr != nil {
			break
		}
		specificTxType = types.TransactionTypeMultiSigCancelRecovery

	case bytes.Equal(methodID, multiSigExecuteRecoveryMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigExecuteRecoveryMethod))
		_, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data) // No args
		if parsingErr != nil {
			break
		}
		specificTxType = types.TransactionTypeMultiSigExecuteRecovery

	case bytes.Equal(methodID, multiSigProposeRecoveryAddressChangeMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigProposeRecoveryAddressChangeMethod))
		parsedArgs, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "newRecoveryAddress" must match the ABI.
		newAddr, errExtract := abiTools.GetAddressFromArgs(parsedArgs, "newRecoveryAddress")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigNewRecoveryAddressMetadataKey] = newAddr.String()
		specificTxType = types.TransactionTypeMultiSigProposeRecoveryAddressChange

	case bytes.Equal(methodID, multiSigSignRecoveryAddressChangeMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigSignRecoveryAddressChangeMethod))
		parsedArgs, parsingErr = abiTools.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "proposalId" must match the ABI.
		proposalIDBytes, errExtract := abiTools.GetBytes32FromArgs(parsedArgs, "proposalId")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigProposalIDMetadataKey] = proposalIDBytes
		specificTxType = types.TransactionTypeMultiSigSignRecoveryAddressChange

	default:
		log.Debug("Transaction data method ID does not match any known MultiSig methods for detailed parsing.",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_id_hex", hex.EncodeToString(methodID)))
		return false, nil // Not an error, just not a MultiSig call we're decoding this way.
	}

	if parsingErr != nil {
		log.Warn("Failed to parse MultiSig transaction input or extract arguments for a recognized method.",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_name_attempted", currentMethodNameForParsing),
			logger.String("method_id_hex", hex.EncodeToString(methodID)),
			logger.Error(parsingErr),
		)
		// Return error because we identified a method but couldn't parse its details.
		return false, fmt.Errorf("failed to parse args for known MultiSig method %s (ID: %s): %w", currentMethodNameForParsing, hex.EncodeToString(methodID), parsingErr)
	}

	if specificTxType == "" {
		log.Error("Internal logic error: MultiSig method processed but specificTxType not set, despite no parsing error.",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_name_parsed", currentMethodNameForParsing),
			logger.String("method_id_hex", hex.EncodeToString(methodID)))
		return false, fmt.Errorf("internal logic error determining MultiSig tx type for method ID %s", hex.EncodeToString(methodID))
	}

	if tx.Metadata == nil {
		tx.Metadata = make(types.TxMetadata)
	}

	err = tx.Metadata.SetAll(metadataToSet)
	if err != nil {
		log.Error("Failed to set MultiSig metadata map after parsing",
			logger.String("tx_hash", tx.Hash),
			logger.String("type", string(specificTxType)),
			logger.Error(err))
		return false, fmt.Errorf("failed to set MultiSig metadata values: %w", err)
	}

	tx.Type = specificTxType

	log.Info("Successfully parsed and populated MultiSig transaction metadata.",
		logger.String("tx_hash", tx.Hash),
		logger.String("parsed_type", string(specificTxType)))
	return true, nil
}

func decodeMultiSigWithdrawalRequest(tx *types.Transaction) (*types.MultiSigWithdrawalRequest, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigWithdrawalRequest {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigWithdrawalRequest))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigWithdrawalRequest")
	}

	return createMultiSigWithdrawalRequestFromMetadata(tx)
}

func decodeMultiSigSignWithdrawal(tx *types.Transaction) (*types.MultiSigSignWithdrawal, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigSignWithdrawal {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigSignWithdrawal))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigSignWithdrawal")
	}

	return createMultiSigSignWithdrawalFromMetadata(tx)
}

func decodeMultiSigExecuteWithdrawal(tx *types.Transaction) (*types.MultiSigExecuteWithdrawal, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigExecuteWithdrawal {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigExecuteWithdrawal))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigExecuteWithdrawal")
	}

	return createMultiSigExecuteWithdrawalFromMetadata(tx)
}

func decodeMultiSigAddSupportedToken(tx *types.Transaction) (*types.MultiSigAddSupportedToken, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigAddSupportedToken {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigAddSupportedToken))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigAddSupportedToken")
	}

	return createMultiSigAddSupportedTokenFromMetadata(tx)
}

func decodeMultiSigRemoveSupportedToken(tx *types.Transaction) (*types.MultiSigRemoveSupportedToken, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigRemoveSupportedToken {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigRemoveSupportedToken))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigRemoveSupportedToken")
	}

	return createMultiSigRemoveSupportedTokenFromMetadata(tx)
}

func decodeMultiSigRecoveryRequest(tx *types.Transaction) (*types.MultiSigRecoveryRequest, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigRecoveryRequest {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigRecoveryRequest))
	}

	return createMultiSigRecoveryRequestFromMetadata(tx)
}

func decodeMultiSigCancelRecovery(tx *types.Transaction) (*types.MultiSigCancelRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigCancelRecovery {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigCancelRecovery))
	}

	return createMultiSigCancelRecoveryFromMetadata(tx)
}

func decodeMultiSigExecuteRecovery(tx *types.Transaction) (*types.MultiSigExecuteRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigExecuteRecovery {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigExecuteRecovery))
	}

	return createMultiSigExecuteRecoveryFromMetadata(tx)
}

func decodeMultiSigProposeRecoveryAddressChange(tx *types.Transaction) (*types.MultiSigProposeRecoveryAddressChange, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigProposeRecoveryAddressChange {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigProposeRecoveryAddressChange))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigProposeRecoveryAddressChange")
	}

	return createMultiSigProposeRecoveryAddressChangeFromMetadata(tx)
}

func decodeMultiSigSignRecoveryAddressChange(tx *types.Transaction) (*types.MultiSigSignRecoveryAddressChange, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigSignRecoveryAddressChange {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigSignRecoveryAddressChange))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigSignRecoveryAddressChange")
	}

	return createMultiSigSignRecoveryAddressChangeFromMetadata(tx)
}
