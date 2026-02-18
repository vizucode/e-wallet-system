CREATE TABLE IF NOT EXISTS wallets (
    id          UUID PRIMARY KEY,
    owner_id    UUID NOT NULL,
    currency    CHAR(3) NOT NULL,
    balance     DECIMAL(20,2) NOT NULL DEFAULT 0.00,
    status      VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    version     INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_wallet_owner
        FOREIGN KEY (owner_id) REFERENCES users(id),

    CONSTRAINT uq_wallet_owner_currency
        UNIQUE (owner_id, currency),

    CONSTRAINT chk_wallet_balance
        CHECK (balance >= 0)
);
