# US-010: grant Young Adult Scout access to an existing scout profile

- **Epic:** [Young Adult Scout Access](README.md)
- **Source use cases:** [UC-2A](../../use-cases.md#use-case-2a-older-scout-becomes-a-young-adult-scout)
- **Primary actor:** Family Manager

**As a** Family Manager, **I want** to grant an older scout limited access through their own passkey login, **so that** they can manage their own participation without gaining authority over anyone else.

## Scope

Authorize access from an existing scout profile, issue a profile-specific invitation as a link or QR code, accept enrollment with a claimed email and passkey, and link one authenticated identity to that profile with Young Adult Scout permissions.

## Preconditions

- The scout already has one profile in at least one household.
- The actor is a Family Manager who manages that profile, or an Admin.

## Acceptance criteria

1. **Given** an authorized Family Manager or Admin starts from an existing scout profile, **when** access is granted, **then** the system creates a unique, single-use, expiring Young Adult Scout invitation bound to that profile and purpose and presents a link or QR code for out-of-band delivery.
2. **Given** the scout opens a valid invitation, **when** they claim an email, register a passkey, and accept, **then** the system links a new authenticated identity to the existing scout profile without creating another family member.
3. **Given** the claimed email is already linked to a Family Manager, Young Adult Scout, Committee Member, or Admin identity, **when** acceptance is attempted, **then** the invitation is rejected and cannot silently create or select a different profile.
4. **Given** the scout has accepted access, **when** the identity is resolved for a self-service request, **then** it resolves to the linked scout profile so downstream scheduling, attendance, and reporting stories can enforce self-only permissions.
5. **Given** the scout has Young Adult Scout access, **when** they attempt to view another person's private details or schedule, act for another person, edit family members, invite managers, or change family settings, **then** the server denies the action.
6. **Given** the scout belongs to multiple linked households, **when** access is accepted or the scout signs in, **then** the same identity resolves to the same scout profile, family memberships, and personal schedule across those households.
7. **Given** Young Adult Scout access is active, **when** a Family Manager views or manages the scout, **then** Family Manager visibility and management authority over the scout's assignments remain.
8. **Given** the scout has not confirmed the current-season agreement, **when** they attempt shift signup or attendance, **then** those participation actions remain disabled until the scout confirms.
9. **Given** the authenticated scout submits a self-service action, **when** the server selects its target, **then** it uses the identity's linked scout profile and does not trust a client-submitted person identifier.

## Business rules

- Young Adult Scout is an application permission level, not a statement of legal adulthood.
- Only a Family Manager for the scout's profile or an Admin may grant or revoke this access.
- Family membership exists on the profile before authenticated access.
- The invitation references one existing scout profile and its authorizing actor.
- A claimed normalized email belongs to only one active identity system-wide.
- The scout receives self-only permissions; Family Managers retain oversight.
- Young Adult Scout access does not change the person's adult/scout classification.
- Authentication occurs before agreement confirmation, but confirmation gates participation.

## Dependencies

- US-002 provides passkey sign-in and secure identity sessions.
- US-007 provides the existing scout profile and Family Manager authority.

## Out of scope

- Creating a new scout or household.
- Determining legal adulthood.
- Granting access when the claimed email is already linked to another identity.
- Giving the scout authority over another person or household settings.
- Implementing the downstream signup, cancellation, attendance, or statistics workflows.
- Confirming the seasonal agreement.
- Mailbox verification of the claimed email.
