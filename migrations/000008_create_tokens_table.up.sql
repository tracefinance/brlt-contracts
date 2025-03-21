-- Create tokens table
CREATE TABLE IF NOT EXISTS tokens (
    address TEXT NOT NULL,
    chain_type TEXT NOT NULL,
    symbol TEXT NOT NULL,
    decimals INTEGER NOT NULL,
    type TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (address)
);

-- Add indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_tokens_symbol ON tokens(symbol);
CREATE INDEX IF NOT EXISTS idx_tokens_chain_type ON tokens(chain_type);
CREATE INDEX IF NOT EXISTS idx_tokens_type ON tokens(type);
