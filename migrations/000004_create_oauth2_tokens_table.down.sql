-- Drop tokens table and related indexes
DROP INDEX IF EXISTS idx_tokens_client_id;
DROP INDEX IF EXISTS idx_tokens_user_id;
DROP INDEX IF EXISTS idx_tokens_code;
DROP INDEX IF EXISTS idx_tokens_access;
DROP INDEX IF EXISTS idx_tokens_refresh;
DROP TABLE IF EXISTS tokens;