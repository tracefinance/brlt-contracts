package transaction

import (
	"strconv"
	"time"

	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	ID           string    `json:"id"`
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
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
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

// ToResponse converts a service transaction to a response transaction
func ToResponse(tx *transaction.Transaction, token *types.Token) TransactionResponse {
	// Get the native token for the chain to format gas price
	nativeToken, err := types.NewNativeToken(tx.ChainType)
	if err != nil {
		// Fallback to 18 decimals if native token lookup fails (should not happen for valid chains)
		nativeToken = &types.Token{Decimals: 18}
	}

	// Format valueStr using the specific token's decimals
	valueStr := "0"
	if !tx.Value.IsZero() {
		valueStr = token.ToBigFloat(tx.Value.ToBigInt()).Text('f', int(token.Decimals))
	}

	// Format gasPriceStr using the NATIVE token's decimals
	gasPriceStr := "0"
	if !tx.GasPrice.IsZero() {
		gasPriceStr = nativeToken.ToBigFloat(tx.GasPrice.ToBigInt()).Text('f', int(nativeToken.Decimals))
	}

	dataStr := ""
	if len(tx.Data) > 0 {
		dataStr = "0x" + string(tx.Data)
	}

	return TransactionResponse{
		ID:           strconv.FormatInt(tx.ID, 10),
		WalletID:     strconv.FormatInt(tx.WalletID, 10),
		ChainType:    string(tx.ChainType),
		Hash:         tx.Hash,
		FromAddress:  tx.FromAddress,
		ToAddress:    tx.ToAddress,
		Value:        valueStr,
		Data:         dataStr,
		Nonce:        tx.Nonce,
		GasPrice:     gasPriceStr,
		GasLimit:     tx.GasLimit,
		Type:         tx.Type,
		TokenAddress: tx.TokenAddress,
		TokenSymbol:  tx.TokenSymbol,
		Status:       tx.Status,
		Timestamp:    tx.Timestamp,
		CreatedAt:    tx.CreatedAt,
		UpdatedAt:    tx.UpdatedAt,
	}
}
