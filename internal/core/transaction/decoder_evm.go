package transaction

import (
	"bytes"
	"context"
	"fmt"

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
		return false, nil // Not an error, just not parsable as this type
	}
	if len(tx.Data) < 4 {
		m.logger.Debug("Skipping ERC20 parse: tx.Data too short", logger.String("tx_hash", tx.Hash), logger.Int("data_len", len(tx.Data)))
		return false, nil // Not an error, just not parsable
	}

	methodID := m.abiUtils.ExtractMethodID(tx.Data)

	// Optimization: Check against precomputed standard ERC20 transfer ID.
	if !bytes.Equal(methodID, erc20TransferMethodID) {
		// Method ID doesn't match, this isn't an ERC20 transfer. Not an error.
		return false, nil
	}

	m.logger.Debug("Method ID matches ERC20 transfer, attempting parse", logger.String("tx_hash", tx.Hash))

	// Load ABI specifically for the target contract address.
	contractAddr := common.HexToAddress(tx.BaseTransaction.To)
	erc20ABI, err := m.abiUtils.LoadABIByAddress(ctx, contractAddr)
	if err != nil {
		m.logger.Warn("Failed to load ABI by address for potential ERC20 transfer parse",
			logger.String("address", contractAddr.Hex()),
			logger.String("tx_hash", tx.Hash),
			logger.Error(err),
		)
		// Cannot proceed without ABI, but don't return error, just indicate parsing failed.
		// The caller (ToTypedTransaction) might handle this gracefully.
		return false, fmt.Errorf("failed to load ABI for address %s: %w", contractAddr.Hex(), err) // Return error to indicate ABI load failure
	}

	// Verify the method signature in the loaded ABI matches "transfer".
	transferMethod, err := m.abiUtils.GetMethodFromABI(erc20ABI, methodID)
	if err != nil || transferMethod.Name != "transfer" {
		m.logger.Warn("Method ID matched standard, but not 'transfer' in loaded ABI",
			logger.String("tx_hash", tx.Hash),
			logger.String("address", contractAddr.Hex()),
			logger.String("method_id", common.Bytes2Hex(methodID)),
			logger.Error(err), // Log potential GetMethodFromABI error
		)
		// ABI structure mismatch, not a standard transfer. Indicate parsing failed.
		return false, nil
	}

	// Parse the input data (excluding the 4-byte method ID).
	parsedArgs, err := m.abiUtils.ParseContractInput(transferMethod, tx.Data[4:])
	if err != nil {
		m.logger.Warn("Failed to parse ERC20 transfer input data",
			logger.String("tx_hash", tx.Hash),
			logger.String("address", contractAddr.Hex()),
			logger.Error(err),
		)
		// Parsing failed, return error.
		return false, fmt.Errorf("failed to parse ERC20 transfer input: %w", err)
	}

	// Extract recipient address ('to') and amount ('value').
	recipientAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "to")
	if err != nil {
		return false, fmt.Errorf("failed to get recipient address ('to') from parsed args: %w", err)
	}

	amountBigIntParsed, err := m.abiUtils.GetBigIntFromArgs(parsedArgs, "value")
	if err != nil {
		return false, fmt.Errorf("failed to get transfer amount ('value') from parsed args: %w", err)
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

	// Basic validation, but allow short data for specific recovery methods
	if len(tx.Data) < 4 {
		// Check later if it's a known short-data method like requestRecovery
		m.logger.Debug("MultiSig parse: tx.Data potentially too short, will check method", logger.String("tx_hash", tx.Hash), logger.Int("data_len", len(tx.Data)))
		// Don't return yet
	}

	methodID := m.abiUtils.ExtractMethodID(tx.Data) // Works even if len(tx.Data) < 4, returns empty slice

	// Load the MultiSig ABI
	multiSigABI, err := m.abiUtils.LoadABIByName(ctx, abiutils.ABITypeMultiSig)
	if err != nil {
		m.logger.Error("Failed to load MultiSig ABI for parsing", logger.String("tx_hash", tx.Hash), logger.Error(err))
		// Critical error, cannot proceed with MultiSig parsing
		return false, fmt.Errorf("failed to load MultiSig ABI: %w", err)
	}

	// Get the method from the ABI
	method, err := m.abiUtils.GetMethodFromABI(multiSigABI, methodID)
	if err != nil {
		// Method ID not found in the MultiSig ABI, this isn't a known MultiSig call. Not an error.
		m.logger.Debug("Method ID not found in MultiSig ABI", logger.String("tx_hash", tx.Hash), logger.String("method_id", common.Bytes2Hex(methodID)))
		return false, nil
	}

	m.logger.Debug("Method ID matches MultiSig method, attempting parse",
		logger.String("tx_hash", tx.Hash),
		logger.String("method_name", method.Name),
		logger.String("method_id", common.Bytes2Hex(methodID)),
	)

	// Handle methods potentially having short data
	isShortDataAllowed := method.Name == "requestRecovery" || method.Name == "cancelRecovery" || method.Name == "executeRecovery"
	if len(tx.Data) < 4 && !isShortDataAllowed {
		m.logger.Warn("MultiSig parse: tx.Data too short for non-recovery method",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_name", method.Name),
		)
		return false, fmt.Errorf("transaction data too short for MultiSig method %s", method.Name)
	}

	var parsedArgs map[string]any
	// Only parse if data is long enough (or method allows short data and has no inputs expected)
	if len(tx.Data) >= 4 {
		parsedArgs, err = m.abiUtils.ParseContractInput(method, tx.Data[4:])
		if err != nil {
			m.logger.Warn("Failed to parse MultiSig input data",
				logger.String("tx_hash", tx.Hash),
				logger.String("method_name", method.Name),
				logger.Error(err),
			)
			return false, fmt.Errorf("failed to parse MultiSig %s input: %w", method.Name, err)
		}
	} else {
		// Handle case for short-data methods (e.g. recovery methods) that have no inputs
		if len(method.Inputs) > 0 {
			// This case should ideally be caught by the len(tx.Data) < 4 check above, but double-check
			return false, fmt.Errorf("transaction data too short for MultiSig method %s which expects inputs", method.Name)
		}
		// No inputs expected, initialize parsedArgs to avoid nil map issues later
		parsedArgs = make(map[string]any)
	}

	// Prepare metadata map
	if tx.Metadata == nil {
		tx.Metadata = make(types.TxMetadata)
	}

	// Switch on the method name to extract specific args, populate metadata, and set type
	var specificTxType types.TransactionType
	metadataToSet := make(map[string]any)

	switch method.Name {
	case "requestWithdrawal":
		specificTxType = types.TransactionTypeMultiSigWithdrawalRequest
		tokenAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "token")
		if err != nil {
			return false, fmt.Errorf("requestWithdrawal: %w", err)
		}
		recipientAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "recipient")
		if err != nil {
			return false, fmt.Errorf("requestWithdrawal: %w", err)
		}
		amountBigInt, err := m.abiUtils.GetBigIntFromArgs(parsedArgs, "amount")
		if err != nil {
			return false, fmt.Errorf("requestWithdrawal: %w", err)
		}
		withdrawalNonce, err := m.abiUtils.GetUint64FromArgs(parsedArgs, "withdrawalNonce")
		if err != nil {
			return false, fmt.Errorf("requestWithdrawal: %w", err)
		}

		metadataToSet[types.MultiSigTokenMetadataKey] = tokenAddr
		metadataToSet[types.MultiSigRecipientMetadataKey] = recipientAddr
		metadataToSet[types.MultiSigAmountMetadataKey] = amountBigInt.ToBigInt()
		metadataToSet[types.MultiSigWithdrawalNonceMetadataKey] = withdrawalNonce

	case "signWithdrawal", "executeWithdrawal":
		if method.Name == "signWithdrawal" {
			specificTxType = types.TransactionTypeMultiSigSignWithdrawal
		} else {
			specificTxType = types.TransactionTypeMultiSigExecuteWithdrawal
		}
		requestIDBytes32, err := m.abiUtils.GetBytes32FromArgs(parsedArgs, "requestID")
		if err != nil {
			return false, fmt.Errorf("%s: %w", method.Name, err)
		}
		metadataToSet[types.MultiSigRequestIDMetadataKey] = requestIDBytes32

	case "addSupportedToken":
		specificTxType = types.TransactionTypeMultiSigAddSupportedToken
		tokenAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "token")
		if err != nil {
			return false, fmt.Errorf("addSupportedToken: %w", err)
		}
		metadataToSet[types.MultiSigTokenMetadataKey] = tokenAddr

	case "requestRecovery":
		specificTxType = types.TransactionTypeMultiSigRecoveryRequest
		// No arguments

	case "cancelRecovery":
		specificTxType = types.TransactionTypeMultiSigCancelRecovery
		// No arguments

	case "executeRecovery":
		specificTxType = types.TransactionTypeMultiSigExecuteRecovery
		// No arguments

	case "proposeRecoveryAddressChange":
		specificTxType = types.TransactionTypeMultiSigProposeRecoveryAddressChange
		newRecoveryAddr, err := m.abiUtils.GetAddressFromArgs(parsedArgs, "newRecoveryAddress")
		if err != nil {
			return false, fmt.Errorf("proposeRecoveryAddressChange: %w", err)
		}
		metadataToSet[types.MultiSigNewRecoveryAddressMetadataKey] = newRecoveryAddr

	case "signRecoveryAddressChange":
		specificTxType = types.TransactionTypeMultiSigSignRecoveryAddressChange
		proposalIDBytes32, err := m.abiUtils.GetBytes32FromArgs(parsedArgs, "proposalId")
		if err != nil {
			return false, fmt.Errorf("signRecoveryAddressChange: %w", err)
		}
		metadataToSet[types.MultiSigProposalIDMetadataKey] = proposalIDBytes32

	default:
		// Should not happen if GetMethodFromABI succeeded, but handle defensively
		m.logger.Warn("Parsed MultiSig method name not handled in switch",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_name", method.Name),
		)
		return false, nil // Not a supported MultiSig type for mapping
	}

	// Set the common type metadata key
	metadataToSet[types.TransactionTypeMetadaKey] = string(specificTxType)

	// Set all collected metadata
	err = tx.Metadata.SetAll(metadataToSet)
	if err != nil {
		m.logger.Error("Failed to set MultiSig metadata after parsing",
			logger.String("tx_hash", tx.Hash),
			logger.String("method_name", method.Name),
			logger.Error(err),
		)
		return false, fmt.Errorf("failed to set MultiSig metadata for %s: %w", method.Name, err)
	}

	// Update transaction type
	tx.Type = specificTxType

	m.logger.Info("Successfully parsed and populated MultiSig transaction metadata",
		logger.String("tx_hash", tx.Hash),
		logger.String("method_name", method.Name),
		logger.String("new_type", string(tx.Type)),
	)
	return true, nil // Parsing successful
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
			m.logger.Warn("Transaction has unrecognized specific type", logger.String("tx_hash", tx.Hash), logger.String("type", string(tx.Type)))
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
		// Use the specific ToERC20Transfer which now uses createERC20TransferFromMetadata
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
