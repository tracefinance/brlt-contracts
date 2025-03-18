-- Drop indexes for signer_addresses table
DROP INDEX IF EXISTS idx_signer_addresses_signer_id;
DROP INDEX IF EXISTS idx_signer_addresses_chain;
DROP INDEX IF EXISTS idx_signer_addresses_address;

-- Drop signer_addresses table
DROP TABLE IF EXISTS signer_addresses;

-- Drop indexes for signers table
DROP INDEX IF EXISTS idx_signers_user_id;
DROP INDEX IF EXISTS idx_signers_type;

-- Drop signers table
DROP TABLE IF EXISTS signers; 