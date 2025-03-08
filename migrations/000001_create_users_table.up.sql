-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT
);

-- Add index on username for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
