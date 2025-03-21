package transaction

import (
	"time"

	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

// TransactionResponse represents a transaction in API responses
type TransactionResponse struct {
	ID           int64     `json:"id"`
	WalletID     int64     `json:"wallet_id,omitempty"`
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

// PagedTransactionsResponse represents a paginated list of transactions
type PagedTransactionsResponse struct {
	Items   []TransactionResponse `json:"items"`
	Limit   int                   `json:"limit"`
	Offset  int                   `json:"offset"`
	HasMore bool                  `json:"has_more"`
}

// SyncTransactionsResponse represents the response for a transaction sync operation
type SyncTransactionsResponse struct {
	Count int `json:"count"`
}

// FromServiceTransaction converts a service transaction to a response transaction
func FromServiceTransaction(tx *transaction.Transaction, token *types.Token) TransactionResponse {
	valueStr := ""
	if tx.Value != nil {
		valueStr = token.ToBigFloat(tx.Value).Text('f', int(token.Decimals))
	}

	gasPriceStr := ""
	if tx.GasPrice != nil {
		gasPriceStr = token.ToBigFloat(tx.GasPrice).Text('f', int(token.Decimals))
	}

	dataStr := ""
	if len(tx.Data) > 0 {
		dataStr = "0x" + string(tx.Data)
	}

	return TransactionResponse{
		ID:           tx.ID,
		WalletID:     tx.WalletID,
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

// ToResponseList converts a slice of service transactions to a slice of response transactions
func ToResponseList(txs []*transaction.Transaction, tokensMap map[string]*types.Token) []TransactionResponse {
	responses := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		// Get native token for this chain
		nativeToken, ok := tokensMap[tx.TokenAddress]
		if !ok {
			// Fallback to direct conversion if token not found
			nativeToken = &types.Token{Decimals: 18} // Default to 18 decimals
		}
		responses[i] = FromServiceTransaction(tx, nativeToken)
	}
	return responses
}

// ToPagedResponse converts a Page of service transactions to a TransactionListResponse
func ToPagedResponse(page *types.Page[*transaction.Transaction], tokensMap map[string]*types.Token) *PagedTransactionsResponse {
	return &PagedTransactionsResponse{
		Items:   ToResponseList(page.Items, tokensMap),
		Limit:   page.Limit,
		Offset:  page.Offset,
		HasMore: page.HasMore,
	}
}
