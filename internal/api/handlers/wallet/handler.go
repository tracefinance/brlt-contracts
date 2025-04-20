package wallet

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
	"vault0/internal/api/utils"
	"vault0/internal/services/token"
	walletService "vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Handler handles wallet API requests
type Handler struct {
	walletService walletService.Service
	tokenService  token.Service
}

// NewHandler creates a new wallet handler
func NewHandler(walletService walletService.Service, tokenService token.Service) *Handler {
	return &Handler{
		walletService: walletService,
		tokenService:  tokenService,
	}
}

// ActivateTokenRequest represents the request body for activating a token for a wallet
// swagger:model ActivateTokenRequest
type ActivateTokenRequest struct {
	TokenAddress string `json:"token_address" binding:"required"`
}

func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Create error handler middleware
	errorHandler := middleares.NewErrorHandler(nil)

	// Apply middleware to wallet routes group
	walletRoutes := router.Group("/wallets")
	walletRoutes.Use(errorHandler.Middleware())

	// Setup routes
	walletRoutes.POST("", h.CreateWallet)
	walletRoutes.GET("", h.ListWallets)
	walletRoutes.GET("/:chain_type/:address", h.GetWallet)
	walletRoutes.PUT("/:chain_type/:address", h.UpdateWallet)
	walletRoutes.DELETE("/:chain_type/:address", h.DeleteWallet)
	walletRoutes.GET("/:chain_type/:address/balance", h.GetWalletBalance)
	walletRoutes.POST("/:chain_type/:address/activate-token", h.ActivateToken)
}

// CreateWallet handles wallet creation
// @Summary Create a new wallet
// @Description Create a new wallet with the given chain type and name
// @Tags wallets
// @Accept json
// @Produce json
// @Param wallet body CreateWalletRequest true "Wallet data to create"
// @Success 201 {object} WalletResponse "Created wallet details"
// @Failure 400 {object} errors.Vault0Error "Invalid request data"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets [post]
func (h *Handler) CreateWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Create wallet
	walletModel, err := h.walletService.Create(c.Request.Context(), req.ChainType, req.Name, req.Tags)
	if err != nil {
		c.Error(err)
		return
	}

	// Convert to response
	response := ToResponse(walletModel)

	// Write response
	c.JSON(http.StatusCreated, response)
}

// GetWallet handles retrieving a wallet by chain type and address
// @Summary Get a wallet
// @Description Get a wallet's details by chain type and address
// @Tags wallets
// @Produce json
// @Param chain_type path string true "Blockchain network type (e.g., ethereum, bitcoin)"
// @Param address path string true "Wallet address on the blockchain"
// @Success 200 {object} WalletResponse "Wallet details including balance"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address} [get]
func (h *Handler) GetWallet(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Get the wallet
	walletModel, err := h.walletService.GetByAddress(c.Request.Context(), chainType, address)
	if err != nil {
		c.Error(err)
		return
	}

	// Convert to response
	response := ToResponse(walletModel)

	// Write response
	c.JSON(http.StatusOK, response)
}

// UpdateWallet handles updating a wallet
// @Summary Update a wallet
// @Description Update a wallet's name and tags by chain type and address
// @Tags wallets
// @Accept json
// @Produce json
// @Param chain_type path string true "Blockchain network type (e.g., ethereum, bitcoin)"
// @Param address path string true "Wallet address on the blockchain"
// @Param wallet body UpdateWalletRequest true "Wallet properties to update"
// @Success 200 {object} WalletResponse "Updated wallet details"
// @Failure 400 {object} errors.Vault0Error "Invalid request data"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address} [put]
func (h *Handler) UpdateWallet(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	var req UpdateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Update the wallet
	walletModel, err := h.walletService.Update(c.Request.Context(), chainType, address, req.Name, req.Tags)
	if err != nil {
		c.Error(err)
		return
	}

	// Convert to response
	response := ToResponse(walletModel)

	// Write response
	c.JSON(http.StatusOK, response)
}

// DeleteWallet handles deleting a wallet
// @Summary Delete a wallet
// @Description Delete a wallet by chain type and address
// @Tags wallets
// @Param chain_type path string true "Blockchain network type (e.g., ethereum, bitcoin)"
// @Param address path string true "Wallet address on the blockchain"
// @Success 204 "Wallet successfully deleted"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address} [delete]
func (h *Handler) DeleteWallet(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Delete the wallet
	err := h.walletService.Delete(c.Request.Context(), chainType, address)
	if err != nil {
		c.Error(err)
		return
	}

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}

// ListWallets handles listing wallets
// @Summary List wallets
// @Description Get a paginated list of all wallets
// @Tags wallets
// @Produce json
// @Param limit query int false "Maximum number of wallets to return (default: 10, 0 for all)" default(10)
// @Param next_token query string false "Token for retrieving the next page of results"
// @Success 200 {object} utils.PagedResponse[*WalletResponse] "Paginated list of wallets with navigation metadata"
// @Failure 400 {object} errors.Vault0Error "Invalid pagination token"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets [get]
func (h *Handler) ListWallets(c *gin.Context) {
	var req ListWalletsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	// Set default limit if not provided
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Get the wallets using token-based pagination
	walletPage, err := h.walletService.List(c.Request.Context(), limit, req.NextToken)
	if err != nil {
		c.Error(err)
		return
	}

	// Use utils.NewPagedResponse directly
	c.JSON(http.StatusOK, utils.NewPagedResponse(walletPage, ToResponse))
}

// GetWalletBalance handles retrieving a wallet's balances by chain type and address
// @Summary Get a wallet's balances
// @Description Get a wallet's native token and other token balances by chain type and address
// @Tags wallets
// @Produce json
// @Param chain_type path string true "Blockchain network type (e.g., ethereum, bitcoin)"
// @Param address path string true "Wallet address on the blockchain"
// @Success 200 {object} []TokenBalanceResponse "Array of token balances including native currency"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address}/balance [get]
func (h *Handler) GetWalletBalance(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Get the wallet balances
	balances, err := h.walletService.GetWalletBalancesByAddress(c.Request.Context(), chainType, address)
	if err != nil {
		c.Error(err)
		return
	}

	// Convert to response
	response := ToTokenBalanceResponseList(balances)

	// Write response
	c.JSON(http.StatusOK, response)
}

// ActivateToken handles activating a token for a wallet
// @Summary Activate a token for a wallet
// @Description Create a token balance for a wallet and token address
// @Tags wallets
// @Accept json
// @Produce json
// @Param chain_type path string true "Blockchain network type (e.g., ethereum, bitcoin)"
// @Param address path string true "Wallet address on the blockchain"
// @Param body body ActivateTokenRequest true "Token address to activate"
// @Success 204 "Token activated successfully"
// @Failure 400 {object} errors.Vault0Error "Invalid request data"
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address}/activate-token [post]
func (h *Handler) ActivateToken(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	var req ActivateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	err := h.walletService.ActivateToken(c.Request.Context(), chainType, address, req.TokenAddress)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
