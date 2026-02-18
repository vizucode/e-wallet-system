CREATE INDEX IF NOT EXISTS idx_ledger_created_at
    ON ledger_entries(created_at);