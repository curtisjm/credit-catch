-- 002_oauth_and_statement_day.up.sql
-- Add OAuth support and statement close day for credit period anchoring.

-- Allow OAuth-only accounts (no password).
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- OAuth provider links.
CREATE TABLE oauth_accounts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL, -- google, apple
    provider_uid    TEXT NOT NULL, -- provider's user ID
    email           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, provider_uid)
);

CREATE INDEX idx_oauth_accounts_user_id ON oauth_accounts(user_id);

-- Statement close day drives monthly credit period boundaries.
ALTER TABLE user_cards ADD COLUMN statement_close_day INTEGER
    CHECK (statement_close_day >= 1 AND statement_close_day <= 31);
