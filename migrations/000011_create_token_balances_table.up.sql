CREATE TABLE IF NOT EXISTS token_balances (
    wallet_id BIGINT NOT NULL,
    token_id BIGINT NOT NULL,
    balance DECIMAL(36, 0) NOT NULL DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (wallet_id, token_id),
    FOREIGN KEY (wallet_id) REFERENCES wallets(id),
    FOREIGN KEY (token_id) REFERENCES tokens(id)
);

-- Create index for faster queries
CREATE INDEX IF NOT EXISTS idx_token_balances_wallet_id ON token_balances(wallet_id);
CREATE INDEX IF NOT EXISTS idx_token_balances_token_id ON token_balances(token_id); 