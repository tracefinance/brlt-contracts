-- Create keys table for key management
CREATE TABLE IF NOT EXISTS keys (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    key_type TEXT NOT NULL,
    curve TEXT, -- Curve type for ECDSA keys (e.g., P256, P384, P521, secp256k1)
    tags TEXT, -- JSON encoded map of tags
    created_at INTEGER NOT NULL,
    private_key BLOB, -- Encrypted private key material
    public_key BLOB -- Public key material (if applicable)
);

-- Create indexes for common lookups
CREATE INDEX IF NOT EXISTS idx_keys_name ON keys(name);
CREATE INDEX IF NOT EXISTS idx_keys_type ON keys(key_type);
CREATE INDEX IF NOT EXISTS idx_keys_created_at ON keys(created_at);
CREATE INDEX IF NOT EXISTS idx_keys_curve ON keys(curve); -- Index for curve lookups 