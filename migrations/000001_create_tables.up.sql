-- Users table
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(255) NOT NULL UNIQUE,
  email VARCHAR(255) NOT NULL UNIQUE,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);

-- Wallets table
CREATE TABLE IF NOT EXISTS wallets (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  balance DECIMAL(19, 4) NOT NULL DEFAULT 0,
  currency VARCHAR(10) NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  UNIQUE(user_id)
);

-- Create index on user_id
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
  id SERIAL PRIMARY KEY,
  wallet_id INTEGER NOT NULL REFERENCES wallets(id) ON DELETE CASCADE,
  dest_wallet_id INTEGER REFERENCES wallets(id) ON DELETE SET NULL,
  type VARCHAR(20) NOT NULL,
  amount DECIMAL(19, 4) NOT NULL,
  balance_before DECIMAL(19, 4) NOT NULL,
  balance_after DECIMAL(19, 4) NOT NULL,
  description TEXT,
  transaction_time TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL
);

-- Create index on wallet_id
CREATE INDEX IF NOT EXISTS idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX IF NOT EXISTS idx_transactions_transaction_time ON transactions(transaction_time);