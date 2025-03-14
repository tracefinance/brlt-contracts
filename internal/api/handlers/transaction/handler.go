package transaction

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

// Handler handles transaction-related HTTP requests
type Handler struct {
	transactionService transaction.Service
}

// NewHandler creates a new transaction handler
func NewHandler(transactionService transaction.Service) *Handler {
	return &Handler{
		transactionService: transactionService,
	}
}

// SetupRoutes sets up the transaction routes
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Wallet-scoped transaction routes
	walletRoutes := router.Group("/wallets/:chain_type/:address/transactions")
	walletRoutes.GET("", h.GetTransactionsByAddress)
	walletRoutes.GET("/:hash", h.GetTransaction)
	walletRoutes.POST("/sync", h.SyncTransactions)

	// Direct transaction routes
	transactionRoutes := router.Group("/transactions")
	transactionRoutes.GET("/:hash", h.GetTransaction)
}

// GetTransaction handles GET /wallets/:chain_type/:address/transactions/:hash
// or GET /transactions/:hash
func (h *Handler) GetTransaction(c *gin.Context) {
	// Get chain type either from URL or try to infer from hash
	var chainType types.ChainType
	if c.Param("chain_type") != "" {
		chainType = types.ChainType(c.Param("chain_type"))
	}

	hash := c.Param("hash")

	tx, err := h.transactionService.GetTransaction(c.Request.Context(), chainType, hash)
	if err != nil {
		if errors.Is(err, transaction.ErrTransactionNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transaction"})
		return
	}

	c.JSON(http.StatusOK, FromServiceTransaction(tx))
}

// GetTransactionsByAddress handles GET /wallets/:chain_type/:address/transactions
// or GET /transactions/:chain_type/:address
func (h *Handler) GetTransactionsByAddress(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Get transactions
	txs, err := h.transactionService.GetTransactionsByAddress(c.Request.Context(), chainType, address, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get transactions"})
		return
	}

	// Count total transactions
	total, err := h.transactionService.CountTransactions(c.Request.Context(), "")
	if err != nil {
		total = len(txs)
	}

	// Convert to response format
	var transactions []TransactionResponse
	for _, tx := range txs {
		transactions = append(transactions, FromServiceTransaction(tx))
	}

	c.JSON(http.StatusOK, TransactionListResponse{
		Transactions: transactions,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	})
}

// SyncTransactions handles POST /wallets/:chain_type/:address/transactions/sync
// or POST /transactions/:chain_type/:address/sync
func (h *Handler) SyncTransactions(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Sync transactions
	count, err := h.transactionService.SyncTransactionsByAddress(c.Request.Context(), chainType, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to sync transactions"})
		return
	}

	c.JSON(http.StatusOK, SyncTransactionsResponse{
		Count: count,
	})
}
