package blockchain

import (
	"net/http"
	"vault0/internal/services/blockchain"
	"vault0/internal/types"

	"github.com/gin-gonic/gin"
)

// Handler handles blockchain-related HTTP requests
type Handler struct {
	service blockchain.Service
}

// NewHandler creates a new blockchain handler
func NewHandler(service blockchain.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// ActivateBlockchain handles POST /blockchains/:chain_type/activate
func (h *Handler) ActivateBlockchain(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))

	if err := h.service.Activate(c.Request.Context(), chainType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	blockchain, err := h.service.Get(c.Request.Context(), chainType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToResponse(blockchain))
}

// DeactivateBlockchain handles POST /blockchains/:chain_type/deactivate
func (h *Handler) DeactivateBlockchain(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))

	if err := h.service.Deactivate(c.Request.Context(), chainType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	blockchain, err := h.service.Get(c.Request.Context(), chainType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToResponse(blockchain))
}

// GetBlockchain handles GET /blockchains/:chain_type
func (h *Handler) GetBlockchain(c *gin.Context) {
	chainType := types.ChainType(c.Param("chain_type"))

	blockchain, err := h.service.Get(c.Request.Context(), chainType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if blockchain == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "blockchain not found"})
		return
	}

	c.JSON(http.StatusOK, ToResponse(blockchain))
}

// ListActiveBlockchains handles GET /blockchains
func (h *Handler) ListActiveBlockchains(c *gin.Context) {
	blockchains, err := h.service.ListActive(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ToResponseList(blockchains))
}

func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Register blockchain routes
	blockchainRoutes := router.Group("/blockchains")
	blockchainRoutes.POST("/:chain_type/activate", h.ActivateBlockchain)
	blockchainRoutes.POST("/:chain_type/deactivate", h.DeactivateBlockchain)
	blockchainRoutes.GET("/:chain_type", h.GetBlockchain)
	blockchainRoutes.GET("", h.ListActiveBlockchains)
}
