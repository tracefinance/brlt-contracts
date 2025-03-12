-- Create the blockchains table
CREATE TABLE IF NOT EXISTS blockchains (
    chain_type TEXT PRIMARY KEY,
    chain_id INTEGER NOT NULL,
    wallet_id TEXT NOT NULL,
    deactivated_at DATETIME,
    created_at DATETIME NOT NULL
);

ALTER TABLE blockchains ADD CONSTRAINT fk_blockchains_wallet FOREIGN KEY (wallet_id) 
REFERENCES wallets(id) ON DELETE RESTRICT ON UPDATE CASCADE;