# US-006: redeem household invitation and create household

- **Epic:** [Household Onboarding and Family Graph](README.md)
- **Source use cases:** [UC-1](../../use-cases.md#use-case-1-new-family-joins)
- **Primary actor:** Primary Family Manager

**As a** primary Family Manager, **I want** to redeem my invitation, register a passkey, and create the household through my own identity, **so that** I can become its first manager and continue agreement-first onboarding.

## Scope

Redeem a valid new-household invitation, claim an email account identifier, register a passkey, create the household account, link the recipient's authenticated identity and profile, and grant that person first Family Manager authority.

## Preconditions

- An Admin has issued an unused, unexpired new-household invitation.
- The recipient uses a supported browser with JavaScript available for WebAuthn.

## Acceptance criteria

1. **Given** a valid new-household invitation, **when** the recipient opens it, claims an email, registers a passkey, and creates the household, **then** the recipient becomes its first Family Manager through their own identity.
2. **Given** the invitation is expired, already used, or for another purpose, **when** redemption is attempted, **then** no household or manager authority is created.
3. **Given** the claimed email is already linked to another active authenticated identity, **when** redemption is attempted, **then** the system rejects creating a second identity without revealing the other identity or household.
4. **Given** household creation succeeds, **when** the manager enters the application, **then** they can complete their profile and continue restricted agreement-first onboarding.
5. **Given** the household exists but a participant is not Confirmed for the current agreement, **when** the manager continues onboarding, **then** profile management and agreement confirmation remain available while that participant's scheduling remains locked.

## Business rules

- Every authenticated person uses their own claimed, system-wide unique normalized email and one or more passkeys.
- The claimed email is not mailbox-verified during enrollment.
- The household has at least one Family Manager.
- An authenticated identity is distinct from a person profile and a household account.
- The invitation is single-use, expiring, and purpose-bound.
- Invitation and authentication occur before agreement confirmation.
- Scheduling unlocks separately for each person only after that person's current-season agreement is Confirmed.

## Dependencies

- US-005 provides the authorized new-household invitation.
- US-002 provides secure passkey sign-in and personal authenticated sessions.

## Out of scope

- Open self-registration.
- Shared household credentials.
- Adding additional people or adult-to-scout relationships.
- Inviting a co-manager.
- Granting Young Adult Scout access.
- Performing agreement confirmation itself.
- Mailbox verification of the claimed email.
