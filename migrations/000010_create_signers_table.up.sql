-- Create signers table
CREATE TABLE IF NOT EXISTS signers (
    id BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('internal', 'external')),
    user_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX idx_signers_user_id ON signers(user_id);
CREATE INDEX idx_signers_type ON signers(type);

-- Create signer addresses table
CREATE TABLE IF NOT EXISTS signer_addresses (
    id BIGINT PRIMARY KEY,
    signer_id BIGINT NOT NULL,
    chain_type TEXT NOT NULL,
    address TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (signer_id) REFERENCES signers(id) ON DELETE CASCADE,
    UNIQUE (address, chain_type)
);

-- Create indexes
CREATE INDEX idx_signer_addresses_signer_id ON signer_addresses(signer_id);
CREATE INDEX idx_signer_addresses_chain ON signer_addresses(chain_type);
CREATE INDEX idx_signer_addresses_address ON signer_addresses(address); 