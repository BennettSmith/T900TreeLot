# US-005: invite a new household

- **Epic:** [Household Onboarding and Family Graph](README.md)
- **Source use cases:** [UC-1](../../use-cases.md#use-case-1-new-family-joins)
- **Primary actor:** Admin

**As an** Admin, **I want** to create a purpose-bound household invitation link or QR code for a new household's first manager, **so that** the household can be established without open self-registration.

## Scope

Create a unique, expiring new-household invitation that can be provided out of band as a link or QR code, for example at a troop meeting or on a printed enrollment form.

## Preconditions

- The first Admin has been established and bootstrap is closed.
- The Admin is authenticated and active.

## Acceptance criteria

1. **Given** an active Admin requests a new-household invitation, **when** the invitation is created, **then** the system creates a unique, single-use, expiring invitation bound to the new-household purpose and presents a link and QR code the Admin can provide out of band.
2. **Given** the invitation is created, **when** the Admin shares it, **then** delivery is out of band; the system does not send an authentication SMS.
3. **Given** a Committee Member, Family Manager, Young Adult Scout, or unauthenticated person attempts to create a new-household invitation, **when** the server evaluates authorization, **then** no invitation is created.
4. **Given** an invitation has expired, has already been used, or is presented for another purpose, **when** redemption is attempted, **then** it cannot establish a household.

## Business rules

- Only Admin may create a new-household invitation.
- There is no open self-registration.
- The invitation is unique, single-use, expires after a configured short period, and is bound to the new-household purpose.
- Invitation links and QR codes authorize enrollment; they are not SMS authentication messages.
- Server-side authorization applies regardless of full-page or HTMX navigation.

## Dependencies

- US-001 establishes the Admin who is authorized to create the invitation.

## Out of scope

- Registering the recipient's passkey.
- Creating the household or its first manager identity.
- Adding family-member profiles or relationships.
- Inviting a co-manager.
- Configuring or confirming a seasonal agreement.
