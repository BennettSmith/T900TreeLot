# US-008: add a co-manager

- **Epic:** [Household Onboarding and Family Graph](README.md)
- **Source use cases:** [UC-7](../../use-cases.md#use-case-7-family-manager-adds-co-manager)
- **Primary actor:** Existing Family Manager

**As an** existing Family Manager, **I want** to invite a co-parent or guardian to manage our household through their own passkey login, **so that** either of us can independently manage household members and scheduling.

## Scope

Add the co-parent or guardian as an adult household member, issue a unique expiring co-manager invitation as a link or QR code, accept enrollment with a passkey, and grant management authority for the existing household.

## Preconditions

- The household is active and already has a Family Manager.
- The actor is an active Family Manager for that household, or an Admin acting with the documented authority.

## Acceptance criteria

1. **Given** an authorized existing Family Manager or Admin identifies an adult co-parent or guardian, **when** they invite that person, **then** the system creates a unique, single-use, expiring co-manager invitation bound to the existing household and purpose, optionally records an intended claimed email, and presents a link or QR code for out-of-band delivery.
2. **Given** the intended co-manager opens a valid invitation, **when** they claim an email if needed, register a passkey, and accept household access, **then** they receive Family Manager authority for that household through their own authenticated identity.
3. **Given** the invitee already has an authenticated identity, **when** they prove it with a passkey and accept, **then** household manager authority is linked to that existing identity rather than creating a second person.
4. **Given** the invitation is expired, already used, or presented for another purpose, **when** acceptance is attempted, **then** management authority is not granted.
5. **Given** the claimed email belongs to another active authenticated identity in a conflicting way, **when** acceptance would create or select a different person, **then** the system rejects the conflict without revealing the associated identity or household.
6. **Given** the co-manager has accepted, **when** either manager signs in, **then** each can independently manage family members and select eligible household adults or scouts for shifts using their own attributed session.

## Business rules

- Existing Family Managers and Admin may invite a household co-manager.
- A household has one or more Family Managers.
- Co-managers never share passkeys or browser sessions.
- The invitation is single-use, expiring, and purpose-bound.
- A normalized email is linked to only one active authenticated identity.
- Actions are attributed to the authenticated person who performed them.

## Dependencies

- US-006 establishes the household and its first manager.
- US-007 establishes the adult family-member profile used for the co-manager.
- US-002 provides secure passkey sign-in and a personal authenticated session.

## Out of scope

- Inviting the first manager of a new household.
- Shared household credentials.
- Granting Admin or Committee roles.
- Granting Young Adult Scout access.
- Removing a Family Manager or handling last-manager continuity.
- Scheduling individual shifts.
