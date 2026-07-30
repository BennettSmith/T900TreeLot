# US-013 Confirm Agreement for a Participant

- **Epic:** Seasonal Conduct Agreement
- **Source use cases:** [UC-51](../../use-cases.md)
- **Primary actor:** Participant

**As a** participant,  
**I want** to explicitly confirm the current season's agreement,  
**so that** my readiness is recorded separately from every other person.

## Scope

Record a participant's explicit confirmation either in their own authenticated session or through an authorized Family Manager's device.

## Preconditions

- US-011 has configured the agreement for the participant's season.
- The participant has opened and read the external agreement.
- An authenticated participant acts for themselves, or an authorized Family Manager facilitates for a managed person.

## Acceptance criteria

1. **Given** an authenticated adult or Young Adult Scout, **when** they select the agreement checkbox and submit, **then** the system records confirmation for their own person profile.
2. **Given** a Managed Scout or adult without login access, **when** an authorized Family Manager selects that participant and the participant explicitly selects the checkbox before submission, **then** the system records both participant and facilitating manager.
3. **Given** a successful submission, **when** the confirmation is stored, **then** it includes the person, season, current agreement-link identifier, boolean value, server timestamp, and acting identity.
4. **Given** one participant is confirmed, **when** other household members are viewed, **then** their statuses remain unchanged.

## Business rules

- Each person confirms separately for each season and current agreement link.
- Merely opening the agreement does not confirm it.
- No typed signature, guardian-consent record, paper form, uploaded scan, or agreement document is collected.
- A Family Manager may facilitate but cannot submit confirmation without the participant's explicit checkbox action.

## Dependencies

- US-002 — authenticate the participant or facilitating manager.
- US-006 — provide the participant profile.
- US-007 — establish household membership.
- US-009 — authorize facilitated confirmation.
- US-011 — configure the current agreement.

## Out of scope

- Setting the agreement link.
- Participation actions such as signup, check-in, or walk-in creation.
