package transaction

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
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
	// Create error handler middleware
	errorHandler := middleares.NewErrorHandler(nil)

	// Wallet-scoped transaction routes
	walletRoutes := router.Group("/wallets/:chain_type/:address/transactions")
	walletRoutes.Use(errorHandler.Middleware())
	walletRoutes.GET("", h.GetTransactionsByAddress)
	walletRoutes.GET("/:hash", h.GetTransaction)
	walletRoutes.POST("/sync", h.SyncTransactions)

	// Direct transaction routes
	transactionRoutes := router.Group("/transactions")
	transactionRoutes.Use(errorHandler.Middleware())
	transactionRoutes.GET("/:hash", h.GetTransaction)
}

// GetTransaction handles GET /wallets/:chain_type/:address/transactions/:hash
// or GET /transactions/:hash
// @Summary Get a transaction
// @Description Get transaction details by hash
// @Tags transactions
// @Produce json
// @Param chain_type path string true "Chain type (required only for wallet-scoped route)"
// @Param address path string true "Wallet address (required only for wallet-scoped route)"
// @Param hash path string true "Transaction hash"
// @Success 200 {object} TransactionResponse
// @Failure 404 {object} errors.Vault0Error "Transaction not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /transactions/{hash} [get]
// @Router /wallets/{chain_type}/{address}/transactions/{hash} [get]
func (h *Handler) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")

	tx, err := h.transactionService.GetTransaction(c.Request.Context(), hash)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, FromServiceTransaction(tx))
}

// GetTransactionsByAddress handles GET /wallets/:chain_type/:address/transactions
// @Summary List transactions for an address
// @Description Get a paginated list of transactions for a specific wallet address
// @Tags transactions
// @Produce json
// @Param chain_type path string true "Chain type"
// @Param address path string true "Wallet address"
// @Param limit query int false "Number of items to return (default: 10)" default(10)
// @Param offset query int false "Number of items to skip (default: 0)" default(0)
// @Success 200 {object} PagedTransactionsResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address}/transactions [get]
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

	// Get transactions with pagination
	page, err := h.transactionService.GetTransactionsByAddress(c.Request.Context(), chainType, address, limit, offset)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToPagedResponse(page))
}

// SyncTransactions handles POST /wallets/:chain_type/:address/transactions/sync
// @Summary Sync transactions for an address
// @Description Sync blockchain transactions for a specific wallet address
// @Tags transactions
// @Produce json
// @Param chain_type path string true "Chain type"
// @Param address path string true "Wallet address"
// @Success 200 {object} SyncTransactionsResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address}/transactions/sync [post]
func (h *Handler) SyncTransactions(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Sync transactions
	count, err := h.transactionService.SyncTransactionsByAddress(c.Request.Context(), chainType, address)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, SyncTransactionsResponse{
		Count: count,
	})
}
