-- Drop indexes
DROP INDEX IF EXISTS idx_keys_name;
DROP INDEX IF EXISTS idx_keys_type;
DROP INDEX IF EXISTS idx_keys_created_at;

-- Drop keys table
DROP TABLE IF EXISTS keys; 