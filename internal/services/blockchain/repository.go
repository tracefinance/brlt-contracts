package blockchain

import (
	"context"
	"fmt"
	"time"
	"vault0/internal/core/db"
	"vault0/internal/types"
)

// Repository defines the blockchain data access interface
type Repository interface {
	// Create creates a new blockchain record
	Create(ctx context.Context, blockchain *Blockchain) error
	// FindByChainType finds a blockchain by its chain type
	FindByChainType(ctx context.Context, chainType types.ChainType) (*Blockchain, error)
	// FindActive finds all active blockchains
	FindActive(ctx context.Context) ([]*Blockchain, error)
	// Deactivate sets the deactivated_at timestamp for a blockchain
	Deactivate(ctx context.Context, chainType types.ChainType) error
}

// repository implements Repository interface
type repository struct {
	db *db.DB
}

// NewRepository creates a new blockchain repository
func NewRepository(db *db.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, blockchain *Blockchain) error {
	query := `
		INSERT INTO blockchains (chain_type, chain_id, wallet_id, created_at)
		VALUES (?, ?, ?, ?)
	`

	_, err := r.db.ExecuteStatement(query, []interface{}{
		blockchain.ChainType,
		blockchain.ChainID,
		blockchain.WalletID,
		time.Now(),
	})

	if err != nil {
		return fmt.Errorf("failed to create blockchain: %w", err)
	}

	return nil
}

func (r *repository) FindByChainType(ctx context.Context, chainType types.ChainType) (*Blockchain, error) {
	query := `
		SELECT chain_type, chain_id, wallet_id, deactivated_at, created_at
		FROM blockchains
		WHERE chain_type = ?
	`

	rows, err := r.db.ExecuteQuery(query, []interface{}{chainType})
	if err != nil {
		return nil, fmt.Errorf("failed to find blockchain: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("blockchain not found")
	}

	var blockchain Blockchain
	err = rows.Scan(
		&blockchain.ChainType,
		&blockchain.ChainID,
		&blockchain.WalletID,
		&blockchain.DeactivatedAt,
		&blockchain.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to scan blockchain: %w", err)
	}

	return &blockchain, nil
}

func (r *repository) FindActive(ctx context.Context) ([]*Blockchain, error) {
	query := `
		SELECT chain_type, chain_id, wallet_id, deactivated_at, created_at
		FROM blockchains
		WHERE deactivated_at IS NULL
	`

	rows, err := r.db.ExecuteQuery(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to find active blockchains: %w", err)
	}
	defer rows.Close()

	var blockchains []*Blockchain
	for rows.Next() {
		var blockchain Blockchain
		err := rows.Scan(
			&blockchain.ChainType,
			&blockchain.ChainID,
			&blockchain.WalletID,
			&blockchain.DeactivatedAt,
			&blockchain.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan blockchain row: %w", err)
		}
		blockchains = append(blockchains, &blockchain)
	}

	return blockchains, nil
}

func (r *repository) Deactivate(ctx context.Context, chainType types.ChainType) error {
	query := `
		UPDATE blockchains
		SET deactivated_at = ?
		WHERE chain_type = ?
	`

	_, err := r.db.ExecuteStatement(query, []interface{}{
		time.Now(),
		chainType,
	})

	if err != nil {
		return fmt.Errorf("failed to deactivate blockchain: %w", err)
	}

	return nil
}
