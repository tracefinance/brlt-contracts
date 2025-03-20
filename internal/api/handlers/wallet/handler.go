package wallet

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
	"vault0/internal/services/wallet"
	"vault0/internal/types"
)

// Handler handles wallet API requests
type Handler struct {
	walletService wallet.Service
}

// NewHandler creates a new wallet handler
func NewHandler(walletService wallet.Service) *Handler {
	return &Handler{
		walletService: walletService,
	}
}

func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Create error handler middleware
	errorHandler := middleares.NewErrorHandler(nil)

	// Apply middleware to wallet routes group
	walletRoutes := router.Group("/wallets")
	walletRoutes.Use(errorHandler.Middleware())

	// Setup routes
	walletRoutes.POST("", h.CreateWallet)
	walletRoutes.GET("/:chain_type/:address", h.GetWallet)
	walletRoutes.PUT("/:chain_type/:address", h.UpdateWallet)
	walletRoutes.DELETE("/:chain_type/:address", h.DeleteWallet)
	walletRoutes.GET("", h.ListWallets)
	walletRoutes.GET("/:chain_type/:address/balance", h.GetWalletBalance)
}

// CreateWallet handles wallet creation
// @Summary Create a new wallet
// @Description Create a new wallet with the given chain type and name
// @Tags wallets
// @Accept json
// @Produce json
// @Param wallet body CreateWalletRequest true "Wallet data"
// @Success 201 {object} WalletResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
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
// @Description Get a wallet by chain type and address
// @Tags wallets
// @Produce json
// @Param chain_type path string true "Chain type"
// @Param address path string true "Wallet address"
// @Success 200 {object} WalletResponse
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
// @Description Update a wallet's name and tags
// @Tags wallets
// @Accept json
// @Produce json
// @Param chain_type path string true "Chain type"
// @Param address path string true "Wallet address"
// @Param wallet body UpdateWalletRequest true "Wallet data to update"
// @Success 200 {object} WalletResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
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
// @Param chain_type path string true "Chain type"
// @Param address path string true "Wallet address"
// @Success 204 "No Content"
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
// @Description Get a paginated list of wallets
// @Tags wallets
// @Produce json
// @Param limit query int false "Number of items to return (default: 10)" default(10)
// @Param offset query int false "Number of items to skip (default: 0)" default(0)
// @Success 200 {object} PagedWalletsResponse
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets [get]
func (h *Handler) ListWallets(c *gin.Context) {
	// Get pagination parameters
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")

	limit := 10 // Default limit
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0 // Default offset
	if offsetStr != "" {
		parsedOffset, err := strconv.Atoi(offsetStr)
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Get the wallets
	walletPage, err := h.walletService.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.Error(err)
		return
	}

	// Write response
	c.JSON(http.StatusOK, ToPagedResponse(walletPage))
}

// GetWalletBalance handles retrieving a wallet's balances by chain type and address
// @Summary Get a wallet's balances
// @Description Get a wallet's native and token balances by chain type and address
// @Tags wallets
// @Produce json
// @Param chain_type path string true "Chain type"
// @Param address path string true "Wallet address"
// @Success 200 {object} WalletBalancesResponse
// @Failure 404 {object} errors.Vault0Error "Wallet not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /wallets/{chain_type}/{address}/balance [get]
func (h *Handler) GetWalletBalance(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Get the wallet
	wallet, err := h.walletService.GetByAddress(c.Request.Context(), chainType, address)
	if err != nil {
		c.Error(err)
		return
	}

	// Get the wallet balances
	balances, err := h.walletService.GetWalletBalancesByAddress(c.Request.Context(), chainType, address)
	if err != nil {
		c.Error(err)
		return
	}

	// Convert to response
	response := ToWalletBalancesResponse(wallet, balances)

	// Write response
	c.JSON(http.StatusOK, response)
}
