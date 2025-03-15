-- Drop indexes first
DROP INDEX IF EXISTS idx_tokens_symbol;
DROP INDEX IF EXISTS idx_tokens_chain_type;
DROP INDEX IF EXISTS idx_tokens_type;

-- Drop tokens table
DROP TABLE IF EXISTS tokens;
