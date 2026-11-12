-- +goose Up
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE auth.users (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name          text NOT NULL,
    email         text NOT NULL UNIQUE,
    password_hash text NOT NULL,
    created_at    timestamptz NOT NULL DEFAULT now(),
    updated_at    timestamptz NOT NULL DEFAULT now()
);

-- Only the SHA-256 of the session token is stored, so a database leak does
-- not leak usable session cookies.
CREATE TABLE auth.sessions (
    token_hash bytea PRIMARY KEY,
    user_id    uuid NOT NULL REFERENCES auth.users (id) ON DELETE CASCADE,
    expires_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX sessions_user_id_idx ON auth.sessions (user_id);

-- +goose Down
DROP SCHEMA auth CASCADE;
