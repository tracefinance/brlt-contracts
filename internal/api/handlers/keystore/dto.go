package keystore

import (
	"fmt"
	"time"

	"vault0/internal/core/keystore"
	"vault0/internal/types"
)

// KeyResponse represents a key returned in API responses
type KeyResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      types.KeyType     `json:"type"`
	Curve     *string           `json:"curve,omitempty"`
	Tags      map[string]string `json:"tags,omitempty"`
	PublicKey *string           `json:"public_key,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// CreateKeyRequest represents a request to create a new key
type CreateKeyRequest struct {
	Name  string            `json:"name" binding:"required"`
	Type  types.KeyType     `json:"type" binding:"required"`
	Curve *string           `json:"curve,omitempty"`
	Tags  map[string]string `json:"tags"`
}

// ImportKeyRequest represents a request to import an existing key
type ImportKeyRequest struct {
	Name       string            `json:"name" binding:"required"`
	Type       types.KeyType     `json:"type" binding:"required"`
	Curve      *string           `json:"curve,omitempty"`
	PrivateKey string            `json:"private_key" binding:"required"`
	PublicKey  *string           `json:"public_key,omitempty"`
	Tags       map[string]string `json:"tags"`
}

// SignDataRequest represents a request to sign data with a key
type SignDataRequest struct {
	Data    string `json:"data" binding:"required"`
	RawData *bool  `json:"raw_data,omitempty"`
}

// SignDataResponse represents the response to a sign data request
type SignDataResponse struct {
	Signature string `json:"signature"`
}

// UpdateKeyRequest represents a request to update a key's metadata
type UpdateKeyRequest struct {
	Name string            `json:"name" binding:"required"`
	Tags map[string]string `json:"tags,omitempty"`
}

// ListKeysRequest defines the query parameters for listing keys
type ListKeysRequest struct {
	KeyType   string   `form:"key_type" binding:"omitempty"`
	Tags      []string `form:"tag" binding:"omitempty"`
	Limit     *int     `form:"limit" binding:"omitempty,min=1"`
	NextToken string   `form:"next_token" binding:"omitempty"`
}

// Convert a keystore.Key to a KeyResponse
func toResponse(key *keystore.Key) KeyResponse {
	var publicKeyStr *string
	if key.PublicKey != nil {
		str := fmt.Sprintf("%x", key.PublicKey)
		publicKeyStr = &str
	}

	var curveName *string
	if key.Curve != nil {
		curveName = &key.Curve.Params().Name
	}

	return KeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		Type:      key.Type,
		Curve:     curveName,
		Tags:      key.Tags,
		PublicKey: publicKeyStr,
		CreatedAt: key.CreatedAt,
	}
}
