-- Recent passkey step-up freshness for security-sensitive account actions.

ALTER TABLE sessions
    ADD COLUMN step_up_at TIMESTAMPTZ NULL;
