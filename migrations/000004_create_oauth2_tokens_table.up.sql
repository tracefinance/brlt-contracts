-- Create OAuth2 tokens table
CREATE TABLE IF NOT EXISTS tokens (
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
CREATE INDEX IF NOT EXISTS idx_tokens_client_id ON tokens(client_id);
CREATE INDEX IF NOT EXISTS idx_tokens_user_id ON tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_tokens_code ON tokens(code);
CREATE INDEX IF NOT EXISTS idx_tokens_access ON tokens(access);
CREATE INDEX IF NOT EXISTS idx_tokens_refresh ON tokens(refresh);