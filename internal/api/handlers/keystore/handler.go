package keystore

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
	"vault0/internal/api/utils"
	"vault0/internal/errors"
	keystoreSvc "vault0/internal/services/keystore"
	"vault0/internal/types"
)

// Handler manages keystore-related API endpoints
type Handler struct {
	service keystoreSvc.Service
}

// NewHandler creates a new keystore handler
func NewHandler(service keystoreSvc.Service) *Handler {
	return &Handler{service: service}
}

// SetupRoutes configures the keystore API routes
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	errorHandler := middleares.NewErrorHandler(nil)

	keystoreRoutes := router.Group("/keys")
	keystoreRoutes.Use(errorHandler.Middleware())
	keystoreRoutes.GET("", h.listKeys)
	keystoreRoutes.POST("", h.createKey)
	keystoreRoutes.POST("/import", h.importKey)
	keystoreRoutes.GET("/:id", h.getKey)
	keystoreRoutes.PUT("/:id", h.updateKey)
	keystoreRoutes.DELETE("/:id", h.deleteKey)
	keystoreRoutes.POST("/:id/sign", h.signData)
}

// listKeys handles GET /keys
// @Summary List keys
// @Description Get a list of cryptographic keys with optional filtering and pagination
// @Tags keys
// @Produce json
// @Param key_type query string false "Filter by key type (ECDSA, RSA, Ed25519, Symmetric)"
// @Param tag query string false "Filter by tag (in format key=value, can specify multiple)"
// @Param limit query int false "Maximum number of keys to return (default 50)"
// @Param next_token query string false "Token for pagination (empty for first page)"
// @Success 200 {object} utils.PagedResponse[KeyResponse]
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /keys [get]
func (h *Handler) listKeys(c *gin.Context) {
	var req ListKeysRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errors.NewInvalidParameterError("query", "invalid query parameters format or value"))
		return
	}

	// Set default limit if not provided
	limit := 10 // Default limit
	if req.Limit != nil {
		limit = *req.Limit
	}

	// Build filter
	filter := keystoreSvc.KeyFilter{}

	if req.KeyType != "" {
		keyType := types.KeyType(req.KeyType)
		filter.KeyType = &keyType
	}

	// Parse tag queries (format: key=value)
	if len(req.Tags) > 0 {
		filter.Tags = make(map[string]string)
		for _, tagQuery := range req.Tags {
			// Parse the key=value format
			parts := strings.SplitN(tagQuery, "=", 2)
			if len(parts) == 2 && parts[0] != "" {
				filter.Tags[parts[0]] = parts[1]
			}
		}
	}

	// Get keys with filtering and pagination
	keysPage, err := h.service.ListKeys(c.Request.Context(), filter, limit, req.NextToken)
	if err != nil {
		c.Error(err)
		return
	}

	// Use the generic PagedResponse utility
	response := utils.NewPagedResponse(keysPage, newKeyResponse)

	c.JSON(http.StatusOK, response)
}

// createKey handles POST /keys
// @Summary Create a new key
// @Description Generate a new cryptographic key
// @Tags keys
// @Accept json
// @Produce json
// @Param key body CreateKeyRequest true "Key details"
// @Success 201 {object} KeyResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /keys [post]
func (h *Handler) createKey(c *gin.Context) {
	var req CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Create the key
	key, err := h.service.CreateKey(c.Request.Context(), req.Name, req.Type, req.Curve, req.Tags)
	if err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := newKeyResponse(key)
	c.JSON(http.StatusCreated, response)
}

// importKey handles POST /keys/import
// @Summary Import an existing key
// @Description Import an existing cryptographic key
// @Tags keys
// @Accept json
// @Produce json
// @Param key body ImportKeyRequest true "Key details"
// @Success 201 {object} KeyResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /keys/import [post]
func (h *Handler) importKey(c *gin.Context) {
	var req ImportKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Decode the key material
	privateKey, err := utils.DecodeBytes(req.PrivateKey)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("private_key", "must be valid base64 encoded data"))
		return
	}

	var publicKey []byte
	if req.PublicKey != "" {
		publicKey, err = utils.DecodeBytes(req.PublicKey)
		if err != nil {
			c.Error(errors.NewInvalidParameterError("public_key", "must be valid base64 encoded data"))
			return
		}
	}

	// Import the key
	key, err := h.service.ImportKey(c.Request.Context(), req.Name, req.Type, req.Curve, privateKey, publicKey, req.Tags)
	if err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := newKeyResponse(key)
	c.JSON(http.StatusCreated, response)
}

// getKey handles GET /keys/:id
// @Summary Get key details
// @Description Get details of a specific key
// @Tags keys
// @Produce json
// @Param id path string true "Key ID"
// @Success 200 {object} KeyResponse
// @Failure 404 {object} errors.Vault0Error "Key not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /keys/{id} [get]
func (h *Handler) getKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(errors.NewMissingParameterError("id"))
		return
	}

	key, err := h.service.GetKey(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := newKeyResponse(key)
	c.JSON(http.StatusOK, response)
}

// updateKey handles PUT /keys/:id
// @Summary Update key metadata
// @Description Update the metadata of a specific key
// @Tags keys
// @Accept json
// @Produce json
// @Param id path string true "Key ID"
// @Param key body UpdateKeyRequest true "Key metadata"
// @Success 200 {object} KeyResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Key not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /keys/{id} [put]
func (h *Handler) updateKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(errors.NewMissingParameterError("id"))
		return
	}

	var req UpdateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Update the key
	key, err := h.service.UpdateKey(c.Request.Context(), id, req.Name, req.Tags)
	if err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := newKeyResponse(key)
	c.JSON(http.StatusOK, response)
}

// deleteKey handles DELETE /keys/:id
// @Summary Delete a key
// @Description Delete a specific key
// @Tags keys
// @Param id path string true "Key ID"
// @Success 204 "No Content"
// @Failure 404 {object} errors.Vault0Error "Key not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /keys/{id} [delete]
func (h *Handler) deleteKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(errors.NewMissingParameterError("id"))
		return
	}

	if err := h.service.DeleteKey(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// signData handles POST /keys/:id/sign
// @Summary Sign data with a key
// @Description Sign data using a specific key
// @Tags keys
// @Accept json
// @Produce json
// @Param id path string true "Key ID"
// @Param data body SignDataRequest true "Data to sign"
// @Success 200 {object} SignDataResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Key not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /keys/{id}/sign [post]
func (h *Handler) signData(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.Error(errors.NewMissingParameterError("id"))
		return
	}

	var req SignDataRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Decode the data
	data, err := utils.DecodeBytes(req.Data)
	if err != nil {
		c.Error(errors.NewInvalidParameterError("data", "must be valid base64 encoded data"))
		return
	}

	// Sign the data
	signature, err := h.service.SignData(c.Request.Context(), id, data, req.RawData)
	if err != nil {
		c.Error(err)
		return
	}

	// Build response
	response := SignDataResponse{
		Signature: utils.EncodeBytes(signature),
	}

	c.JSON(http.StatusOK, response)
}
