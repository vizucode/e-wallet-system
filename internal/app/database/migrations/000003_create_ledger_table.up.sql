CREATE TABLE IF NOT EXISTS ledger_entries (
    id                  UUID PRIMARY KEY,
    wallet_id           UUID NOT NULL,
    related_wallet_id   UUID,
    currency            CHAR(3) NOT NULL,
    amount              DECIMAL(20,2) NOT NULL,
    entry_type          VARCHAR(30) NOT NULL,
    reference_id        VARCHAR(100) NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_ledger_wallet
        FOREIGN KEY (wallet_id) REFERENCES wallets(id),

    CONSTRAINT fk_ledger_related_wallet
        FOREIGN KEY (related_wallet_id) REFERENCES wallets(id),

    CONSTRAINT uq_ledger_reference
        UNIQUE (reference_id)
);
