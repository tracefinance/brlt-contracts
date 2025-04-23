package vault

import (
	"time"
	"vault0/internal/types"
	// errors can be imported if needed by other files in this package
)

// VaultStatus represents the current state of a vault
type VaultStatus string

const (
	VaultStatusPending    VaultStatus = "pending"
	VaultStatusDeploying  VaultStatus = "deploying"
	VaultStatusActive     VaultStatus = "active"
	VaultStatusRecovering VaultStatus = "recovering"
	VaultStatusRecovered  VaultStatus = "recovered"
	VaultStatusPaused     VaultStatus = "paused"
	VaultStatusFailed     VaultStatus = "failed"
)

// Vault represents a multi-signature wallet vault
type Vault struct {
	ID                       int64           `db:"id"`
	Name                     string          `db:"name"`
	ContractName             string          `db:"contract_name"`
	WalletID                 int64           `db:"wallet_id"`
	ChainType                string          `db:"chain_type"`
	TxHash                   string          `db:"tx_hash"`
	RecoveryAddress          string          `db:"recovery_address"`
	Signers                  types.JSONArray `db:"signers"`
	Address                  string          `db:"address"`
	Status                   VaultStatus     `db:"status"`
	Quorum                   int             `db:"quorum"`
	RecoveryRequestTimestamp *time.Time      `db:"recovery_request_timestamp"`
	FailureReason            *string         `db:"failure_reason"`
	CreatedAt                time.Time       `db:"created_at"`
	UpdatedAt                time.Time       `db:"updated_at"`
	DeletedAt                *time.Time      `db:"deleted_at"`
}
