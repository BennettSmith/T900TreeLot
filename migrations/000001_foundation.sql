-- Increment 1 foundation primitives: audit, outbox, jobs, sessions.
-- TIMESTAMPTZ defaults use now() so the stored instant is independent of the
-- session TimeZone. Avoid casting now() through timestamp without time zone.

CREATE TABLE audit_events (
    id BIGSERIAL PRIMARY KEY,
    actor_id TEXT NULL,
    action TEXT NOT NULL,
    target_type TEXT NOT NULL,
    target_id TEXT NOT NULL,
    correlation_id TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX audit_events_created_at_idx ON audit_events (created_at);
CREATE INDEX audit_events_target_idx ON audit_events (target_type, target_id);

CREATE TABLE outbox_messages (
    id BIGSERIAL PRIMARY KEY,
    idempotency_key TEXT NOT NULL UNIQUE,
    channel TEXT NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT outbox_messages_status_check CHECK (status IN ('pending', 'processing', 'delivered', 'failed'))
);

CREATE INDEX outbox_messages_claim_idx ON outbox_messages (status, available_at);

CREATE TABLE background_jobs (
    id BIGSERIAL PRIMARY KEY,
    job_type TEXT NOT NULL,
    dedupe_key TEXT NULL,
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INTEGER NOT NULL DEFAULT 0,
    available_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    lease_owner TEXT NULL,
    lease_expires_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT background_jobs_status_check CHECK (status IN ('pending', 'leased', 'completed', 'failed'))
);

CREATE UNIQUE INDEX background_jobs_dedupe_pending_idx
    ON background_jobs (job_type, dedupe_key)
    WHERE dedupe_key IS NOT NULL AND status IN ('pending', 'leased');

CREATE INDEX background_jobs_claim_idx ON background_jobs (status, available_at);

CREATE TABLE sessions (
    id BIGSERIAL PRIMARY KEY,
    token_hash BYTEA NOT NULL UNIQUE,
    csrf_token TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX sessions_expires_at_idx ON sessions (expires_at);
