CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT PRIMARY KEY NOT NULL,
    wallet_id BIGINT NOT NULL,
    chain_type TEXT NOT NULL,
    hash TEXT NOT NULL,
    from_address TEXT NOT NULL,
    to_address TEXT NOT NULL,
    value DECIMAL(36, 0) NOT NULL,
    data BLOB,
    nonce BIGINT UNSIGNED,
    gas_price DECIMAL(36, 0),
    gas_limit BIGINT UNSIGNED,
    type TEXT NOT NULL,
    token_address TEXT,
    token_symbol TEXT,
    status TEXT NOT NULL,
    timestamp INTEGER NOT NULL,
    block_number BIGINT UNSIGNED DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(wallet_id) REFERENCES wallets(id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_chain_type ON transactions(chain_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_hash ON transactions(hash);
CREATE INDEX IF NOT EXISTS idx_transactions_timestamp ON transactions(timestamp);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_block_number ON transactions(block_number); 