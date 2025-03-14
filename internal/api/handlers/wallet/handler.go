package wallet

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

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

// CreateWallet handles wallet creation
func (h *Handler) CreateWallet(c *gin.Context) {
	var req CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Create wallet
	walletModel, err := h.walletService.Create(c.Request.Context(), req.ChainType, req.Name, req.Tags)
	if err != nil {
		if errors.Is(err, wallet.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
		return
	}

	// Convert to response
	response := ToResponse(walletModel)

	// Write response
	c.JSON(http.StatusCreated, response)
}

// GetWallet handles retrieving a wallet by chain type and address
func (h *Handler) GetWallet(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Get the wallet
	walletModel, err := h.walletService.Get(c.Request.Context(), chainType, address)
	if err != nil {
		if errors.Is(err, wallet.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		}
		if errors.Is(err, wallet.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get wallet"})
		return
	}

	// Convert to response
	response := ToResponse(walletModel)

	// Write response
	c.JSON(http.StatusOK, response)
}

// UpdateWallet handles updating a wallet
func (h *Handler) UpdateWallet(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	var req UpdateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Update the wallet
	walletModel, err := h.walletService.Update(c.Request.Context(), chainType, address, req.Name, req.Tags)
	if err != nil {
		if errors.Is(err, wallet.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		}
		if errors.Is(err, wallet.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet"})
		return
	}

	// Convert to response
	response := ToResponse(walletModel)

	// Write response
	c.JSON(http.StatusOK, response)
}

// DeleteWallet handles deleting a wallet
func (h *Handler) DeleteWallet(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))
	address := c.Param("address")

	// Delete the wallet
	err := h.walletService.Delete(c.Request.Context(), chainType, address)
	if err != nil {
		if errors.Is(err, wallet.ErrWalletNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Wallet not found"})
			return
		}
		if errors.Is(err, wallet.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete wallet"})
		return
	}

	// Return 204 No Content
	c.Status(http.StatusNoContent)
}

// ListWallets handles listing wallets
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
	wallets, err := h.walletService.List(c.Request.Context(), limit, offset)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no wallets found, return an empty list
			c.JSON(http.StatusOK, ListWalletsResponse{
				Wallets: []*WalletResponse{},
				Total:   0,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list wallets"})
		return
	}

	// Convert to response
	response := ListWalletsResponse{
		Wallets: ToResponseList(wallets),
		Total:   len(wallets),
	}

	// Write response
	c.JSON(http.StatusOK, response)
}

func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	walletRoutes := router.Group("/wallets")
	walletRoutes.POST("", h.CreateWallet)
	walletRoutes.GET("/:chain_type/:address", h.GetWallet)
	walletRoutes.PUT("/:chain_type/:address", h.UpdateWallet)
	walletRoutes.DELETE("/:chain_type/:address", h.DeleteWallet)
	walletRoutes.GET("", h.ListWallets)
}
