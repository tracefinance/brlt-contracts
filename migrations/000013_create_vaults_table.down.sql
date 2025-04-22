-- Revert migration for creating the vaults table
DROP INDEX IF EXISTS idx_vaults_address;
DROP INDEX IF EXISTS idx_vaults_status;
DROP INDEX IF EXISTS idx_vaults_wallet_id;
DROP TABLE IF EXISTS vaults;
