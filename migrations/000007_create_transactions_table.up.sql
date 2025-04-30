CREATE TABLE transactions (
    -- Service Layer Fields
    id BIGINT PRIMARY KEY NOT NULL,
    wallet_id BIGINT DEFAULT NULL,
    vault_id BIGINT DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP DEFAULT NULL,

    -- Mirrored BaseTransaction Fields
    chain_type TEXT NOT NULL,
    hash TEXT NOT NULL UNIQUE,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    value DECIMAL(36, 0) NOT NULL,
    data BLOB,
    nonce BIGINT UNSIGNED NOT NULL,
    gas_price DECIMAL(36, 0) NOT NULL,
    gas_limit BIGINT UNSIGNED NOT NULL,
    type TEXT NOT NULL,

    -- Mirrored Transaction Execution Fields
    gas_used BIGINT UNSIGNED DEFAULT NULL,
    status TEXT NOT NULL,
    timestamp BIGINT DEFAULT NULL,
    block_number DECIMAL(36, 0) DEFAULT NULL,

    FOREIGN KEY(wallet_id) REFERENCES wallets(id),
    FOREIGN KEY(vault_id) REFERENCES vaults(id)
);

-- Recreate necessary indexes
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_chain_type ON transactions(chain_type);
-- UNIQUE index on hash is created via the table definition
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_block_number ON transactions(block_number);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_from_address ON transactions(from_address);
CREATE INDEX IF NOT EXISTS idx_transactions_to_address ON transactions(to_address);
CREATE INDEX IF NOT EXISTS idx_transactions_deleted_at ON transactions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_transactions_vault_id ON transactions(vault_id);