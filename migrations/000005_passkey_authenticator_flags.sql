-- Preserve WebAuthn credential flags needed to validate backup-eligible passkeys.

ALTER TABLE passkey_credentials
    ADD COLUMN authenticator_flags SMALLINT NULL
    CHECK (authenticator_flags BETWEEN 0 AND 255);
