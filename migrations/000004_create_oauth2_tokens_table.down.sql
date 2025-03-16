-- Drop indexes first
DROP INDEX IF EXISTS idx_oauth_tokens_client_id;
DROP INDEX IF EXISTS idx_oauth_tokens_user_id;
DROP INDEX IF EXISTS idx_oauth_tokens_code;
DROP INDEX IF EXISTS idx_oauth_tokens_access;
DROP INDEX IF EXISTS idx_oauth_tokens_refresh;

-- Drop the table
DROP TABLE IF EXISTS oauth_tokens;