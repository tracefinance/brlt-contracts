-- +migrate Down
DROP INDEX IF EXISTS idx_token_prices_rank;
DROP INDEX IF EXISTS idx_token_prices_updated_at;
DROP TABLE IF EXISTS token_prices; 