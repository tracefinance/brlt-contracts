-- Migration for creating the vaults table

CREATE TABLE vaults (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    wallet_id BIGINT NOT NULL,
    chain_type TEXT NOT NULL,
    tx_hash TEXT NOT NULL,
    recovery_address TEXT NOT NULL,        
    signers TEXT NOT NULL, -- JSON array of signer addresses    
    status TEXT NOT NULL DEFAULT 'pending',
    signature_threshold INT NOT NULL,
    address TEXT,
    recovery_request_timestamp TIMESTAMP,
    failure_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_vaults_wallet_id ON vaults (wallet_id);
CREATE INDEX idx_vaults_status ON vaults (status);
CREATE INDEX idx_vaults_address ON vaults (address);
