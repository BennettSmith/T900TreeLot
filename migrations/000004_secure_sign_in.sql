-- Stable WebAuthn user handles allow discoverable passkeys to resolve one identity.

ALTER TABLE identities
    ADD COLUMN webauthn_user_handle BYTEA NULL UNIQUE;
