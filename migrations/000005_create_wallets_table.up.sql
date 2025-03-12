CREATE TABLE IF NOT EXISTS wallets (
    id TEXT PRIMARY KEY,
    key_id TEXT,
    chain_type TEXT NOT NULL,
    address TEXT NOT NULL,
    name TEXT NOT NULL,
    tags TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_wallets_key_id ON wallets(key_id);
CREATE INDEX IF NOT EXISTS idx_wallets_chain_type ON wallets(chain_type);
CREATE INDEX IF NOT EXISTS idx_wallets_deleted_at ON wallets(deleted_at);
