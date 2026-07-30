# US-006: redeem household invitation and create household

- **Epic:** [Household Onboarding and Family Graph](README.md)
- **Source use cases:** [UC-1](../../use-cases.md#use-case-1-new-family-joins)
- **Primary actor:** Primary Family Manager

**As a** primary Family Manager, **I want** to redeem my invitation and create the household through my own verified identity, **so that** I can become its first manager and continue agreement-first onboarding.

## Scope

Redeem a valid new-household invitation, verify the invitation-bound phone number, create the household account, link the recipient's authenticated identity and profile, and grant that person first Family Manager authority.

## Preconditions

- An Admin has issued an unused, unexpired new-household invitation to the intended normalized phone number.
- The recipient can receive an SMS one-time code or magic link at that number.

## Acceptance criteria

1. **Given** a valid new-household invitation, **when** the recipient opens it and successfully verifies the bound phone number, **then** the recipient may create the household and becomes its first Family Manager through their own identity.
2. **Given** the invitation is expired, already used, for another purpose, or presented by a different normalized phone number, **when** redemption is attempted, **then** no household or manager authority is created.
3. **Given** the verified number is already linked to another active authenticated identity, **when** redemption is attempted, **then** the system rejects creating a second identity without revealing the other identity or household.
4. **Given** household creation succeeds, **when** the manager enters the application, **then** they can complete their profile and continue restricted agreement-first onboarding.
5. **Given** the household exists but a participant is not Confirmed for the current agreement, **when** the manager continues onboarding, **then** profile management and agreement confirmation remain available while that participant's scheduling remains locked.

## Business rules

- Every authenticated person uses their own verified, system-wide unique normalized phone number.
- The household has at least one Family Manager.
- An authenticated identity is distinct from a person profile and a household account.
- The invitation is single-use, expiring, purpose-bound, and phone-bound.
- Invitation and authentication occur before agreement confirmation.
- Scheduling unlocks separately for each person only after that person's current-season agreement is Confirmed.

## Dependencies

- US-005 provides the authorized new-household invitation.
- US-002 provides secure SMS verification and personal authenticated sessions.

## Out of scope

- Open self-registration.
- Shared household credentials.
- Adding additional people or adult-to-scout relationships.
- Inviting a co-manager.
- Granting Young Adult Scout access.
- Performing agreement confirmation itself.
