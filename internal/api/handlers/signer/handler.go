package signer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	_ "vault0/internal/api/docs" // Required for Swagger documentation
	"vault0/internal/api/middleares"
	"vault0/internal/api/utils"
	"vault0/internal/errors"
	"vault0/internal/services/signer"
)

// Handler handles signer-related HTTP requests
type Handler struct {
	signerService signer.Service
}

// NewHandler creates a new signer handler
func NewHandler(signerService signer.Service) *Handler {
	return &Handler{
		signerService: signerService,
	}
}

// SetupRoutes configures the routes for the signer API
func (h *Handler) SetupRoutes(router *gin.RouterGroup) {
	// Create error handler middleware
	errorHandler := middleares.NewErrorHandler(nil)

	// Apply middleware to signer routes group
	signerRoutes := router.Group("/signers")
	signerRoutes.Use(errorHandler.Middleware())

	// Setup routes
	signerRoutes.POST("", h.CreateSigner)
	signerRoutes.PUT("/:id", h.UpdateSigner)
	signerRoutes.DELETE("/:id", h.DeleteSigner)
	signerRoutes.GET("/:id", h.GetSigner)
	signerRoutes.GET("", h.ListSigners)
	signerRoutes.GET("/user/:userId", h.GetSignersByUser)

	// Address routes
	signerRoutes.POST("/:id/addresses", h.AddAddress)
	signerRoutes.DELETE("/:id/addresses/:addressId", h.DeleteAddress)
	signerRoutes.GET("/:id/addresses", h.GetAddresses)
}

// CreateSigner handles POST /signers
// @Summary Create a new signer
// @Description Create a new signer with the given name and type
// @Tags signers
// @Accept json
// @Produce json
// @Param signer body CreateSignerRequest true "Signer data"
// @Success 201 {object} SignerResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers [post]
func (h *Handler) CreateSigner(c *gin.Context) {
	var req CreateSignerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Convert UserID from json.Number to *int64
	var userID *int64
	if req.UserID != "" {
		id, err := req.UserID.Int64()
		if err != nil {
			c.Error(errors.NewInvalidInputError("Invalid user ID format", "user_id", req.UserID))
			return
		}
		userID = &id
	}

	createdSigner, err := h.signerService.Create(c.Request.Context(), req.Name, req.Type, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ToSignerResponse(createdSigner))
}

// UpdateSigner handles PUT /signers/:id
// @Summary Update a signer
// @Description Update a signer's information by ID
// @Tags signers
// @Accept json
// @Produce json
// @Param id path int true "Signer ID"
// @Param signer body UpdateSignerRequest true "Signer data to update"
// @Success 200 {object} SignerResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Signer not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers/{id} [put]
func (h *Handler) UpdateSigner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidInputError("Invalid signer ID format", "id", idStr))
		return
	}

	var req UpdateSignerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	// Convert UserID from json.Number to *int64
	var userID *int64
	if req.UserID != "" {
		userId, err := req.UserID.Int64()
		if err != nil {
			c.Error(errors.NewInvalidInputError("Invalid user ID format", "user_id", req.UserID))
			return
		}
		userID = &userId
	}

	updatedSigner, err := h.signerService.Update(c.Request.Context(), id, req.Name, req.Type, userID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToSignerResponse(updatedSigner))
}

// DeleteSigner handles DELETE /signers/:id
// @Summary Delete a signer
// @Description Delete a signer by ID
// @Tags signers
// @Param id path int true "Signer ID"
// @Success 204 "No Content"
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Signer not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers/{id} [delete]
func (h *Handler) DeleteSigner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidInputError("Invalid signer ID format", "id", idStr))
		return
	}

	err = h.signerService.Delete(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetSigner handles GET /signers/:id
// @Summary Get a signer
// @Description Get a signer by ID
// @Tags signers
// @Produce json
// @Param id path int true "Signer ID"
// @Success 200 {object} SignerResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Signer not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers/{id} [get]
func (h *Handler) GetSigner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidInputError("Invalid signer ID format", "id", idStr))
		return
	}

	foundSigner, err := h.signerService.Get(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToSignerResponse(foundSigner))
}

// ListSigners handles GET /signers
// @Summary List signers
// @Description Get a paginated list of signers
// @Tags signers
// @Produce json
// @Param limit query int false "Number of items to return (default: 10, max: 100)" default(10)
// @Param next_token query string false "Token for fetching the next page"
// @Success 200 {object} docs.SignerPagedResponse
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers [get]
func (h *Handler) ListSigners(c *gin.Context) {
	var req ListSignersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(errors.NewInvalidParameterError("query", "invalid query parameters format or value"))
		return
	}

	// Set default limit if not provided
	limit := 10
	if req.Limit != nil {
		limit = *req.Limit
	}

	page, err := h.signerService.List(c.Request.Context(), limit, req.NextToken)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, utils.NewPagedResponse(page, ToSignerResponse))
}

// GetSignersByUser handles GET /signers/user/:userId
// @Summary Get signers by user
// @Description Get all signers associated with a specific user ID
// @Tags signers
// @Produce json
// @Param userId path int true "User ID"
// @Success 200 {array} SignerResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "User not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers/user/{userId} [get]
func (h *Handler) GetSignersByUser(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidInputError("Invalid user ID format", "userId", userIdStr))
		return
	}

	signers, err := h.signerService.GetByUserID(c.Request.Context(), userId)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToSignerResponseList(signers))
}

// AddAddress handles POST /signers/:id/addresses
// @Summary Add an address to a signer
// @Description Add a new blockchain address to an existing signer
// @Tags signers,addresses
// @Accept json
// @Produce json
// @Param id path int true "Signer ID"
// @Param address body AddAddressRequest true "Address data"
// @Success 201 {object} AddressResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Signer not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers/{id}/addresses [post]
func (h *Handler) AddAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidInputError("Invalid signer ID format", "id", idStr))
		return
	}

	var req AddAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	address, err := h.signerService.AddAddress(c.Request.Context(), id, req.ChainType, req.Address)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ToAddressResponse(address))
}

// DeleteAddress handles DELETE /signers/:id/addresses/:addressId
// @Summary Delete an address
// @Description Delete an address from a signer
// @Tags signers,addresses
// @Param id path int true "Signer ID"
// @Param addressId path int true "Address ID"
// @Success 204 "No Content"
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Address not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers/{id}/addresses/{addressId} [delete]
func (h *Handler) DeleteAddress(c *gin.Context) {
	addressIdStr := c.Param("addressId")
	addressId, err := strconv.ParseInt(addressIdStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidInputError("Invalid address ID format", "addressId", addressIdStr))
		return
	}

	err = h.signerService.DeleteAddress(c.Request.Context(), addressId)
	if err != nil {
		c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetAddresses handles GET /signers/:id/addresses
// @Summary Get signer addresses
// @Description Get all addresses associated with a signer
// @Tags signers,addresses
// @Produce json
// @Param id path int true "Signer ID"
// @Success 200 {array} AddressResponse
// @Failure 400 {object} errors.Vault0Error "Invalid request"
// @Failure 404 {object} errors.Vault0Error "Signer not found"
// @Failure 500 {object} errors.Vault0Error "Internal server error"
// @Router /signers/{id}/addresses [get]
func (h *Handler) GetAddresses(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidInputError("Invalid signer ID format", "id", idStr))
		return
	}

	addresses, err := h.signerService.GetAddresses(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToAddressResponseList(addresses))
}
