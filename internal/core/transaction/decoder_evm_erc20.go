package transaction

import (
	"bytes"
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"vault0/internal/core/abiutils"   // Ensure abiutils is imported
	"vault0/internal/core/tokenstore" // Ensure tokenstore is imported
	"vault0/internal/errors"
	"vault0/internal/types"
)

// Precompute ERC20 transfer method ID locally for dispatcher logic
var erc20TransferMethodID = crypto.Keccak256([]byte("transfer(address,uint256)"))[:4]

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

// parseAndPopulateERC20Metadata attempts to parse tx data as an ERC20 transfer
// and updates tx.Metadata and tx.Type if successful.
// Returns true if parsing was successful and metadata was populated.
func parseAndPopulateERC20Metadata(ctx context.Context, tx *types.Transaction, abiTools abiutils.ABIUtils, ts tokenstore.TokenStore) (bool, error) {
	if tx == nil {
		return false, errors.NewInvalidParameterError("transaction cannot be nil", "tx")
	}

	// Basic validation for parsing
	if tx.BaseTransaction.To == "" {
		return false, nil
	}
	if len(tx.Data) < 4 {
		return false, nil
	}

	// Extract the method ID using abiutils
	methodID := abiTools.ExtractMethodID(tx.Data)

	// Optimization: Check against precomputed standard ERC20 transfer ID.
	if !bytes.Equal(methodID, erc20TransferMethodID) {
		// Method ID doesn't match, this isn't an ERC20 transfer. Not an error.
		return false, nil
	}

	// Load ABI specifically for the target contract address.
	contractAddrEth := common.HexToAddress(tx.BaseTransaction.To)
	// Convert to types.Address
	contractAddr, err := types.NewAddress(tx.BaseTransaction.ChainType, contractAddrEth.Hex())
	if err != nil {
		return false, fmt.Errorf("failed to convert address %s: %w", contractAddrEth.Hex(), err)
	}

	erc20ABI, err := abiTools.LoadABIByAddress(ctx, *contractAddr)
	if err != nil {
		// Cannot proceed without ABI, but don't return error, just indicate parsing failed.
		return false, fmt.Errorf("failed to load ABI for address %s: %w", contractAddr.String(), err) // Return error to indicate ABI load failure
	}

	// Parse the input data with the "transfer" method name
	parsedArgs, err := abiTools.ParseContractInput(erc20ABI, "transfer", tx.Data)
	if err != nil {
		// Parsing failed, return error.
		return false, fmt.Errorf("failed to parse ERC20 transfer input: %w", err)
	}

	// Extract recipient address ('recipient') and amount ('amount').
	recipientAddr, err := abiTools.GetAddressFromArgs(parsedArgs, types.ERC20RecipientMetadataKey)
	if err != nil {
		return false, fmt.Errorf("failed to get recipient address ('recipient') from parsed args: %w", err)
	}

	amountBigIntParsed, err := abiTools.GetBigIntFromArgs(parsedArgs, types.ERC20AmountMetadataKey)
	if err != nil {
		return false, fmt.Errorf("failed to get transfer amount ('amount') from parsed args: %w", err)
	}

	// Resolve token details (symbol, decimals) using the token store.
	// Use address from tx.To which is the token contract address.
	tokenInfo, err := ts.GetToken(ctx, tx.BaseTransaction.To)
	if err != nil {
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
	})
	if err != nil {
		// This should ideally not happen if types are correct, but handle defensively
		return false, fmt.Errorf("failed to set ERC20 metadata: %w", err)
	}

	// Update transaction type
	tx.Type = types.TransactionTypeERC20Transfer

	return true, nil // Parsing successful
}

// decodeERC20Transfer converts a generic transaction to ERC20Transfer.
func decodeERC20Transfer(tx *types.Transaction) (*types.ERC20Transfer, error) {
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
