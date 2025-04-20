package transaction

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "vault0/internal/api/docs" // Required for Swagger documentation
	"vault0/internal/api/middleares"
	"vault0/internal/api/utils"
	_ "vault0/internal/errors" // Required for Swagger documentation
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
	walletRoutes := router.Group("/wallets/:chain_type/:address/transactions")
	walletRoutes.Use(errorHandler.Middleware())
	walletRoutes.GET("", h.GetTransactionsByAddress)
	walletRoutes.GET("/:hash", h.GetTransaction)

	// Direct transaction routes
	transactionRoutes := router.Group("/transactions")
	transactionRoutes.Use(errorHandler.Middleware())
	transactionRoutes.GET("/:hash", h.GetTransaction)
	transactionRoutes.GET("", h.FilterTransactions)
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
// @Success 200 {object} docs.TransactionPagedResponse
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

	// Get token from tx.TokenAddress
	token, err := h.tokenService.GetToken(c.Request.Context(), tx.TokenAddress)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToResponse(tx, token))
}

// GetTransactionsByAddress handles GET /wallets/:chain_type/:address/transactions
// @Summary List transactions for an address
// @Description Get a paginated list of transactions for a specific wallet address
// @Tags transactions
// @Produce json
// @Param chain_type path string true "Chain type"
// @Param address path string true "Wallet address"
// @Param limit query int false "Number of items to return (default: 10)" default(10)
// @Param next_token query string false "Token for fetching the next page"
// @Param token_address query string false "Filter transactions by token address (use 'native' for native transactions)"
// @Success 200 {object} docs.TransactionPagedResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address}/transactions [get]
func (h *Handler) GetTransactionsByAddress(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Parse pagination parameters
	var req ListTransactionsByAddressRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	// Set default limit if not provided
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Create a filter with the chain type and address
	chainTypeVal := chainType
	addressVal := address

	filter := &transaction.Filter{
		ChainType: &chainTypeVal,
		Address:   &addressVal,
	}

	// Add token address filter if provided
	if req.TokenAddress != "" {
		tokenAddressVal := req.TokenAddress
		filter.TokenAddress = &tokenAddressVal
	}

	// Use filter-based transaction retrieval
	page, err := h.transactionService.FilterTransactions(c.Request.Context(), filter, limit, req.NextToken)
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

	// Create a transform function for the paged response
	transformFunc := func(tx *transaction.Transaction) TransactionResponse {
		// Get the token from the map or use a default
		token, ok := tokensMap[tx.TokenAddress]
		if !ok {
			// Fallback to direct conversion if token not found
			token = &types.Token{Decimals: 18} // Default to 18 decimals
		}
		return ToResponse(tx, token)
	}

	c.JSON(http.StatusOK, utils.NewPagedResponse(page, transformFunc))
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
// @Param next_token query string false "Token for fetching the next page"
// @Success 200 {object} docs.TransactionPagedResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /transactions [get]
func (h *Handler) FilterTransactions(c *gin.Context) {
	// Parse pagination and filter parameters
	var req ListTransactionsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	// Set default limit if not provided
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Create a filter with the provided parameters
	filter := &transaction.Filter{}

	// Apply chain type filter if provided
	if req.ChainType != "" {
		chainType := types.ChainType(req.ChainType)
		filter.ChainType = &chainType
	}

	// Apply address filter if provided
	if req.Address != "" {
		filter.Address = &req.Address
	}

	// Apply token address filter if provided
	if req.TokenAddress != "" {
		filter.TokenAddress = &req.TokenAddress
	}

	// Apply status filter if provided
	if req.Status != "" {
		filter.Status = &req.Status
	}

	// Get transactions with the applied filters
	page, err := h.transactionService.FilterTransactions(c.Request.Context(), filter, limit, req.NextToken)
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

	// Create a transform function for the paged response
	transformFunc := func(tx *transaction.Transaction) TransactionResponse {
		// Get the token from the map or use a default
		token, ok := tokensMap[tx.TokenAddress]
		if !ok {
			// Fallback to direct conversion if token not found
			token = &types.Token{Decimals: 18} // Default to 18 decimals
		}
		return ToResponse(tx, token)
	}

	c.JSON(http.StatusOK, utils.NewPagedResponse(page, transformFunc))
}
