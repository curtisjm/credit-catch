-- 001_initial_schema.up.sql
-- Full schema for credit_catch backend.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       TEXT NOT NULL UNIQUE,
    password    TEXT NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Card catalog: all known credit cards
CREATE TABLE card_catalog (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    issuer      TEXT NOT NULL,
    name        TEXT NOT NULL,
    network     TEXT NOT NULL, -- visa, mastercard, amex, discover
    annual_fee  INTEGER NOT NULL DEFAULT 0, -- cents
    image_url   TEXT NOT NULL DEFAULT '',
    active      BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (issuer, name)
);

-- Credit definitions: benefits a card offers (e.g. "$15/mo Uber credit")
CREATE TABLE credit_definitions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    card_catalog_id UUID NOT NULL REFERENCES card_catalog(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    amount_cents    INTEGER NOT NULL,
    period          TEXT NOT NULL CHECK (period IN ('monthly', 'annual', 'one_time')),
    category        TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Rules that match transactions to credit definitions
CREATE TABLE credit_match_rules (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    credit_definition_id  UUID NOT NULL REFERENCES credit_definitions(id) ON DELETE CASCADE,
    field                 TEXT NOT NULL, -- merchant_name, category, etc.
    operator              TEXT NOT NULL CHECK (operator IN ('equals', 'contains', 'starts_with', 'regex')),
    value                 TEXT NOT NULL,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Cards a user holds
CREATE TABLE user_cards (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_catalog_id UUID NOT NULL REFERENCES card_catalog(id) ON DELETE RESTRICT,
    nickname        TEXT NOT NULL DEFAULT '',
    opened_date     DATE,
    annual_fee_date DATE, -- when the annual fee renews
    active          BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, card_catalog_id)
);

-- Tracks credit usage per period
CREATE TABLE credit_periods (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_card_id          UUID NOT NULL REFERENCES user_cards(id) ON DELETE CASCADE,
    credit_definition_id  UUID NOT NULL REFERENCES credit_definitions(id) ON DELETE CASCADE,
    period_start          DATE NOT NULL,
    period_end            DATE NOT NULL,
    used                  BOOLEAN NOT NULL DEFAULT false,
    used_at               TIMESTAMPTZ,
    amount_used_cents     INTEGER NOT NULL DEFAULT 0,
    transaction_id        UUID, -- nullable, linked when matched
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_card_id, credit_definition_id, period_start)
);

-- Notification preferences
CREATE TABLE notification_prefs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    channel         TEXT NOT NULL CHECK (channel IN ('email', 'push', 'sms')),
    enabled         BOOLEAN NOT NULL DEFAULT true,
    remind_days     INTEGER NOT NULL DEFAULT 3, -- days before period end
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, channel)
);

-- Statement uploads for manual credit verification
CREATE TABLE statement_uploads (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_card_id UUID NOT NULL REFERENCES user_cards(id) ON DELETE CASCADE,
    filename    TEXT NOT NULL,
    s3_key      TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Plaid items: linked bank/card connections
CREATE TABLE plaid_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plaid_item_id   TEXT NOT NULL UNIQUE,
    access_token    TEXT NOT NULL,
    institution_id  TEXT NOT NULL DEFAULT '',
    institution_name TEXT NOT NULL DEFAULT '',
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'needs_reauth', 'revoked')),
    last_synced_at  TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Transactions from Plaid or statement parsing
CREATE TABLE transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plaid_item_id   UUID REFERENCES plaid_items(id) ON DELETE SET NULL,
    user_card_id    UUID REFERENCES user_cards(id) ON DELETE SET NULL,
    plaid_txn_id    TEXT UNIQUE,
    merchant_name   TEXT NOT NULL DEFAULT '',
    amount_cents    INTEGER NOT NULL,
    category        TEXT NOT NULL DEFAULT '',
    date            DATE NOT NULL,
    pending         BOOLEAN NOT NULL DEFAULT false,
    source          TEXT NOT NULL DEFAULT 'plaid' CHECK (source IN ('plaid', 'statement', 'manual')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes for common queries
CREATE INDEX idx_user_cards_user_id ON user_cards(user_id);
CREATE INDEX idx_credit_periods_user_card_id ON credit_periods(user_card_id);
CREATE INDEX idx_credit_periods_period ON credit_periods(period_start, period_end);
CREATE INDEX idx_transactions_user_id ON transactions(user_id);
CREATE INDEX idx_transactions_date ON transactions(date);
CREATE INDEX idx_transactions_user_card_id ON transactions(user_card_id);
CREATE INDEX idx_credit_definitions_card ON credit_definitions(card_catalog_id);
CREATE INDEX idx_plaid_items_user_id ON plaid_items(user_id);
