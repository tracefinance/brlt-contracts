-- Create OAuth2 clients table
CREATE TABLE IF NOT EXISTS clients (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    client_id TEXT UNIQUE NOT NULL,
    client_secret TEXT NOT NULL,
    redirect_uri TEXT
);

-- Add index on client_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_clients_client_id ON clients(client_id);