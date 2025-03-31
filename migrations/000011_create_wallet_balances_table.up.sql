CREATE TABLE IF NOT EXISTS wallet_balances (
    wallet_id BIGINT NOT NULL,
    token_address TEXT NOT NULL,
    balance DECIMAL(36, 0) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (wallet_id, token_address),
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    FOREIGN KEY (token_address) REFERENCES tokens(address)
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_wallet_balances_wallet_id ON wallet_balances(wallet_id);
CREATE INDEX IF NOT EXISTS idx_wallet_balances_token_address ON wallet_balances(token_address);
