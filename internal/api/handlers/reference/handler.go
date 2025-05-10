package reference

import (
	"net/http"
	"sort"
	"strconv"

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
		refRoutes.GET("/native-tokens", h.ListNativeTokens)
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
			ID:          strconv.FormatInt(chain.ID, 10),
			Type:        chain.Type,
			Layer:       chain.Layer,
			Name:        chain.Name,
			Symbol:      chain.Symbol,
			ExplorerURL: chain.ExplorerUrl,
		})
	}

	// Sort the response by ID for consistent ordering
	sort.Slice(response, func(i, j int) bool {
		return response[i].ID < response[j].ID
	})

	c.JSON(http.StatusOK, response)
}

// ListNativeTokens godoc
// @Summary      List native tokens
// @Description  Retrieves a list of native tokens for all supported blockchain networks.
// @Tags         Reference
// @Produce      json
// @Success      200  {array}  TokenResponse  "A list of native tokens for supported blockchains"
// @Router       /references/native-tokens [get]
func (h *Handler) ListNativeTokens(c *gin.Context) {
	chainList := h.chains.List()
	response := make([]TokenResponse, 0, len(chainList))

	for _, chain := range chainList {
		token, err := types.NewNativeToken(chain.Type)
		if err != nil {
			// Skip chains with errors creating native tokens
			continue
		}

		response = append(response, TokenResponse{
			Address:   token.Address,
			ChainType: token.ChainType,
			Symbol:    token.Symbol,
			Decimals:  token.Decimals,
			Type:      token.Type,
		})
	}

	// Sort the response by chain type for consistent ordering
	sort.Slice(response, func(i, j int) bool {
		return string(response[i].ChainType) < string(response[j].ChainType)
	})

	c.JSON(http.StatusOK, response)
}
