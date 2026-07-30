# US-003: change own phone number

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-2B](../../use-cases.md#use-case-2b-authenticated-person-changes-phone-number)
- **Primary actor:** Authenticated person

**As an** authenticated person who can still use my current phone number, **I want** to replace it with a newly verified unique number, **so that** I retain my identity and history while securing future sign-ins.

## Scope

Support the self-service phone-number change path when the current number remains available, including recent verification, uniqueness validation, atomic credential replacement, session revocation, and security notices.

## Preconditions

- The actor is a Family Manager, Young Adult Scout, Committee Member, or Admin with an existing authenticated identity.
- The actor can complete recent authentication using the current phone number.
- The proposed new number can receive SMS verification.

## Acceptance criteria

1. **Given** the actor can use the current number, **when** they request a phone-number change, **then** the system requires recent verification through that current number before continuing.
2. **Given** recent verification succeeded and the proposed number is not linked to another active identity, **when** the actor successfully verifies the new number by one-time code or magic link, **then** the system atomically replaces the phone number on the existing identity.
3. **Given** the change succeeds, **when** the replacement is committed, **then** all existing browser sessions are revoked, the old number stops working immediately, and the person must sign in with the new number.
4. **Given** the change succeeds, **when** security notifications are issued, **then** a notice is sent to both the old and new numbers.
5. **Given** the normalized new number is already linked to another person, **when** the actor attempts the change, **then** the request is rejected without revealing that person's identity and the old login remains unchanged.
6. **Given** the phone number is replaced, **when** the person signs in again, **then** the same internal identity, profile, roles, household memberships, assignments, and history remain intact.

## Business rules

- A normalized phone number is system-wide unique among active authenticated identities.
- Verification of both the current and new numbers is required for self-service change.
- The operation updates an existing identity; it does not create a replacement person or migrate history.
- All prior sessions are revoked after a successful change.
- Phone-number conflict messages do not reveal another identity or household.
- Authentication and security SMS messages are transactional and are not suppressed by notification preferences.

## Dependencies

- US-002 provides authenticated access and revocable secure sessions.

## Out of scope

- Recovery when the old number is unavailable.
- Reassigning a number from another person's active identity.
- Changing profiles, roles, memberships, assignments, or historical records.
- Removing authenticated access.
