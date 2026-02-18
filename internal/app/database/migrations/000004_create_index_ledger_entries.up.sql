CREATE INDEX IF NOT EXISTS idx_ledger_wallet_id
    ON ledger_entries(wallet_id);