-- Drop foreign key constraint
ALTER TABLE blockchains 
DROP CONSTRAINT fk_blockchains_wallet;

-- Drop the table
DROP TABLE IF EXISTS blockchains; 