-- Bind bootstrap identity and profile claims to the WebAuthn ceremony that
-- attests them. Columns remain nullable for non-bootstrap ceremony purposes
-- and for ceremonies that were already in flight during deployment.

ALTER TABLE webauthn_ceremonies
    ADD COLUMN bootstrap_email TEXT NULL,
    ADD COLUMN bootstrap_first_name TEXT NULL,
    ADD COLUMN bootstrap_last_name TEXT NULL,
    ADD COLUMN bootstrap_preferred_display_name TEXT NULL;
