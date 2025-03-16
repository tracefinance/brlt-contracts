package transaction

import (
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
func FromServiceTransaction(tx *transaction.Transaction) TransactionResponse {
	valueStr := ""
	if tx.Value != nil {
		valueStr = tx.Value.String()
	}

	gasPriceStr := ""
	if tx.GasPrice != nil {
		gasPriceStr = tx.GasPrice.String()
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
		Status:       tx.Status,
		Timestamp:    tx.Timestamp,
		CreatedAt:    tx.CreatedAt,
		UpdatedAt:    tx.UpdatedAt,
	}
}

// ToResponseList converts a slice of service transactions to a slice of response transactions
func ToResponseList(txs []*transaction.Transaction) []TransactionResponse {
	responses := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = FromServiceTransaction(tx)
	}
	return responses
}

// ToPagedResponse converts a Page of service transactions to a TransactionListResponse
func ToPagedResponse(page *types.Page[*transaction.Transaction]) *PagedTransactionsResponse {
	return &PagedTransactionsResponse{
		Items:   ToResponseList(page.Items),
		Limit:   page.Limit,
		Offset:  page.Offset,
		HasMore: page.HasMore,
	}
}
