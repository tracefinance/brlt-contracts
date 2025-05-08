package transaction

import (
	"bytes"
	"context"
	"encoding/hex" // For logging method ID
	"fmt"
	"strings" // For getMethodNameFromSignature

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto" // Need crypto for Keccak256

	"vault0/internal/core/abiutils"
	"vault0/internal/core/tokenstore"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/types"
)

// Precompute ERC20 transfer method ID locally for dispatcher logic
var erc20TransferMethodID = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

// Precomputed MultiSig method IDs
var (
	multiSigRequestWithdrawalMethodID            []byte
	multiSigSignWithdrawalMethodID               []byte
	multiSigExecuteWithdrawalMethodID            []byte // Placeholder, assuming "executeWithdrawal(bytes32)"
	multiSigAddSupportedTokenMethodID            []byte
	multiSigRequestRecoveryMethodID              []byte
	multiSigCancelRecoveryMethodID               []byte
	multiSigExecuteRecoveryMethodID              []byte
	multiSigProposeRecoveryAddressChangeMethodID []byte
	multiSigSignRecoveryAddressChangeMethodID    []byte
)

// assumedMultiSigExecuteWithdrawalSignature is used if a method like "executeWithdrawal(bytes32)"
// is part of the MultiSig ABI for TransactionTypeMultiSigExecuteWithdrawal.
const assumedMultiSigExecuteWithdrawalSignature = "executeWithdrawal(bytes32)"

func init() {
	multiSigRequestWithdrawalMethodID = crypto.Keccak256([]byte(types.MultiSigRequestWithdrawalMethod))[:4]
	multiSigSignWithdrawalMethodID = crypto.Keccak256([]byte(types.MultiSigSignWithdrawalMethod))[:4]
	multiSigExecuteWithdrawalMethodID = crypto.Keccak256([]byte(assumedMultiSigExecuteWithdrawalSignature))[:4]
	multiSigAddSupportedTokenMethodID = crypto.Keccak256([]byte(types.MultiSigAddSupportedTokenMethod))[:4]
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
	// Log or handle error if this is critical.
	return fullSignature
}

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

// --- Metadata-Based Construction Helpers ---

// createERC20TransferFromMetadata constructs an ERC20Transfer from metadata.
func createERC20TransferFromMetadata(tx *types.Transaction) (*types.ERC20Transfer, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}
	tokenAddr, ok := tx.Metadata.GetString(types.ERC20TokenAddressMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.ERC20TokenAddressMetadataKey)
	}
	recipient, ok := tx.Metadata.GetString(types.ERC20RecipientMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.ERC20RecipientMetadataKey)
	}
	amountBigInt, ok := tx.Metadata.GetBigInt(types.ERC20AmountMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing or invalid metadata: "+types.ERC20AmountMetadataKey)
	}
	decimals, ok := tx.Metadata.GetUint8(types.ERC20TokenDecimalsMetadataKey)
	if !ok {
		decimals = 18 // Default to 18 decimals if not specified
	}
	tokenSymbol, _ := tx.Metadata.GetString(types.ERC20TokenSymbolMetadataKey)

	erc20Tx := &types.ERC20Transfer{
		Transaction:   *tx.Copy(),
		TokenAddress:  tokenAddr,
		TokenSymbol:   tokenSymbol, // Will be empty string if not present
		TokenDecimals: decimals,    // Will be 0 if not present
		Recipient:     recipient,
		Amount:        amountBigInt.ToBigInt(),
	}
	erc20Tx.Type = types.TransactionTypeERC20Transfer // Ensure type is correct

	return erc20Tx, nil
}

// createMultiSigWithdrawalRequestFromMetadata constructs a MultiSigWithdrawalRequest from metadata.
func createMultiSigWithdrawalRequestFromMetadata(tx *types.Transaction) (*types.MultiSigWithdrawalRequest, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	token, ok := tx.Metadata.GetString(types.MultiSigTokenMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.MultiSigTokenMetadataKey)
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

	msTx := &types.MultiSigWithdrawalRequest{
		Transaction:     *tx.Copy(),
		Token:           token,
		Amount:          amount.ToBigInt(),
		Recipient:       recipient,
		WithdrawalNonce: withdrawalNonce,
	}
	msTx.Type = types.TransactionTypeMultiSigWithdrawalRequest

	return msTx, nil
}

// createMultiSigSignWithdrawalFromMetadata constructs a MultiSigSignWithdrawal from metadata.
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

// createMultiSigExecuteWithdrawalFromMetadata constructs a MultiSigExecuteWithdrawal from metadata.
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

// createMultiSigAddSupportedTokenFromMetadata constructs a MultiSigAddSupportedToken from metadata.
func createMultiSigAddSupportedTokenFromMetadata(tx *types.Transaction) (*types.MultiSigAddSupportedToken, error) {
	if tx == nil || tx.Metadata == nil {
		return nil, errors.NewInvalidParameterError("transaction or metadata cannot be nil", "tx")
	}

	token, ok := tx.Metadata.GetString(types.MultiSigTokenMetadataKey)
	if !ok {
		return nil, errors.NewMappingError(tx.Hash, "missing metadata: "+types.MultiSigTokenMetadataKey)
	}

	msTx := &types.MultiSigAddSupportedToken{
		Transaction: *tx.Copy(),
		Token:       token,
	}
	msTx.Type = types.TransactionTypeMultiSigAddSupportedToken

	return msTx, nil
}

// createMultiSigRecoveryRequestFromMetadata constructs a MultiSigRecoveryRequest from metadata.
func createMultiSigRecoveryRequestFromMetadata(tx *types.Transaction) (*types.MultiSigRecoveryRequest, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	msTx := &types.MultiSigRecoveryRequest{
		Transaction: *tx.Copy(),
	}
	msTx.Type = types.TransactionTypeMultiSigRecoveryRequest
	// No specific metadata fields required for this type

	return msTx, nil
}

// createMultiSigCancelRecoveryFromMetadata constructs a MultiSigCancelRecovery from metadata.
func createMultiSigCancelRecoveryFromMetadata(tx *types.Transaction) (*types.MultiSigCancelRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	msTx := &types.MultiSigCancelRecovery{
		Transaction: *tx.Copy(),
	}
	msTx.Type = types.TransactionTypeMultiSigCancelRecovery
	// No specific metadata fields required

	return msTx, nil
}

// createMultiSigExecuteRecoveryFromMetadata constructs a MultiSigExecuteRecovery from metadata.
func createMultiSigExecuteRecoveryFromMetadata(tx *types.Transaction) (*types.MultiSigExecuteRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	msTx := &types.MultiSigExecuteRecovery{
		Transaction: *tx.Copy(),
	}
	msTx.Type = types.TransactionTypeMultiSigExecuteRecovery
	// No specific metadata fields required

	return msTx, nil
}

// createMultiSigProposeRecoveryAddressChangeFromMetadata constructs a MultiSigProposeRecoveryAddressChange from metadata.
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

// createMultiSigSignRecoveryAddressChangeFromMetadata constructs a MultiSigSignRecoveryAddressChange from metadata.
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

// parseAndPopulateERC20Metadata attempts to parse tx data as an ERC20 transfer
// and updates tx.Metadata and tx.Type if successful.
// Returns true if parsing was successful and metadata was populated.
func (m *evmDecoder) parseAndPopulateERC20Metadata(ctx context.Context, tx *types.Transaction) (bool, error) {
	if tx == nil {
		return false, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	// Basic validation for parsing
	if tx.BaseTransaction.To == "" {
		m.logger.Debug("Skipping ERC20 parse: tx.To is empty", logger.String("tx_hash", tx.Hash))
		return false, nil
	}
	if len(tx.Data) < 4 {
		m.logger.Debug("Skipping ERC20 parse: tx.Data too short",
			logger.String("tx_hash", tx.Hash),
			logger.Int("data_len", len(tx.Data)))
		return false, nil
	}

	// Extract the method ID using abiutils
	methodID := m.abiUtils.ExtractMethodID(tx.Data)

	// Optimization: Check against precomputed standard ERC20 transfer ID.
	if !bytes.Equal(methodID, erc20TransferMethodID) {
		// Method ID doesn't match, this isn't an ERC20 transfer. Not an error.
		return false, nil
	}

	m.logger.Debug("Method ID matches ERC20 transfer, attempting parse", logger.String("tx_hash", tx.Hash))

	// Load ABI specifically for the target contract address.
	contractAddrEth := common.HexToAddress(tx.BaseTransaction.To)
	// Convert to types.Address
	contractAddr, err := types.NewAddress(tx.BaseTransaction.ChainType, contractAddrEth.Hex())
	if err != nil {
		m.logger.Warn("Failed to convert address for ERC20 transfer parse",
			logger.String("address", contractAddrEth.Hex()),
			logger.String("tx_hash", tx.Hash),
			logger.Error(err),
		)
		return false, fmt.Errorf("failed to convert address %s: %w", contractAddrEth.Hex(), err)
	}

	erc20ABI, err := m.abiUtils.LoadABIByAddress(ctx, *contractAddr)
	if err != nil {
		m.logger.Warn("Failed to load ABI by address for potential ERC20 transfer parse",
			logger.String("address", contractAddr.String()),
			logger.String("tx_hash", tx.Hash),
			logger.Error(err),
		)
		// Cannot proceed without ABI, but don't return error, just indicate parsing failed.
		// The caller (ToTypedTransaction) might handle this gracefully.
		return false, fmt.Errorf("failed to load ABI for address %s: %w", contractAddr.String(), err) // Return error to indicate ABI load failure
	}

	// Parse the input data with the "transfer" method name
	parsedArgs, err := m.abiUtils.ParseContractInput(erc20ABI, "transfer", tx.Data)
	if err != nil {
		m.logger.Warn("Failed to parse ERC20 transfer input data",
			logger.String("tx_hash", tx.Hash),
			logger.String("address", contractAddr.String()),
			logger.Error(err),
		)
		// Parsing failed, return error.
		return false, fmt.Errorf("failed to parse ERC20 transfer input: %w", err)
	}

	// Extract recipient address ('recipient') and amount ('amount').
	recipientAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, types.ERC20RecipientMetadataKey)
	if err != nil {
		return false, fmt.Errorf("failed to get recipient address ('recipient') from parsed args: %w", err)
	}

	amountBigIntParsed, err := m.abiUtils.GetBigIntFromArgs(parsedArgs, types.ERC20AmountMetadataKey)
	if err != nil {
		return false, fmt.Errorf("failed to get transfer amount ('amount') from parsed args: %w", err)
	}

	// Resolve token details (symbol, decimals) using the token store.
	// Use address from tx.To which is the token contract address.
	tokenInfo, err := m.tokenStore.GetToken(ctx, tx.BaseTransaction.To)
	if err != nil {
		m.logger.Warn("Failed to resolve token details via token store for parsed ERC20 transfer",
			logger.String("chain", string(tx.BaseTransaction.ChainType)),
			logger.String("token_address", tx.BaseTransaction.To),
			logger.String("tx_hash", tx.Hash),
			logger.Error(err),
		)
		// Use fallback details if token is not registered in the store.
		tokenInfo = &types.Token{Address: tx.BaseTransaction.To, Symbol: "UNKNOWN", Decimals: 0}
	}

	// Ensure metadata map exists
	if tx.Metadata == nil {
		tx.Metadata = make(types.TxMetadata)
	}

	// Populate metadata
	err = tx.Metadata.SetAll(map[string]any{
		types.ERC20TokenAddressMetadataKey:  tokenInfo.Address, // Same as tx.To
		types.ERC20TokenSymbolMetadataKey:   tokenInfo.Symbol,
		types.ERC20TokenDecimalsMetadataKey: tokenInfo.Decimals,
		types.ERC20RecipientMetadataKey:     recipientAddr.String(),
		types.ERC20AmountMetadataKey:        amountBigIntParsed.ToBigInt(), // Convert from abiutils type
		types.TransactionTypeMetadaKey:      string(types.TransactionTypeERC20Transfer),
	})
	if err != nil {
		// This should ideally not happen if types are correct, but handle defensively
		m.logger.Error("Failed to set ERC20 metadata after parsing", logger.String("tx_hash", tx.Hash), logger.Error(err))
		return false, fmt.Errorf("failed to set ERC20 metadata: %w", err)
	}

	// Update transaction type
	tx.Type = types.TransactionTypeERC20Transfer

	m.logger.Info("Successfully parsed and populated ERC20 transfer metadata", logger.String("tx_hash", tx.Hash))
	return true, nil // Parsing successful
}

// parseAndPopulateMultiSigMetadata attempts to parse tx data as a known MultiSig interaction
// and updates tx.Metadata and tx.Type if successful.
// Returns true if parsing was successful and metadata was populated.
func (m *evmDecoder) parseAndPopulateMultiSigMetadata(ctx context.Context, tx *types.Transaction) (bool, error) {
	if tx == nil {
		return false, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}
	if tx.BaseTransaction.To == "" {
		m.logger.Debug("Skipping MultiSig parse: tx.To is empty", logger.String("tx_hash", tx.Hash))
		return false, nil
	}
	if len(tx.Data) < 4 {
		m.logger.Debug("Skipping MultiSig parse: tx.Data too short for method ID",
			logger.String("tx_hash", tx.Hash),
			logger.Int("data_len", len(tx.Data)))
		return false, nil
	}

	multiSigABIString, err := m.abiUtils.LoadABIByName(ctx, abiutils.ABITypeMultiSig)
	if err != nil {
		m.logger.Error("Failed to load MultiSig ABI for parsing",
			logger.String("tx_hash", tx.Hash),
			logger.Error(err))
		// This is a critical failure for any attempt to parse as MultiSig.
		return false, fmt.Errorf("failed to load MultiSig ABI: %w", err)
	}

	methodID := m.abiUtils.ExtractMethodID(tx.Data)
	// ExtractMethodID returns nil if len(data) < 4, already checked.

	var parsedArgs map[string]any
	var currentMethodNameForParsing string
	var specificTxType types.TransactionType
	metadataToSet := make(map[string]any)
	var parsingErr error // To capture errors from ABI parsing or argument extraction

	switch {
	case bytes.Equal(methodID, multiSigRequestWithdrawalMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigRequestWithdrawalMethod))
		parsedArgs, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument names ("token", "amount", "to") must match the MultiSig contract's ABI definition.
		tokenAddr, errExtract := m.abiUtils.GetAddressFromArgs(parsedArgs, "token")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}
		amountBigInt, errExtract := m.abiUtils.GetBigIntFromArgs(parsedArgs, "amount")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}
		recipientAddr, errExtract := m.abiUtils.GetAddressFromArgs(parsedArgs, "to")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigTokenMetadataKey] = tokenAddr.String()
		metadataToSet[types.MultiSigAmountMetadataKey] = amountBigInt
		metadataToSet[types.MultiSigRecipientMetadataKey] = recipientAddr.String()
		specificTxType = types.TransactionTypeMultiSigWithdrawalRequest

	case bytes.Equal(methodID, multiSigSignWithdrawalMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigSignWithdrawalMethod))
		parsedArgs, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "requestId" must match the ABI.
		requestIDBytes, errExtract := m.abiUtils.GetBytes32FromArgs(parsedArgs, "requestId")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigRequestIDMetadataKey] = requestIDBytes
		specificTxType = types.TransactionTypeMultiSigSignWithdrawal

	case bytes.Equal(methodID, multiSigExecuteWithdrawalMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(assumedMultiSigExecuteWithdrawalSignature)
		parsedArgs, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "requestId" must match the ABI.
		requestIDBytes, errExtract := m.abiUtils.GetBytes32FromArgs(parsedArgs, "requestId")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigRequestIDMetadataKey] = requestIDBytes
		specificTxType = types.TransactionTypeMultiSigExecuteWithdrawal

	case bytes.Equal(methodID, multiSigAddSupportedTokenMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigAddSupportedTokenMethod))
		parsedArgs, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "token" must match the ABI.
		tokenAddr, errExtract := m.abiUtils.GetAddressFromArgs(parsedArgs, "token")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigTokenMetadataKey] = tokenAddr.String()
		specificTxType = types.TransactionTypeMultiSigAddSupportedToken

	case bytes.Equal(methodID, multiSigRequestRecoveryMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigRequestRecoveryMethod))
		_, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data) // No args to extract
		if parsingErr != nil {
			break
		}
		specificTxType = types.TransactionTypeMultiSigRecoveryRequest

	case bytes.Equal(methodID, multiSigCancelRecoveryMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigCancelRecoveryMethod))
		_, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data) // No args
		if parsingErr != nil {
			break
		}
		specificTxType = types.TransactionTypeMultiSigCancelRecovery

	case bytes.Equal(methodID, multiSigExecuteRecoveryMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigExecuteRecoveryMethod))
		_, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data) // No args
		if parsingErr != nil {
			break
		}
		specificTxType = types.TransactionTypeMultiSigExecuteRecovery

	case bytes.Equal(methodID, multiSigProposeRecoveryAddressChangeMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigProposeRecoveryAddressChangeMethod))
		parsedArgs, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "newRecoveryAddress" must match the ABI.
		newAddr, errExtract := m.abiUtils.GetAddressFromArgs(parsedArgs, "newRecoveryAddress")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigNewRecoveryAddressMetadataKey] = newAddr.String()
		specificTxType = types.TransactionTypeMultiSigProposeRecoveryAddressChange

	case bytes.Equal(methodID, multiSigSignRecoveryAddressChangeMethodID):
		currentMethodNameForParsing = getMethodNameFromSignature(string(types.MultiSigSignRecoveryAddressChangeMethod))
		parsedArgs, parsingErr = m.abiUtils.ParseContractInput(multiSigABIString, currentMethodNameForParsing, tx.Data)
		if parsingErr != nil {
			break
		}
		// Argument name "proposalId" must match the ABI.
		proposalIDBytes, errExtract := m.abiUtils.GetBytes32FromArgs(parsedArgs, "proposalId")
		if errExtract != nil {
			parsingErr = errExtract
			break
		}

		metadataToSet[types.MultiSigProposalIDMetadataKey] = proposalIDBytes
		specificTxType = types.TransactionTypeMultiSigSignRecoveryAddressChange

	default:
		m.logger.Debug("Transaction data method ID does not match any known MultiSig methods for detailed parsing.",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_id_hex", hex.EncodeToString(methodID)))
		return false, nil // Not an error, just not a MultiSig call we're decoding this way.
	}

	if parsingErr != nil {
		m.logger.Warn("Failed to parse MultiSig transaction input or extract arguments for a recognized method.",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_name_attempted", currentMethodNameForParsing), // Will be empty if methodID didn't match any case
			logger.String("method_id_hex", hex.EncodeToString(methodID)),
			logger.Error(parsingErr),
		)
		// Return error because we identified a method but couldn't parse its details.
		return false, fmt.Errorf("failed to parse args for known MultiSig method %s (ID: %s): %w", currentMethodNameForParsing, hex.EncodeToString(methodID), parsingErr)
	}

	// If specificTxType is not set here, it means a case was matched, parsingErr was nil,
	// but logic didn't set the type. This indicates an issue in the switch case itself.
	// Or, no case matched, which is handled by the `default` path returning (false, nil).
	if specificTxType == "" {
		// This condition implies that a methodID matched a case, parsingErr was nil, but specificTxType was not set.
		// This should not happen if all switch cases that don't `break` due to `parsingErr` set `specificTxType`.
		// The `default` case already returned `false, nil`.
		// `parsingErr != nil` case already returned `false, parsingErr`.
		// So if we are here, it's an unexpected state, likely a logic error in one of the switch cases.
		m.logger.Error("Internal logic error: MultiSig method processed but specificTxType not set, despite no parsing error.",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_name_parsed", currentMethodNameForParsing),
			logger.String("method_id_hex", hex.EncodeToString(methodID)))
		return false, fmt.Errorf("internal logic error determining MultiSig tx type for method ID %s", hex.EncodeToString(methodID))
	}

	if tx.Metadata == nil {
		tx.Metadata = make(types.TxMetadata)
	}

	// Also store the determined type in the metadata itself for consistency.
	metadataToSet[types.TransactionTypeMetadaKey] = string(specificTxType)

	err = tx.Metadata.SetAll(metadataToSet)
	if err != nil {
		m.logger.Error("Failed to set MultiSig metadata map after parsing",
			logger.String("tx_hash", tx.Hash),
			logger.String("type", string(specificTxType)),
			logger.Error(err))
		return false, fmt.Errorf("failed to set MultiSig metadata values: %w", err)
	}

	tx.Type = specificTxType

	m.logger.Info("Successfully parsed and populated MultiSig transaction metadata.",
		logger.String("tx_hash", tx.Hash),
		logger.String("parsed_type", string(specificTxType)))
	return true, nil
}

// --- Interface Implementation ---

// DecodeERC20Transfer implements Mapper.
func (m *evmDecoder) DecodeERC20Transfer(ctx context.Context, tx *types.Transaction) (*types.ERC20Transfer, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	// Validate type and metadata existence
	if tx.Type != types.TransactionTypeERC20Transfer {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeERC20Transfer))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to ERC20Transfer")
	}

	// Attempt creation from metadata
	return createERC20TransferFromMetadata(tx)
}

// DecodeTransaction implements Mapper. It attempts to map a generic transaction
// into a more specific type (like ERC20Transfer or a MultiSig operation)
// based on its metadata or by parsing its data if necessary.
func (m *evmDecoder) DecodeTransaction(ctx context.Context, tx *types.Transaction) (types.CoreTransaction, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	// --- Step 1: Check if already specific type ---
	// If the type is already specific (not Unknown or ContractCall), assume it's correctly typed.
	// We could add validation here if needed, but for now, trust the existing type.
	if tx.Type != "" && tx.Type != types.TransactionTypeContractCall {
		m.logger.Debug("Transaction already has a specific type, returning as is", logger.String("tx_hash", tx.Hash), logger.String("type", string(tx.Type)))
		// Directly use the existing specific creation functions which validate metadata
		switch tx.Type {
		case types.TransactionTypeNative:
			return tx, nil
		case types.TransactionTypeDeploy:
			return tx, nil
		case types.TransactionTypeERC20Transfer:
			return m.DecodeERC20Transfer(ctx, tx)
		case types.TransactionTypeMultiSigWithdrawalRequest:
			return m.DecodeMultiSigWithdrawalRequest(ctx, tx)
		case types.TransactionTypeMultiSigSignWithdrawal:
			return m.DecodeMultiSigSignWithdrawal(ctx, tx)
		case types.TransactionTypeMultiSigExecuteWithdrawal:
			return m.DecodeMultiSigExecuteWithdrawal(ctx, tx)
		case types.TransactionTypeMultiSigAddSupportedToken:
			return m.DecodeMultiSigAddSupportedToken(ctx, tx)
		case types.TransactionTypeMultiSigRecoveryRequest:
			return m.DecodeMultiSigRecoveryRequest(ctx, tx)
		case types.TransactionTypeMultiSigCancelRecovery:
			return m.ToMultiSigCancelRecovery(ctx, tx)
		case types.TransactionTypeMultiSigExecuteRecovery:
			return m.ToMultiSigExecuteRecovery(ctx, tx)
		case types.TransactionTypeMultiSigProposeRecoveryAddressChange:
			return m.ToMultiSigProposeRecoveryAddressChange(ctx, tx)
		case types.TransactionTypeMultiSigSignRecoveryAddressChange:
			return m.ToMultiSigSignRecoveryAddressChange(ctx, tx)
		default:
			// This case shouldn't be reached given the initial check, but return original tx if it does
			m.logger.Warn("Transaction has unrecognized specific type",
				logger.String("tx_hash", tx.Hash),
				logger.String("type", string(tx.Type)))
			return tx, nil
		}
	}

	// --- Step 2: Attempt Parsing if Type is Generic (ContractCall or Unknown) ---
	m.logger.Debug("Transaction type is generic, attempting to parse data",
		logger.String("tx_hash", tx.Hash),
		logger.String("type", string(tx.Type)))

	// Ensure basic requirements for parsing contract calls are met
	if tx.BaseTransaction.To == "" || len(tx.Data) < 4 {
		m.logger.Debug("Transaction not eligible for parsing (missing To address or short data)",
			logger.String("tx_hash", tx.Hash))
		return tx, nil // Return original tx as it cannot be parsed
	}

	// Make a copy to attempt parsing without modifying the original until success
	txCopy := tx.Copy()

	// Try parsing as ERC20 Transfer first
	parsedAsERC20, errERC20 := m.parseAndPopulateERC20Metadata(ctx, txCopy)
	if errERC20 != nil {
		// Log critical error during ERC20 parsing (e.g., ABI load failure) but continue to try MultiSig
		m.logger.Error("Error attempting to parse as ERC20 transfer, proceeding to check MultiSig",
			logger.String("tx_hash", txCopy.Hash),
			logger.Error(errERC20),
		)
	} else if parsedAsERC20 {
		m.logger.Debug("Successfully parsed as ERC20Transfer", logger.String("tx_hash", txCopy.Hash))
		// Parsing succeeded, create the typed object from the populated metadata in the copy
		// Use the specific DecodeERC20Transfer which now uses createERC20TransferFromMetadata
		return m.DecodeERC20Transfer(ctx, txCopy)
	}

	// If not parsed as ERC20, try parsing as MultiSig
	parsedAsMultiSig, errMultiSig := m.parseAndPopulateMultiSigMetadata(ctx, txCopy)
	if errMultiSig != nil {
		// Log critical error during MultiSig parsing (e.g., ABI load failure)
		m.logger.Error("Error attempting to parse as MultiSig",
			logger.String("tx_hash", txCopy.Hash),
			logger.Error(errMultiSig),
		)
		// Don't return the error, just return the original transaction
		return tx, nil
	} else if parsedAsMultiSig {
		m.logger.Debug("Successfully parsed as a MultiSig transaction", logger.String("tx_hash", txCopy.Hash), logger.String("new_type", string(txCopy.Type)))
		// Parsing succeeded, create the typed object from the populated metadata in the copy
		// Call the appropriate specific mapper function based on the *updated* type in txCopy
		switch txCopy.Type {
		case types.TransactionTypeMultiSigWithdrawalRequest:
			return m.DecodeMultiSigWithdrawalRequest(ctx, txCopy)
		case types.TransactionTypeMultiSigSignWithdrawal:
			return m.DecodeMultiSigSignWithdrawal(ctx, txCopy)
		case types.TransactionTypeMultiSigExecuteWithdrawal:
			return m.DecodeMultiSigExecuteWithdrawal(ctx, txCopy)
		case types.TransactionTypeMultiSigAddSupportedToken:
			return m.DecodeMultiSigAddSupportedToken(ctx, txCopy)
		case types.TransactionTypeMultiSigRecoveryRequest:
			return m.DecodeMultiSigRecoveryRequest(ctx, txCopy)
		case types.TransactionTypeMultiSigCancelRecovery:
			return m.ToMultiSigCancelRecovery(ctx, txCopy)
		case types.TransactionTypeMultiSigExecuteRecovery:
			return m.ToMultiSigExecuteRecovery(ctx, txCopy)
		case types.TransactionTypeMultiSigProposeRecoveryAddressChange:
			return m.ToMultiSigProposeRecoveryAddressChange(ctx, txCopy)
		case types.TransactionTypeMultiSigSignRecoveryAddressChange:
			return m.ToMultiSigSignRecoveryAddressChange(ctx, txCopy)
		default:
			// This case indicates a logic error in parseAndPopulateMultiSigMetadata if reached
			m.logger.Error("MultiSig parsing succeeded but resulted in an unexpected type",
				logger.String("tx_hash", txCopy.Hash),
				logger.String("type", string(txCopy.Type)),
			)
			return tx, nil // Return original transaction
		}
	}

	// --- Step 3: Return Original if No Specific Type Identified ---
	m.logger.Debug("Transaction data did not match known ERC20 or MultiSig patterns", logger.String("tx_hash", tx.Hash))
	return tx, nil // Return the original transaction if no specific type could be determined
}

// DecodeMultiSigWithdrawalRequest implements Mapper.
func (m *evmDecoder) DecodeMultiSigWithdrawalRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigWithdrawalRequest, error) {
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

// DecodeMultiSigSignWithdrawal implements Mapper.
func (m *evmDecoder) DecodeMultiSigSignWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigSignWithdrawal, error) {
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

// DecodeMultiSigExecuteWithdrawal implements Mapper.
func (m *evmDecoder) DecodeMultiSigExecuteWithdrawal(ctx context.Context, tx *types.Transaction) (*types.MultiSigExecuteWithdrawal, error) {
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

// DecodeMultiSigAddSupportedToken implements Mapper.
func (m *evmDecoder) DecodeMultiSigAddSupportedToken(ctx context.Context, tx *types.Transaction) (*types.MultiSigAddSupportedToken, error) {
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

// DecodeMultiSigRecoveryRequest implements Mapper.
func (m *evmDecoder) DecodeMultiSigRecoveryRequest(ctx context.Context, tx *types.Transaction) (*types.MultiSigRecoveryRequest, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigRecoveryRequest {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigRecoveryRequest))
	}
	// Metadata might be nil for this type, but the creation helper handles nil tx check

	return createMultiSigRecoveryRequestFromMetadata(tx)
}

// ToMultiSigCancelRecovery implements Mapper.
func (m *evmDecoder) ToMultiSigCancelRecovery(ctx context.Context, tx *types.Transaction) (*types.MultiSigCancelRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigCancelRecovery {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigCancelRecovery))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigCancelRecovery")
	}

	return createMultiSigCancelRecoveryFromMetadata(tx)
}

// ToMultiSigExecuteRecovery implements Mapper.
func (m *evmDecoder) ToMultiSigExecuteRecovery(ctx context.Context, tx *types.Transaction) (*types.MultiSigExecuteRecovery, error) {
	if tx == nil {
		return nil, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	if tx.Type != types.TransactionTypeMultiSigExecuteRecovery {
		return nil, errors.NewMappingError(tx.Hash, fmt.Sprintf("invalid transaction type %s, expected %s", tx.Type, types.TransactionTypeMultiSigExecuteRecovery))
	}
	if tx.Metadata == nil {
		return nil, errors.NewMappingError(tx.Hash, "metadata is required to map to MultiSigExecuteRecovery")
	}

	return createMultiSigExecuteRecoveryFromMetadata(tx)
}

// ToMultiSigProposeRecoveryAddressChange implements Mapper.
func (m *evmDecoder) ToMultiSigProposeRecoveryAddressChange(ctx context.Context, tx *types.Transaction) (*types.MultiSigProposeRecoveryAddressChange, error) {
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

// ToMultiSigSignRecoveryAddressChange implements Mapper.
func (m *evmDecoder) ToMultiSigSignRecoveryAddressChange(ctx context.Context, tx *types.Transaction) (*types.MultiSigSignRecoveryAddressChange, error) {
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
