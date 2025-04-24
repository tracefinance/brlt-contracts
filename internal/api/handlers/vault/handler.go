package vault

import (
	"net/http"
	"strconv"
	"time"

	"vault0/internal/api/middleares"
	"vault0/internal/api/utils"
	"vault0/internal/errors"
	"vault0/internal/logger"
	"vault0/internal/services/vault"

	"github.com/gin-gonic/gin"
)

// Handler manages vault API endpoints
type Handler struct {
	service vault.Service
	log     logger.Logger
}

// NewHandler creates a new vault handler instance
func NewHandler(service vault.Service, log logger.Logger) *Handler {
	return &Handler{
		service: service,
		log:     log,
	}
}

// SetupRoutes configures routes for vault management
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Create error handler middleware
	errorHandler := middleares.NewErrorHandler(nil)

	// Apply middleware to vault routes group
	vaultsGroup := router.Group("/vaults")
	vaultsGroup.Use(errorHandler.Middleware())
	{
		vaultsGroup.POST("", h.CreateVault)
		vaultsGroup.GET("", h.ListVaults)
		vaultsGroup.GET("/:id", h.GetVault)
		vaultsGroup.PUT("/:id", h.UpdateVault)

		// Token management endpoints
		vaultsGroup.POST("/:id/tokens", h.AddToken)
		vaultsGroup.DELETE("/:id/tokens/:address", h.RemoveToken)

		// Recovery endpoints
		vaultsGroup.POST("/:id/recovery/start", h.StartRecovery)
		vaultsGroup.POST("/:id/recovery/cancel", h.CancelRecovery)
		vaultsGroup.POST("/:id/recovery/execute", h.ExecuteRecovery)
	}
}

// CreateVault handles POST /vaults requests
// @Summary Create a new vault
// @Description Create a new vault with specified parameters
// @Tags vaults
// @Accept json
// @Produce json
// @Param request body CreateVaultRequest true "Vault creation parameters"
// @Success 201 {object} VaultResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults [post]
func (h *Handler) CreateVault(c *gin.Context) {
	var req CreateVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind JSON for CreateVaultRequest",
			logger.Error(err),
			logger.String("endpoint", "CreateVault"))
		c.Error(errors.NewValidationError(map[string]any{
			"request": "Invalid request format",
		}))
		return
	}

	// Validate the create vault request
	if err := ValidateCreateVaultRequest(&req); err != nil {
		h.log.Error("Validation failed for CreateVaultRequest",
			logger.Error(err),
			logger.String("endpoint", "CreateVault"))
		c.Error(err)
		return
	}

	// For now, use a placeholder wallet ID
	// In a real implementation, this would come from authentication
	walletID := int64(1)

	h.log.Info("Creating vault",
		logger.String("name", req.Name),
		logger.Int64("wallet_id", walletID))

	vault, err := h.service.CreateVault(
		c.Request.Context(),
		walletID,
		req.Name,
		req.RecoveryAddress,
		req.SignerAddresses,
		req.SignatureThreshold,
		req.WhitelistedTokens,
	)

	if err != nil {
		h.log.Error("Failed to create vault",
			logger.Error(err),
			logger.String("name", req.Name),
			logger.Int64("wallet_id", walletID))
		c.Error(err)
		return
	}

	h.log.Info("Vault created successfully",
		logger.String("name", req.Name),
		logger.Int64("vault_id", vault.ID))

	c.JSON(http.StatusCreated, ToVaultResponse(vault))
}

// ListVaults handles GET /vaults requests
// @Summary List vaults
// @Description Get a paginated list of vaults
// @Tags vaults
// @Produce json
// @Param limit query int false "Number of items to return (default: 10, max: 100)" default(10)
// @Param next_token query string false "Token for fetching the next page"
// @Success 200 {object} docs.VaultPagedResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults [get]
func (h *Handler) ListVaults(c *gin.Context) {
	var req ListVaultsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		h.log.Error("Failed to bind query for ListVaultsRequest",
			logger.Error(err),
			logger.String("endpoint", "ListVaults"))
		c.Error(errors.NewValidationError(map[string]any{
			"query": "Invalid query parameters",
		}))
		return
	}

	// Set default limit if not provided
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	filter := ToVaultFilter(&req)

	h.log.Info("Listing vaults",
		logger.Int("limit", limit),
		logger.String("next_token", req.NextToken))

	page, err := h.service.ListVaults(c.Request.Context(), filter, limit, req.NextToken)
	if err != nil {
		h.log.Error("Failed to list vaults",
			logger.Error(err),
			logger.Int("limit", limit))
		c.Error(err)
		return
	}

	h.log.Info("Vaults listed successfully",
		logger.Int("count", len(page.Items)))

	c.JSON(http.StatusOK, utils.NewPagedResponse(page, ToVaultResponse))
}

// GetVault handles GET /vaults/:id requests
// @Summary Get a vault
// @Description Get a vault by ID
// @Tags vaults
// @Produce json
// @Param id path int true "Vault ID"
// @Success 200 {object} VaultResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Vault not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults/{id} [get]
func (h *Handler) GetVault(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Error("Invalid vault ID format",
			logger.Error(err),
			logger.String("vault_id_param", c.Param("id")))
		c.Error(errors.NewValidationError(map[string]any{
			"id": "Invalid vault ID format",
		}))
		return
	}

	h.log.Info("Getting vault", logger.Int64("vault_id", id))

	vault, err := h.service.GetVaultByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get vault",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	h.log.Info("Vault retrieved successfully", logger.Int64("vault_id", id))
	c.JSON(http.StatusOK, ToVaultResponse(vault))
}

// UpdateVault handles PUT /vaults/:id requests
// @Summary Update a vault
// @Description Update a vault's name by ID
// @Tags vaults
// @Accept json
// @Produce json
// @Param id path int true "Vault ID"
// @Param request body UpdateVaultRequest true "Vault update parameters"
// @Success 200 {object} VaultResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Vault not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults/{id} [put]
func (h *Handler) UpdateVault(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Error("Invalid vault ID format",
			logger.Error(err),
			logger.String("vault_id_param", c.Param("id")))
		c.Error(errors.NewValidationError(map[string]any{
			"id": "Invalid vault ID format",
		}))
		return
	}

	var req UpdateVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind JSON for UpdateVaultRequest",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(errors.NewValidationError(map[string]any{
			"request": "Invalid request format",
		}))
		return
	}

	if req.Name == "" {
		h.log.Error("Empty vault name in update request", logger.Int64("vault_id", id))
		c.Error(errors.NewValidationError(map[string]any{
			"name": "Vault name cannot be empty",
		}))
		return
	}

	h.log.Info("Updating vault name",
		logger.Int64("vault_id", id),
		logger.String("new_name", req.Name))

	vault, err := h.service.UpdateVault(c.Request.Context(), id, req.Name)
	if err != nil {
		h.log.Error("Failed to update vault name",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	h.log.Info("Vault name updated successfully", logger.Int64("vault_id", id))
	c.JSON(http.StatusOK, ToVaultResponse(vault))
}

// AddToken handles POST /vaults/:id/tokens requests
// @Summary Add token to vault
// @Description Add a token to the vault's whitelist
// @Tags vaults
// @Accept json
// @Produce json
// @Param id path int true "Vault ID"
// @Param request body TokenRequest true "Token address"
// @Success 200 {object} TokenAddedResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Vault not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults/{id}/tokens [post]
func (h *Handler) AddToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Error("Invalid vault ID format",
			logger.Error(err),
			logger.String("vault_id_param", c.Param("id")))
		c.Error(errors.NewValidationError(map[string]any{
			"id": "Invalid vault ID format",
		}))
		return
	}

	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("Failed to bind JSON for TokenRequest",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(errors.NewValidationError(map[string]any{
			"request": "Invalid request format",
		}))
		return
	}

	if err := ValidateTokenAddress(req.Address); err != nil {
		h.log.Error("Invalid token address",
			logger.Error(err),
			logger.String("token_address", req.Address))
		c.Error(err)
		return
	}

	h.log.Info("Adding token to vault",
		logger.Int64("vault_id", id),
		logger.String("token_address", req.Address))

	txHash, err := h.service.AddSupportedToken(c.Request.Context(), id, req.Address)
	if err != nil {
		h.log.Error("Failed to add token to vault",
			logger.Error(err),
			logger.Int64("vault_id", id),
			logger.String("token_address", req.Address))
		c.Error(err)
		return
	}

	h.log.Info("Token added successfully",
		logger.Int64("vault_id", id),
		logger.String("token_address", req.Address),
		logger.String("tx_hash", txHash))

	c.JSON(http.StatusOK, TokenAddedResponse{
		VaultID:      id,
		TokenAddress: req.Address,
		TxHash:       txHash,
	})
}

// RemoveToken handles DELETE /vaults/:id/tokens/:address requests
// @Summary Remove token from vault
// @Description Remove a token from the vault's whitelist
// @Tags vaults
// @Produce json
// @Param id path int true "Vault ID"
// @Param address path string true "Token address"
// @Success 200 {object} TokenRemovedResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Vault not found or token not whitelisted"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults/{id}/tokens/{address} [delete]
func (h *Handler) RemoveToken(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Error("Invalid vault ID format",
			logger.Error(err),
			logger.String("vault_id_param", c.Param("id")))
		c.Error(errors.NewValidationError(map[string]any{
			"id": "Invalid vault ID format",
		}))
		return
	}

	address := c.Param("address")
	if err := ValidateTokenAddress(address); err != nil {
		h.log.Error("Invalid token address",
			logger.Error(err),
			logger.String("token_address", address))
		c.Error(err)
		return
	}

	h.log.Info("Removing token from vault",
		logger.Int64("vault_id", id),
		logger.String("token_address", address))

	txHash, err := h.service.RemoveSupportedToken(c.Request.Context(), id, address)
	if err != nil {
		h.log.Error("Failed to remove token from vault",
			logger.Error(err),
			logger.Int64("vault_id", id),
			logger.String("token_address", address))
		c.Error(err)
		return
	}

	h.log.Info("Token removed successfully",
		logger.Int64("vault_id", id),
		logger.String("token_address", address),
		logger.String("tx_hash", txHash))

	c.JSON(http.StatusOK, TokenRemovedResponse{
		VaultID:      id,
		TokenAddress: address,
		TxHash:       txHash,
	})
}

// StartRecovery handles POST /vaults/:id/recovery/start requests
// @Summary Start vault recovery process
// @Description Initiate the recovery process for a vault
// @Tags vaults
// @Produce json
// @Param id path int true "Vault ID"
// @Success 200 {object} RecoveryResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Vault not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults/{id}/recovery/start [post]
func (h *Handler) StartRecovery(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Error("Invalid vault ID format",
			logger.Error(err),
			logger.String("vault_id_param", c.Param("id")))
		c.Error(errors.NewValidationError(map[string]any{
			"id": "Invalid vault ID format",
		}))
		return
	}

	h.log.Info("Starting recovery process", logger.Int64("vault_id", id))

	txHash, err := h.service.StartRecovery(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to start recovery process",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	// Fetch updated vault to get recovery timestamp
	vault, err := h.service.GetVaultByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get updated vault after starting recovery",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	var executableAfter *time.Time
	if vault.RecoveryRequestTimestamp != nil {
		t := vault.RecoveryRequestTimestamp.Add(72 * time.Hour)
		executableAfter = &t
	}

	h.log.Info("Recovery started successfully",
		logger.Int64("vault_id", id),
		logger.String("tx_hash", txHash))

	c.JSON(http.StatusOK, RecoveryResponse{
		VaultID:           id,
		Status:            string(vault.Status),
		Action:            "start",
		TxHash:            txHash,
		RecoveryInitiated: vault.RecoveryRequestTimestamp,
		ExecutableAfter:   executableAfter,
	})
}

// CancelRecovery handles POST /vaults/:id/recovery/cancel requests
// @Summary Cancel vault recovery process
// @Description Cancel an in-progress recovery process for a vault
// @Tags vaults
// @Produce json
// @Param id path int true "Vault ID"
// @Success 200 {object} RecoveryResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Vault not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults/{id}/recovery/cancel [post]
func (h *Handler) CancelRecovery(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Error("Invalid vault ID format",
			logger.Error(err),
			logger.String("vault_id_param", c.Param("id")))
		c.Error(errors.NewValidationError(map[string]any{
			"id": "Invalid vault ID format",
		}))
		return
	}

	h.log.Info("Canceling recovery process", logger.Int64("vault_id", id))

	txHash, err := h.service.CancelRecovery(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to cancel recovery process",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	// Fetch updated vault
	vault, err := h.service.GetVaultByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get updated vault after canceling recovery",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	h.log.Info("Recovery canceled successfully",
		logger.Int64("vault_id", id),
		logger.String("tx_hash", txHash))

	c.JSON(http.StatusOK, RecoveryResponse{
		VaultID: id,
		Status:  string(vault.Status),
		Action:  "cancel",
		TxHash:  txHash,
	})
}

// ExecuteRecovery handles POST /vaults/:id/recovery/execute requests
// @Summary Execute vault recovery
// @Description Execute a recovery process after the waiting period has passed
// @Tags vaults
// @Produce json
// @Param id path int true "Vault ID"
// @Success 200 {object} RecoveryResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request or recovery not ready for execution"
// @Failure 404 {object} errors.Vault0Error "Vault not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /vaults/{id}/recovery/execute [post]
func (h *Handler) ExecuteRecovery(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.Error("Invalid vault ID format",
			logger.Error(err),
			logger.String("vault_id_param", c.Param("id")))
		c.Error(errors.NewValidationError(map[string]any{
			"id": "Invalid vault ID format",
		}))
		return
	}

	h.log.Info("Executing recovery process", logger.Int64("vault_id", id))

	txHash, err := h.service.ExecuteRecovery(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to execute recovery process",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	// Fetch updated vault
	vault, err := h.service.GetVaultByID(c.Request.Context(), id)
	if err != nil {
		h.log.Error("Failed to get updated vault after executing recovery",
			logger.Error(err),
			logger.Int64("vault_id", id))
		c.Error(err)
		return
	}

	h.log.Info("Recovery executed successfully",
		logger.Int64("vault_id", id),
		logger.String("tx_hash", txHash))

	c.JSON(http.StatusOK, RecoveryResponse{
		VaultID: id,
		Status:  string(vault.Status),
		Action:  "execute",
		TxHash:  txHash,
	})
}
