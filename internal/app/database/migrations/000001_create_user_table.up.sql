CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY,
    email       VARCHAR(100),
    name        VARCHAR(100),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
