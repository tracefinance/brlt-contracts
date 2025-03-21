CREATE TABLE IF NOT EXISTS wallets (
    id BIGINT PRIMARY KEY NOT NULL,
    key_id BIGINT,
    chain_type TEXT NOT NULL,
    address TEXT NOT NULL,
    name TEXT NOT NULL,
    tags TEXT,
    balance DECIMAL(36, 0) NOT NULL DEFAULT 0,
    last_block_number INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(chain_type, address)
);

CREATE INDEX IF NOT EXISTS idx_wallets_key_id ON wallets(key_id);
CREATE INDEX IF NOT EXISTS idx_wallets_chain_type ON wallets(chain_type);
CREATE INDEX IF NOT EXISTS idx_wallets_deleted_at ON wallets(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_wallets_chain_address ON wallets(chain_type, address);
