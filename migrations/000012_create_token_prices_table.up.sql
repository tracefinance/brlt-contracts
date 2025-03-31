-- +migrate Up
CREATE TABLE IF NOT EXISTS token_prices (
    symbol TEXT PRIMARY KEY NOT NULL,
    rank INTEGER NOT NULL,
    price_usd DECIMAL(19,6) NOT NULL,
    supply DECIMAL(19,6) NOT NULL,
    market_cap_usd DECIMAL(19,6) NOT NULL,
    volume_usd_24h DECIMAL(19,6) NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_token_prices_rank ON token_prices(rank);
CREATE INDEX IF NOT EXISTS idx_token_prices_updated_at ON token_prices(updated_at DESC); 