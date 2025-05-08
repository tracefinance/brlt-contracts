package transaction

import (
	"encoding/hex"

	"vault0/internal/types"
)

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	ChainType     string `json:"chain_type"`
	Hash          string `json:"hash"`
	FromAddress   string `json:"from_address"`
	ToAddress     string `json:"to_address"`
	Value         string `json:"value"`
	Data          string `json:"data,omitempty"`
	Nonce         uint64 `json:"nonce"`
	GasPrice      string `json:"gas_price,omitempty"`
	GasLimit      uint64 `json:"gas_limit,omitempty"`
	Type          string `json:"type"`
	TokenAddress  string `json:"token_address,omitempty"`
	TokenSymbol   string `json:"token_symbol,omitempty"`
	TokenDecimals *uint8 `json:"token_decimals,omitempty"`
	Status        string `json:"status"`
	Timestamp     int64  `json:"timestamp"`

	// MultiSig specific fields
	WithdrawalNonce    *uint64 `json:"withdrawal_nonce,omitempty"`
	RequestID          string  `json:"request_id,omitempty"`
	ProposalID         string  `json:"proposal_id,omitempty"`
	TargetTokenAddress string  `json:"target_token_address,omitempty"`
	NewRecoveryAddress string  `json:"new_recovery_address,omitempty"`
}

// ListTransactionsRequest defines the query parameters for listing transactions
type ListTransactionsRequest struct {
	NextToken    string `form:"next_token"`
	Limit        *int   `form:"limit" binding:"omitempty,min=1"`
	ChainType    string `form:"chain_type"`
	Address      string `form:"address"`
	TokenAddress string `form:"token_address"`
	Status       string `form:"status"`
}

// ListTransactionsByAddressRequest defines query parameters for listing transactions by address
type ListTransactionsByAddressRequest struct {
	NextToken    string `form:"next_token"`
	Limit        *int   `form:"limit" binding:"omitempty,min=1"`
	TokenAddress string `form:"token_address"`
}

// SyncTransactionsResponse represents the response for a transaction sync operation
type SyncTransactionsResponse struct {
	Count int `json:"count"`
}

// ToResponse converts a CoreTransaction to a response transaction
func ToResponse(tx types.CoreTransaction) TransactionResponse {
	// Get the native token for the chain to format gas price and default value
	nativeToken, err := types.NewNativeToken(tx.GetChainType())
	if err != nil {
		// Fallback to 18 decimals if native token lookup fails (should not happen for valid chains)
		nativeToken = &types.Token{Decimals: 18}
	}

	// Format valueStr: Default to native token decimals. Will be overridden for specific token types.
	valueStr := "0"
	if tx.GetValue() != nil && tx.GetValue().Sign() > 0 {
		valueStr = nativeToken.ToBigFloat(tx.GetValue()).Text('f', int(nativeToken.Decimals))
	}

	// Format gasPriceStr using the NATIVE token's decimals
	gasPriceStr := "0"
	if tx.GetGasPrice() != nil && tx.GetGasPrice().Sign() > 0 {
		gasPriceStr = nativeToken.ToBigFloat(tx.GetGasPrice()).Text('f', int(nativeToken.Decimals))
	}

	dataStr := ""
	if len(tx.GetData()) > 0 {
		// Ensure data is hex prefixed if it's not empty.
		// Some blockchain explorers/systems expect hex data for transactions.
		dataStr = "0x" + hex.EncodeToString(tx.GetData())
	}

	// Create base response
	response := TransactionResponse{
		ChainType:   string(tx.GetChainType()),
		Hash:        tx.GetHash(),
		FromAddress: tx.GetFrom(),
		ToAddress:   tx.GetTo(),
		Value:       valueStr,
		Data:        dataStr,
		Nonce:       tx.GetNonce(),
		GasPrice:    gasPriceStr,
		GasLimit:    tx.GetGasLimit(),
		Type:        string(tx.GetType()),
	}

	// Check if transaction is a concrete Transaction with more fields
	if concreteTx, ok := tx.(*types.Transaction); ok {
		response.Status = string(concreteTx.Status)
		response.Timestamp = concreteTx.Timestamp
		// Note: ID, WalletID, CreatedAt, UpdatedAt are part of TransactionResponse struct,
		// but are not populated here from types.Transaction as it may not have them.
		// They would typically be set by the calling service if available from a data store.
	}

	// Handle specific transaction types
	switch typedTx := tx.(type) {
	case *types.ERC20Transfer:
		amount := typedTx.Amount
		response.TokenAddress = typedTx.TokenAddress
		response.TokenSymbol = typedTx.TokenSymbol
		decimals := typedTx.TokenDecimals
		response.TokenDecimals = &decimals
		// Override value formatting using ERC20's own decimals
		if amount != nil && amount.Sign() > 0 {
			// Create a temporary token struct with the correct decimals for formatting
			tempTokenForFormatting := &types.Token{Decimals: typedTx.TokenDecimals}
			response.Value = tempTokenForFormatting.ToBigFloat(amount).Text('f', int(typedTx.TokenDecimals))
		}
	case *types.MultiSigWithdrawalRequest:
		amount := typedTx.Amount
		response.TokenAddress = typedTx.TokenAddress
		response.TokenSymbol = typedTx.TokenSymbol // Use symbol from MultiSigWithdrawalRequest
		decimals := typedTx.TokenDecimals          // Use decimals from MultiSigWithdrawalRequest
		response.TokenDecimals = &decimals
		response.WithdrawalNonce = &typedTx.WithdrawalNonce
		// Override value formatting using the token's own decimals for this multisig op
		// tx.GetValue() should correspond to typedTx.Amount for this transaction type
		if amount != nil && amount.Sign() > 0 {
			// Create a temporary token struct with the correct decimals for formatting
			tempTokenForFormatting := &types.Token{Decimals: typedTx.TokenDecimals}
			response.Value = tempTokenForFormatting.ToBigFloat(amount).Text('f', int(typedTx.TokenDecimals))
		}
	case *types.MultiSigSignWithdrawal:
		response.RequestID = hex.EncodeToString(typedTx.RequestID[:])
	case *types.MultiSigExecuteWithdrawal:
		response.RequestID = hex.EncodeToString(typedTx.RequestID[:])
	case *types.MultiSigAddSupportedToken:
		response.TargetTokenAddress = typedTx.Token
	case *types.MultiSigProposeRecoveryAddressChange:
		response.NewRecoveryAddress = typedTx.NewRecoveryAddress
	case *types.MultiSigSignRecoveryAddressChange:
		response.ProposalID = hex.EncodeToString(typedTx.ProposalID[:])
		// Other MultiSig types (RecoveryRequest, CancelRecovery, ExecuteRecovery)
		// currently do not have additional specific fields for the DTO beyond base Transaction.
	}

	return response
}
