package vault

import (
	"time"
	"vault0/internal/services/vault"
)

// CreateVaultRequest represents the request payload for creating a new vault
type CreateVaultRequest struct {
	Name               string   `json:"name" binding:"required"`
	RecoveryAddress    string   `json:"recovery_address" binding:"required"`
	SignerAddresses    []string `json:"signer_addresses" binding:"required"`
	SignatureThreshold int      `json:"signature_threshold" binding:"required,min=1"`
	WhitelistedTokens  []string `json:"whitelisted_tokens"`
}

// ListVaultsRequest represents the request payload for listing vaults with pagination
type ListVaultsRequest struct {
	Limit     *int   `form:"limit"`
	NextToken string `form:"next_token"`
	Address   string `form:"address"`
	Status    string `form:"status"`
}

// UpdateVaultRequest represents the request payload for updating a vault's name
type UpdateVaultRequest struct {
	Name string `json:"name" binding:"required"`
}

// TokenRequest represents the request payload for adding or removing a token
type TokenRequest struct {
	Address string `json:"address" binding:"required"`
}

// VaultResponse represents a vault in API responses
type VaultResponse struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	ContractName     string    `json:"contract_name"`
	WalletID         int64     `json:"wallet_id"`
	ChainType        string    `json:"chain_type"`
	RecoveryAddress  string    `json:"recovery_address"`
	Signers          []string  `json:"signers"`
	Address          string    `json:"address,omitempty"`
	Status           string    `json:"status"`
	Quorum           int       `json:"quorum"`
	InRecovery       bool      `json:"in_recovery"`
	RecoveryDeadline *string   `json:"recovery_deadline,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// TokenAddedResponse represents the response after adding a token
type TokenAddedResponse struct {
	VaultID      int64  `json:"vault_id"`
	TokenAddress string `json:"token_address"`
	TxHash       string `json:"tx_hash"`
}

// TokenRemovedResponse represents the response after removing a token
type TokenRemovedResponse struct {
	VaultID      int64  `json:"vault_id"`
	TokenAddress string `json:"token_address"`
	TxHash       string `json:"tx_hash"`
}

// RecoveryResponse represents the response after a recovery action
type RecoveryResponse struct {
	VaultID           int64      `json:"vault_id"`
	Status            string     `json:"status"`
	Action            string     `json:"action"`
	TxHash            string     `json:"tx_hash"`
	RecoveryInitiated *time.Time `json:"recovery_initiated,omitempty"`
	ExecutableAfter   *time.Time `json:"executable_after,omitempty"`
}

// ToVaultResponse converts a vault domain model to API response model
func ToVaultResponse(v *vault.Vault) *VaultResponse {
	var signers []string

	var recoveryDeadline *string
	if v.RecoveryRequestTimestamp != nil {
		deadline := v.RecoveryRequestTimestamp.Add(72 * time.Hour).Format(time.RFC3339)
		recoveryDeadline = &deadline
	}

	return &VaultResponse{
		ID:               v.ID,
		Name:             v.Name,
		ContractName:     v.ContractName,
		WalletID:         v.WalletID,
		ChainType:        v.ChainType,
		RecoveryAddress:  v.RecoveryAddress,
		Signers:          signers,
		Address:          v.Address,
		Status:           string(v.Status),
		Quorum:           v.Quorum,
		InRecovery:       v.Status == vault.VaultStatusRecovering,
		RecoveryDeadline: recoveryDeadline,
		CreatedAt:        v.CreatedAt,
		UpdatedAt:        v.UpdatedAt,
	}
}

// ToVaultFilter converts ListVaultsRequest to VaultFilter
func ToVaultFilter(req *ListVaultsRequest) vault.VaultFilter {
	filter := vault.VaultFilter{}

	if req.Status != "" {
		status := vault.VaultStatus(req.Status)
		filter.Status = &status
	}

	if req.Address != "" {
		filter.Address = &req.Address
	}

	return filter
}
