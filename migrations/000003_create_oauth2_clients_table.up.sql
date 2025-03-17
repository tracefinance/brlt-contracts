-- Create OAuth2 clients table
CREATE TABLE IF NOT EXISTS oauth_clients (
    id BIGINT PRIMARY KEY NOT NULL,
    client_id TEXT UNIQUE NOT NULL,
    client_secret TEXT NOT NULL,
    redirect_uri TEXT
);

-- Add index on client_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_oauth_clients_client_id ON oauth_clients(client_id);