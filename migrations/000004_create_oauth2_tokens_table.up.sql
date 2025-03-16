-- Create OAuth2 tokens table
CREATE TABLE IF NOT EXISTS oauth_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id TEXT,
    user_id TEXT,
    redirect_uri TEXT,
    scope TEXT,
    code TEXT UNIQUE,
    code_created_at INTEGER,
    code_expires_in INTEGER,
    access TEXT UNIQUE,
    access_created_at INTEGER,
    access_expires_in INTEGER,
    refresh TEXT UNIQUE,
    refresh_created_at INTEGER,
    refresh_expires_in INTEGER
);

-- Add indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_client_id ON oauth_tokens(client_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_user_id ON oauth_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_code ON oauth_tokens(code);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_access ON oauth_tokens(access);
CREATE INDEX IF NOT EXISTS idx_oauth_tokens_refresh ON oauth_tokens(refresh);