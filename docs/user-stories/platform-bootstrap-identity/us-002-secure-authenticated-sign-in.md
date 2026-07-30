# US-002: secure authenticated sign-in

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-2](../../use-cases.md#use-case-2-authenticated-family-member-signs-in)
- **Primary actor:** Authenticated person

**As a** Family Manager or Young Adult Scout, **I want** to sign in with my verified mobile number and a one-time SMS credential, **so that** I can securely access only the functions permitted to my identity and roles.

## Scope

Provide passwordless browser sign-in for an active authenticated identity, create a secure revocable session, resolve that identity's permissions, and direct the person to the appropriate authorized starting view.

## Preconditions

- The person has an active authenticated identity with a system-wide unique verified mobile number.
- The identity was established through bootstrap, an authorized invitation, or role assignment.
- SMS authentication is available.

## Acceptance criteria

1. **Given** an active Family Manager or Young Adult Scout enters their registered number, **when** they successfully use the short-lived, single-use SMS code or magic link, **then** the system creates a secure session for that authenticated identity.
2. **Given** a Family Manager completes verification, **when** the session resolves their role, **then** the web app opens the family dashboard with only their permitted household functions.
3. **Given** a Young Adult Scout completes verification, **when** the session resolves the linked scout profile, **then** the web app opens that scout's personal schedule without authority over another person.
4. **Given** any phone number is submitted for sign-in, **when** the system sends its response, **then** the response is rate-limited and does not disclose whether the number is registered.
5. **Given** a code or link is expired or has already been used, **when** verification is attempted, **then** no authenticated session is created.
6. **Given** the signed-in person's current-season agreement is not Confirmed, **when** they enter the app, **then** the app prominently directs them to the Agreement Center and their own participation actions remain disabled.
7. **Given** a session has been revoked by the user, an authorized Family Manager for a managed Young Adult Scout profile, or an Admin, **when** that session is presented again, **then** it no longer grants access.

## Business rules

- Authentication belongs to a person, never to a shared household credential.
- Each authenticated person uses their own verified mobile number.
- A normalized phone number is linked to at most one active authenticated identity system-wide.
- One identity carries all roles held by the same person.
- Sign-in uses SMS one-time codes or magic links; social identity providers are not used.
- Browser sessions use secure, HTTP-only, same-site cookies and can be revoked.
- Invitation and authentication precede seasonal-agreement confirmation.

## Dependencies

- US-001 establishes the first Admin and closes bootstrap.

## Out of scope

- Open self-registration.
- Shared household login credentials.
- Google, Apple, or other social sign-in.
- Changing or recovering a phone number.
- Granting roles or Young Adult Scout access.
- Recording seasonal-agreement confirmation.
