# US-015 Enforce Participation Eligibility

- **Epic:** Seasonal Conduct Agreement
- **Source use cases:** [UC-53](../../use-cases.md)
- **Primary actor:** System

**As the** scheduling system,  
**I want** to gate participation by the selected person's current agreement confirmation,  
**so that** nobody participates without explicitly accepting the applicable rules.

## Scope

Enforce agreement eligibility for signup, assignment, check-in, and walk-in creation.

## Preconditions

- US-011 has configured the agreement for the shift's season.
- The attempted action identifies the selected participant.

## Acceptance criteria

1. **Given** a participation action, **when** the selected person is Confirmed for the shift's season and current link, **then** agreement eligibility passes and the action may continue through its other validations.
2. **Given** the selected person is Not Confirmed, **when** signup, assignment, check-in, or walk-in creation is attempted, **then** the server rejects the action without changing capacity or attendance.
3. **Given** a rejected action, **when** the response is shown, **then** it explains that the selected person must read and confirm the agreement and gives Family Managers an Agreement Center link.
4. **Given** a Committee Member or Admin attempts an override, **when** the person is Not Confirmed, **then** the system still rejects the action.

## Business rules

- Eligibility belongs to the selected participant, not the acting manager or household.
- Confirmation must match both the shift's season and the current agreement-link identifier.
- Committee and Admin cannot override the gate or mark a person confirmed without the explicit checkbox submission.
- Checkout remains allowed for an existing open attendance record if the agreement later changes.

## Dependencies

- US-011 — configure the agreement whose identifier is checked.
- US-013 — record explicit participant confirmation.

## Out of scope

- Other signup, capacity, role-slot, attendance-window, or coverage validations.
- Creating or changing confirmations.
