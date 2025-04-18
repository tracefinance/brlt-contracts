package signer

import (
	"encoding/json"
	"strconv"
	"time"
	"vault0/internal/services/signer"
	"vault0/internal/types"
)

// CreateSignerRequest represents data needed to create a signer
type CreateSignerRequest struct {
	Name   string            `json:"name" binding:"required"`
	Type   signer.SignerType `json:"type" binding:"required,oneof=internal external"`
	UserID json.Number       `json:"user_id,omitempty" swaggertype:"string"`
}

// UpdateSignerRequest represents data for updating a signer
type UpdateSignerRequest struct {
	Name   string            `json:"name" binding:"required"`
	Type   signer.SignerType `json:"type" binding:"required,oneof=internal external"`
	UserID json.Number       `json:"user_id,omitempty" swaggertype:"string"`
}

// AddAddressRequest represents data for adding an address to a signer
type AddAddressRequest struct {
	ChainType string `json:"chain_type" binding:"required"`
	Address   string `json:"address" binding:"required"`
}

// SignerResponse represents a signer response
type SignerResponse struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	Type      string             `json:"type"`
	UserID    *string            `json:"user_id,omitempty"`
	Addresses []*AddressResponse `json:"addresses,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// AddressResponse represents an address response
type AddressResponse struct {
	ID        string    `json:"id"`
	SignerID  string    `json:"signer_id"`
	ChainType string    `json:"chain_type"`
	Address   string    `json:"address"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PagedSignersResponse represents a paginated list of signers
type PagedSignersResponse struct {
	Items   []*SignerResponse `json:"items"`
	Limit   int               `json:"limit"`
	Offset  int               `json:"offset"`
	HasMore bool              `json:"has_more"`
}

// ToAddressResponse converts an address model to a response
func ToAddressResponse(address *signer.Address) *AddressResponse {
	return &AddressResponse{
		ID:        strconv.FormatInt(address.ID, 10),
		SignerID:  strconv.FormatInt(address.SignerID, 10),
		ChainType: address.ChainType,
		Address:   address.Address,
		CreatedAt: address.CreatedAt,
		UpdatedAt: address.UpdatedAt,
	}
}

// ToAddressResponseList converts a slice of address models to a slice of responses
func ToAddressResponseList(addresses []*signer.Address) []*AddressResponse {
	responses := make([]*AddressResponse, len(addresses))
	for i, address := range addresses {
		responses[i] = ToAddressResponse(address)
	}
	return responses
}

// ToSignerResponse converts a signer model to a response
func ToSignerResponse(signer *signer.Signer) *SignerResponse {
	var userIDStr *string
	if signer.UserID != nil {
		uid := strconv.FormatInt(*signer.UserID, 10)
		userIDStr = &uid
	}

	return &SignerResponse{
		ID:        strconv.FormatInt(signer.ID, 10),
		Name:      signer.Name,
		Type:      string(signer.Type),
		UserID:    userIDStr,
		Addresses: ToAddressResponseList(signer.Addresses),
		CreatedAt: signer.CreatedAt,
		UpdatedAt: signer.UpdatedAt,
	}
}

// ToSignerResponseList converts a slice of signer models to a slice of responses
func ToSignerResponseList(signers []*signer.Signer) []*SignerResponse {
	responses := make([]*SignerResponse, len(signers))
	for i, signer := range signers {
		responses[i] = ToSignerResponse(signer)
	}
	return responses
}

// ToPagedResponse converts a Page of signer models to a PagedSignersResponse
func ToPagedResponse(page *types.Page[*signer.Signer]) *PagedSignersResponse {
	return &PagedSignersResponse{
		Items:   ToSignerResponseList(page.Items),
		Limit:   page.Limit,
		Offset:  page.Offset,
		HasMore: page.HasMore,
	}
}
