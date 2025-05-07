package transaction

import (
	"time"

	"vault0/internal/types"
)

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	ID           string    `json:"id,omitempty"`
	WalletID     string    `json:"wallet_id,omitempty"`
	ChainType    string    `json:"chain_type"`
	Hash         string    `json:"hash"`
	FromAddress  string    `json:"from_address"`
	ToAddress    string    `json:"to_address"`
	Value        string    `json:"value"`
	Data         string    `json:"data,omitempty"`
	Nonce        uint64    `json:"nonce"`
	GasPrice     string    `json:"gas_price,omitempty"`
	GasLimit     uint64    `json:"gas_limit,omitempty"`
	Type         string    `json:"type"`
	TokenAddress string    `json:"token_address,omitempty"`
	TokenSymbol  string    `json:"token_symbol,omitempty"`
	Status       string    `json:"status"`
	Timestamp    int64     `json:"timestamp"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
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
func ToResponse(tx types.CoreTransaction, token *types.Token) TransactionResponse {
	if token == nil {
		// Fallback to default token with 18 decimals if none provided
		token = &types.Token{Decimals: 18}
	}

	// Get the native token for the chain to format gas price
	nativeToken, err := types.NewNativeToken(tx.GetChainType())
	if err != nil {
		// Fallback to 18 decimals if native token lookup fails (should not happen for valid chains)
		nativeToken = &types.Token{Decimals: 18}
	}

	// Format valueStr using the specific token's decimals
	valueStr := "0"
	if tx.GetValue() != nil && tx.GetValue().Sign() > 0 {
		valueStr = token.ToBigFloat(tx.GetValue()).Text('f', int(token.Decimals))
	}

	// Format gasPriceStr using the NATIVE token's decimals
	gasPriceStr := "0"
	if tx.GetGasPrice() != nil && tx.GetGasPrice().Sign() > 0 {
		gasPriceStr = nativeToken.ToBigFloat(tx.GetGasPrice()).Text('f', int(nativeToken.Decimals))
	}

	dataStr := ""
	if len(tx.GetData()) > 0 {
		dataStr = "0x" + string(tx.GetData())
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
	}

	// Handle specific transaction types
	switch typedTx := tx.(type) {
	case *types.ERC20Transfer:
		response.TokenAddress = typedTx.TokenAddress
		response.TokenSymbol = token.Symbol
	case *types.MultiSigWithdrawalRequest:
		response.TokenAddress = typedTx.Token
		response.TokenSymbol = token.Symbol
	case *types.MultiSigExecuteWithdrawal:
		// Additional fields could be added for this type
	}

	return response
}
