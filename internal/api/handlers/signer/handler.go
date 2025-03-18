package signer

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"vault0/internal/api/middleares"
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
func (h *Handler) CreateSigner(c *gin.Context) {
	var req CreateSignerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid request body"))
		return
	}

	createdSigner, err := h.signerService.Create(c.Request.Context(), req.Name, req.Type, req.UserID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, ToSignerResponse(createdSigner))
}

// UpdateSigner handles PUT /signers/:id
func (h *Handler) UpdateSigner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid signer ID format"))
		return
	}

	var req UpdateSignerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid request body"))
		return
	}

	updatedSigner, err := h.signerService.Update(c.Request.Context(), id, req.Name, req.Type, req.UserID)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToSignerResponse(updatedSigner))
}

// DeleteSigner handles DELETE /signers/:id
func (h *Handler) DeleteSigner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid signer ID format"))
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
func (h *Handler) GetSigner(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid signer ID format"))
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
func (h *Handler) ListSigners(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	pagedSigners, err := h.signerService.List(c.Request.Context(), limit, offset)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToPagedResponse(pagedSigners))
}

// GetSignersByUser handles GET /signers/user/:userId
func (h *Handler) GetSignersByUser(c *gin.Context) {
	userIdStr := c.Param("userId")
	userId, err := strconv.ParseInt(userIdStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid user ID format"))
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
func (h *Handler) AddAddress(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid signer ID format"))
		return
	}

	var req AddAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid request body"))
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
func (h *Handler) DeleteAddress(c *gin.Context) {
	addressIdStr := c.Param("addressId")
	addressId, err := strconv.ParseInt(addressIdStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid address ID format"))
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
func (h *Handler) GetAddresses(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.Error(errors.NewInvalidRequestError("Invalid signer ID format"))
		return
	}

	addresses, err := h.signerService.GetAddresses(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, ToAddressResponseList(addresses))
}
