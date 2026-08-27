# US-003: manage own passkeys and account email

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-2B](../../use-cases.md#use-case-2b-authenticated-person-manages-credentials-and-account-email)
- **Primary actor:** Authenticated person

**As an** authenticated person who still controls at least one passkey, **I want** to add or remove passkeys and change my claimed account email, **so that** I retain my identity and history while keeping sign-in credentials current.

## Scope

Support self-service passkey management and claimed-email change when a current passkey remains available, including recent passkey step-up, uniqueness validation, atomic email replacement on the existing identity, and session revocation after email change.

## Preconditions

- The actor is a Family Manager, Young Adult Scout, Committee Member, or Admin with an existing authenticated identity.
- The actor can complete a recent passkey assertion.
- For email change, the proposed new email is not linked to another active identity.

## Acceptance criteria

1. **Given** the actor can use an existing passkey, **when** they open account security settings, **then** the system requires a recent passkey step-up before credential changes.
2. **Given** step-up succeeded, **when** the actor registers an additional passkey on the current device, **then** that credential is bound to the same identity and can be used on later sign-ins.
3. **Given** the actor still has another passkey, **when** they remove one passkey they control, **then** that credential can no longer authenticate the identity.
4. **Given** recent passkey step-up succeeded and the proposed email is not linked to another active identity, **when** the actor changes their account email, **then** the system atomically replaces the claimed email on the existing identity and marks it unverified.
5. **Given** the email change succeeds, **when** the replacement is committed, **then** all existing browser sessions are revoked and the person must sign in again with a passkey.
6. **Given** the normalized new email is already linked to another person, **when** the actor attempts the change, **then** the request is rejected without revealing that person's identity and the old email remains unchanged.
7. **Given** credentials or email are updated, **when** the person signs in again, **then** the same internal identity, profile, roles, household memberships, assignments, and history remain intact.

## Business rules

- A normalized email is system-wide unique among active authenticated identities.
- Claiming or changing an email does not prove mailbox ownership.
- Unverified email cannot be used for notifications or email-based recovery.
- The operation updates an existing identity; it does not create a replacement person or migrate history.
- An identity must retain at least one passkey unless access is being removed.
- All prior sessions are revoked after a successful email change.
- Email-conflict messages do not reveal another identity or household.

## Dependencies

- US-002 provides authenticated access and revocable secure sessions.

## Out of scope

- Recovery when no passkey remains.
- Mailbox verification of the claimed email.
- Reassigning an email from another person's active identity.
- Changing profiles, roles, memberships, assignments, or historical records.
- Removing authenticated access.
