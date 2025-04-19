package keystore

import (
	"vault0/internal/api/utils"
	"vault0/internal/core/keystore"
	"vault0/internal/types"
)

// KeyResponse represents a key returned in API responses
type KeyResponse struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Type      types.KeyType     `json:"type"`
	Tags      map[string]string `json:"tags,omitempty"`
	PublicKey string            `json:"public_key,omitempty"`
	CreatedAt int64             `json:"created_at"`
}

// CreateKeyRequest represents a request to create a new key
type CreateKeyRequest struct {
	Name  string            `json:"name" binding:"required"`
	Type  types.KeyType     `json:"type" binding:"required"`
	Curve string            `json:"curve"`
	Tags  map[string]string `json:"tags"`
}

// ImportKeyRequest represents a request to import an existing key
type ImportKeyRequest struct {
	Name       string            `json:"name" binding:"required"`
	Type       types.KeyType     `json:"type" binding:"required"`
	Curve      string            `json:"curve"`
	PrivateKey string            `json:"private_key" binding:"required"`
	PublicKey  string            `json:"public_key"`
	Tags       map[string]string `json:"tags"`
}

// SignDataRequest represents a request to sign data with a key
type SignDataRequest struct {
	Data    string `json:"data" binding:"required"`
	RawData bool   `json:"raw_data"`
}

// SignDataResponse represents the response to a sign data request
type SignDataResponse struct {
	Signature string `json:"signature"`
}

// UpdateKeyRequest represents a request to update a key's metadata
type UpdateKeyRequest struct {
	Name string            `json:"name"`
	Tags map[string]string `json:"tags"`
}

// KeyListResponse represents a list of keys in an API response
type KeyListResponse struct {
	Items []KeyResponse `json:"items"`
}

// Convert a keystore.Key to a KeyResponse
func newKeyResponse(key *keystore.Key) KeyResponse {
	var publicKeyStr string
	if key.PublicKey != nil {
		publicKeyStr = utils.EncodeBytes(key.PublicKey)
	}

	return KeyResponse{
		ID:        key.ID,
		Name:      key.Name,
		Type:      key.Type,
		Tags:      key.Tags,
		PublicKey: publicKeyStr,
		CreatedAt: key.CreatedAt,
	}
}
