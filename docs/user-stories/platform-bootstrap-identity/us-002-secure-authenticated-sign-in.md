# US-002: secure authenticated sign-in

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-2](../../use-cases.md#use-case-2-authenticated-family-member-signs-in)
- **Primary actor:** Authenticated person

**As a** Family Manager or Young Adult Scout, **I want** to sign in with a passkey registered to my identity, **so that** I can securely access only the functions permitted to my identity and roles.

## Scope

Provide passwordless browser sign-in for an active authenticated identity using WebAuthn/passkeys, create a secure revocable session, resolve that identity's permissions, and direct the person to the appropriate authorized starting view.

## Preconditions

- The person has an active authenticated identity with at least one registered passkey and a system-wide unique claimed email.
- The identity was established through bootstrap, an authorized invitation, or role assignment.
- The browser supports WebAuthn and JavaScript is available.

## Acceptance criteria

1. **Given** an active Family Manager or Young Adult Scout with a registered passkey, **when** they successfully complete a WebAuthn assertion, **then** the system creates a secure session for that authenticated identity.
2. **Given** a Family Manager completes sign-in, **when** the session resolves their role, **then** the web app opens the family dashboard with only their permitted household functions.
3. **Given** a Young Adult Scout completes sign-in, **when** the session resolves the linked scout profile, **then** the web app opens that scout's personal schedule without authority over another person.
4. **Given** discoverable credentials are available, **when** the person chooses sign-in with passkey, **then** they are not required to type an email first.
5. **Given** an account hint is needed, **when** an email is submitted before assertion, **then** the response is rate-limited and does not disclose whether the email is registered.
6. **Given** a session has been revoked by the user, an authorized Family Manager for a managed Young Adult Scout profile, or an Admin, **when** that session is presented again, **then** it no longer grants access.

## Business rules

- Authentication belongs to a person, never to a shared household credential.
- Each authenticated person uses their own claimed email and one or more passkeys.
- A normalized email is linked to at most one active authenticated identity system-wide.
- One identity carries all roles held by the same person.
- Sign-in uses passkeys; SMS one-time codes, magic links, and social identity providers are not used.
- Unverified email is not a sign-in factor.
- Browser sessions use secure, HTTP-only, same-site cookies and can be revoked.
- Invitation and authentication precede seasonal-agreement confirmation.

## Dependencies

- US-001 establishes the first Admin and closes bootstrap.

## Out of scope

- Open self-registration.
- Shared household login credentials.
- Google, Apple, or other social sign-in.
- Changing email or managing passkeys.
- Assisted recovery when no passkey remains.
- Granting roles or Young Adult Scout access.
- Recording seasonal-agreement confirmation.
- Agreement Center direction and participation eligibility enforcement, which are delivered by the seasonal-agreement stories in INC-04.
