# US-014 Review Confirmation Status

- **Epic:** Seasonal Conduct Agreement
- **Source use cases:** [UC-54](../../use-cases.md)
- **Primary actor:** Admin or Committee Member

**As an** authorized troop leader,  
**I want** to review agreement-confirmation status for the season,  
**so that** I can identify participants who are not ready to participate.

## Scope

Provide an authorized, filterable season agreement-status view with role-appropriate detail.

## Preconditions

- The leader is authenticated as Admin or Committee Member.
- US-011 has configured the season's agreement.

## Acceptance criteria

1. **Given** an Admin or Committee Member, **when** they open the agreement-status view, **then** every listed person is shown as Confirmed or Not Confirmed for the current agreement.
2. **Given** a Committee Member, **when** they review the list, **then** private account, session, confirmation-time, and facilitating-actor metadata is not exposed.
3. **Given** an Admin, **when** a confirmation was facilitated, **then** the confirmation time and acting identity are available.
4. **Given** status data, **when** the leader filters by household, person role, or confirmation status, **then** only matching people are shown.

## Business rules

- Committee sees status only; Admin receives the additional confirmation evidence defined by UC-54.
- Status is evaluated for the selected season and its current agreement-link identifier.
- Review authority does not grant authority to confirm for another person.

## Dependencies

- US-002 — authenticate and authorize the leader.
- US-011 — configure the season agreement.
- US-013 — record participant confirmations.

## Out of scope

- Editing profiles or household relationships.
- Marking a participant confirmed.
- Overriding participation eligibility.
