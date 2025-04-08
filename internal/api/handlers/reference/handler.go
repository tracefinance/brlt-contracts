package reference

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"vault0/internal/types"
)

// Handler manages reference data API endpoints.
type Handler struct {
	chains *types.Chains
}

// NewHandler creates a new reference handler instance.
func NewHandler(chains *types.Chains) *Handler {
	return &Handler{chains: chains}
}

// SetupRoutes configures routes for reference data operations.
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	refRoutes := router.Group("/references")
	{
		refRoutes.GET("/chains", h.ListChains)
	}
}

// ListChains godoc
// @Summary      List supported blockchains
// @Description  Retrieves a list of blockchain networks supported and configured in the system.
// @Tags         Reference
// @Produce      json
// @Success      200  {array}  ChainResponse  "A list of supported blockchain configurations"
// @Router       /references/chains [get]
func (h *Handler) ListChains(c *gin.Context) {
	chainList := h.chains.List()
	response := make([]ChainResponse, 0, len(chainList))

	for _, chain := range chainList {
		response = append(response, ChainResponse{
			ID:          chain.ID,
			Type:        chain.Type,
			Layer:       chain.Layer,
			Name:        chain.Name,
			Symbol:      chain.Symbol,
			ExplorerURL: chain.ExplorerUrl,
		})
	}

	c.JSON(http.StatusOK, response)
}
