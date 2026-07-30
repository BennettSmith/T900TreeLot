# US-005: invite a new household

- **Epic:** [Household Onboarding and Family Graph](README.md)
- **Source use cases:** [UC-1](../../use-cases.md#use-case-1-new-family-joins)
- **Primary actor:** Admin

**As an** Admin, **I want** to send a purpose-bound invitation to a new household's first manager, **so that** the household can be established without open self-registration.

## Scope

Create a unique, expiring new-household invitation bound to the intended first Family Manager's normalized phone number and send it to that number.

## Preconditions

- The first Admin has been established and bootstrap is closed.
- The Admin is authenticated and active.
- The intended primary manager's mobile phone number is available.

## Acceptance criteria

1. **Given** an active Admin provides the intended primary manager's phone number, **when** the Admin creates a new-household invitation, **then** the system creates a unique, single-use, expiring invitation bound to that purpose and normalized phone number.
2. **Given** the invitation is created, **when** delivery is requested, **then** it is sent to the intended primary manager's mobile number.
3. **Given** a Committee Member, Family Manager, Young Adult Scout, or unauthenticated person attempts to create a new-household invitation, **when** the server evaluates authorization, **then** no invitation is created.
4. **Given** an invitation has expired, has already been used, is presented for another purpose, or is redeemed with a different normalized number, **when** redemption is attempted, **then** it cannot establish a household.

## Business rules

- Only Admin may create a new-household invitation.
- There is no open self-registration.
- The invitation is unique, single-use, expires after a configured short period, and is bound to the new-household purpose.
- The invitation is bound to the intended normalized phone number.
- Invitation and authentication messages are transactional and are not controlled by notification preferences.
- Server-side authorization applies regardless of full-page or HTMX navigation.

## Dependencies

- US-001 establishes the Admin who is authorized to create the invitation.

## Out of scope

- Verifying the recipient's phone number.
- Creating the household or its first manager identity.
- Adding family-member profiles or relationships.
- Inviting a co-manager.
- Configuring or confirming a seasonal agreement.
