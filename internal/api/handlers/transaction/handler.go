package transaction

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
	"vault0/internal/services/token"
	"vault0/internal/services/transaction"
	"vault0/internal/types"
)

// Handler handles transaction-related HTTP requests
type Handler struct {
	transactionService transaction.Service
	tokenService       token.Service
}

// NewHandler creates a new transaction handler
func NewHandler(transactionService transaction.Service, tokenService token.Service) *Handler {
	return &Handler{
		transactionService: transactionService,
		tokenService:       tokenService,
	}
}

// SetupRoutes sets up the transaction routes
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Create error handler middleware
	errorHandler := middleares.NewErrorHandler(nil)

	// Wallet-scoped transaction routes
	walletRoutes := router.Group("/wallets/:address/:chain_type/transactions")
	walletRoutes.Use(errorHandler.Middleware())
	walletRoutes.GET("", h.GetTransactionsByAddress)
	walletRoutes.GET("/:hash", h.GetTransaction)
	walletRoutes.POST("/sync", h.SyncTransactions)

	// Direct transaction routes
	transactionRoutes := router.Group("/transactions")
	transactionRoutes.Use(errorHandler.Middleware())
	transactionRoutes.GET("/:hash", h.GetTransaction)
	transactionRoutes.GET("", h.FilterTransactions)
}

// GetTransaction handles GET /wallets/:address/:chain_type/transactions/:hash
// or GET /transactions/:hash
// @Summary Get a transaction
// @Description Get transaction details by hash
// @Tags transactions
// @Produce json
// @Param address path string true "Wallet address (required only for wallet-scoped route)"
// @Param chain_type path string true "Chain type (required only for wallet-scoped route)"
// @Param hash path string true "Transaction hash"
// @Success 200 {object} TransactionResponse
// @Failure 404 {object} errors.Vault0Error "Transaction not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /transactions/{hash} [get]
// @Router /wallets/{address}/{chain_type}/transactions/{hash} [get]
func (h *Handler) GetTransaction(c *gin.Context) {
	hash := c.Param("hash")

	tx, err := h.transactionService.GetTransaction(c.Request.Context(), hash)
	if err != nil {
		c.Error(err)
		return
	}

	// Get token from tx.TokenAddress
	token, err := h.tokenService.GetToken(c.Request.Context(), tx.TokenAddress)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, FromServiceTransaction(tx, token))
}

// GetTransactionsByAddress handles GET /wallets/:address/:chain_type/transactions
// @Summary List transactions for an address
// @Description Get a paginated list of transactions for a specific wallet address
// @Tags transactions
// @Produce json
// @Param address path string true "Wallet address"
// @Param chain_type path string true "Chain type"
// @Param limit query int false "Number of items to return (default: 10)" default(10)
// @Param offset query int false "Number of items to skip (default: 0)" default(0)
// @Param token_address query string false "Filter transactions by token address (use 'native' for native transactions)"
// @Success 200 {object} PagedTransactionsResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{address}/{chain_type}/transactions [get]
func (h *Handler) GetTransactionsByAddress(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")
	tokenAddress := c.Query("token_address")

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

	var page *types.Page[*transaction.Transaction]

	if tokenAddress != "" {
		// Create a filter for transactions with the specified token address
		filter := transaction.NewFilter().
			WithChainType(chainType).
			WithAddress(address).
			WithTokenAddress(tokenAddress).
			WithPagination(limit, offset)

		// Use filter-based transaction retrieval
		page, err = h.transactionService.FilterTransactions(c.Request.Context(), filter)
	} else {
		// Use the standard address-based retrieval
		page, err = h.transactionService.GetTransactionsByAddress(c.Request.Context(), chainType, address, limit, offset)
	}

	if err != nil {
		c.Error(err)
		return
	}

	// Get all tokenAddresses from transactions
	tokenAddresses := make([]string, 0, len(page.Items))
	for _, tx := range page.Items {
		tokenAddresses = append(tokenAddresses, tx.TokenAddress)
	}

	// Get all tokens for the addresses
	tokens, err := h.tokenService.ListTokensByAddresses(c.Request.Context(), chainType, tokenAddresses)
	if err != nil {
		c.Error(err)
		return
	}

	// Create a map of tokens by address for efficient lookup
	tokensMap := make(map[string]*types.Token)
	for i := range tokens {
		tokensMap[tokens[i].Address] = &tokens[i]
	}

	c.JSON(http.StatusOK, ToPagedResponse(page, tokensMap))
}

// SyncTransactions handles POST /wallets/:address/:chain_type/transactions/sync
// @Summary Sync transactions for an address
// @Description Sync blockchain transactions for a specific wallet address
// @Tags transactions
// @Produce json
// @Param address path string true "Wallet address"
// @Param chain_type path string true "Chain type"
// @Success 200 {object} SyncTransactionsResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{address}/{chain_type}/transactions/sync [post]
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

// FilterTransactions handles GET /transactions
// @Summary Filter transactions
// @Description Get a paginated list of transactions based on filter criteria
// @Tags transactions
// @Produce json
// @Param chain_type query string false "Filter by chain type"
// @Param address query string false "Filter by wallet address (from or to)"
// @Param token_address query string false "Filter by token address (use 'native' for native transactions)"
// @Param status query string false "Filter by transaction status"
// @Param limit query int false "Number of items to return (default: 10)" default(10)
// @Param offset query int false "Number of items to skip (default: 0)" default(0)
// @Success 200 {object} PagedTransactionsResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /transactions [get]
func (h *Handler) FilterTransactions(c *gin.Context) {
	// Create a new filter
	filter := transaction.NewFilter()

	// Apply chain type filter if provided
	chainTypeStr := c.Query("chain_type")
	if chainTypeStr != "" {
		chainType := types.ChainType(chainTypeStr)
		filter.WithChainType(chainType)
	}

	// Apply address filter if provided
	address := c.Query("address")
	if address != "" {
		filter.WithAddress(address)
	}

	// Apply token address filter if provided
	tokenAddress := c.Query("token_address")
	if tokenAddress != "" {
		filter.WithTokenAddress(tokenAddress)
	}

	// Apply status filter if provided
	status := c.Query("status")
	if status != "" {
		filter.WithStatus(status)
	}

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

	filter.WithPagination(limit, offset)

	// Get transactions with the applied filters
	page, err := h.transactionService.FilterTransactions(c.Request.Context(), filter)
	if err != nil {
		c.Error(err)
		return
	}

	// Get all chain types to prepare for token lookup
	chainTypes := make(map[types.ChainType]bool)
	for _, tx := range page.Items {
		chainTypes[types.ChainType(tx.ChainType)] = true
	}

	// Get all addresses from transactions
	addressesByChain := make(map[types.ChainType][]string)
	for chainType := range chainTypes {
		addressesByChain[chainType] = []string{}
	}

	for _, tx := range page.Items {
		chainType := types.ChainType(tx.ChainType)
		addressesByChain[chainType] = append(
			addressesByChain[chainType],
			tx.FromAddress,
			tx.ToAddress,
		)
	}

	// Get tokens for each chain
	tokensMap := make(map[string]*types.Token)
	for chainType, addresses := range addressesByChain {
		tokens, err := h.tokenService.ListTokensByAddresses(c.Request.Context(), chainType, addresses)
		if err != nil {
			c.Error(err)
			return
		}

		// Add tokens to the map
		for i := range tokens {
			tokensMap[tokens[i].Address] = &tokens[i]
		}
	}

	c.JSON(http.StatusOK, ToPagedResponse(page, tokensMap))
}
