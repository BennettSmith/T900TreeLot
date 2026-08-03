-- Increment 2 identity bootstrap foundations: personal profiles, passkeys,
-- bootstrap closure, authentication sessions, and rate-limit buckets.

CREATE TABLE people (
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    preferred_display_name TEXT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE identities (
    id TEXT PRIMARY KEY,
    person_id TEXT NOT NULL UNIQUE REFERENCES people (id),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE identity_emails (
    identity_id TEXT NOT NULL REFERENCES identities (id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    email_normalized TEXT NOT NULL,
    verified_at TIMESTAMPTZ NULL,
    active BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (identity_id, email_normalized)
);

CREATE UNIQUE INDEX identity_emails_active_normalized_idx
    ON identity_emails (email_normalized)
    WHERE active;

CREATE TABLE identity_roles (
    identity_id TEXT NOT NULL REFERENCES identities (id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('admin', 'committee', 'family_manager', 'young_adult_scout')),
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE (identity_id, role)
);

CREATE TABLE passkey_credentials (
    id TEXT PRIMARY KEY,
    identity_id TEXT NOT NULL REFERENCES identities (id) ON DELETE CASCADE,
    credential_id BYTEA NOT NULL UNIQUE,
    public_key BYTEA NOT NULL,
    attestation_type TEXT NOT NULL,
    aaguid TEXT NOT NULL,
    sign_count BIGINT NOT NULL,
    transports TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ NULL
);

ALTER TABLE sessions
    ADD COLUMN identity_id TEXT NULL REFERENCES identities (id),
    ADD COLUMN authenticated_at TIMESTAMPTZ NULL;

CREATE TABLE webauthn_ceremonies (
    id TEXT PRIMARY KEY,
    session_id BIGINT NULL REFERENCES sessions (id) ON DELETE CASCADE,
    purpose TEXT NOT NULL,
    challenge BYTEA NOT NULL,
    identity_id TEXT NULL REFERENCES identities (id) ON DELETE CASCADE,
    user_handle BYTEA NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX webauthn_ceremonies_session_idx ON webauthn_ceremonies (session_id);
CREATE INDEX webauthn_ceremonies_expiry_idx ON webauthn_ceremonies (expires_at);

CREATE TABLE bootstrap_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    closed_at TIMESTAMPTZ NULL,
    closed_by_identity_id TEXT NULL REFERENCES identities (id)
);

INSERT INTO bootstrap_state (id, closed_at, closed_by_identity_id)
VALUES (1, NULL, NULL);

CREATE TABLE rate_limit_buckets (
    bucket_key TEXT PRIMARY KEY,
    window_started_at TIMESTAMPTZ NOT NULL,
    count INTEGER NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
