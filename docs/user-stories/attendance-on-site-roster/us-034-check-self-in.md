# US-034 — Check self in

- **Epic:** Attendance & On-Site Roster
- **Source use cases:** [UC-27](../../use-cases.md#use-case-27-authenticated-volunteer-checks-self-in), [UC-53](../../use-cases.md#use-case-53-agreement-confirmation-blocks-participation), [UC-58](../../use-cases.md#use-case-58-system-enforces-the-local-two-deep-coverage-rule-during-a-shift)
- **Primary actor:** Family Manager or Young Adult Scout

As an assigned authenticated volunteer,
I want to check myself in at the lot,
so that my presence and work time are recorded accurately.

## Scope

Create an immutable self check-in event during the real-time window and exercise the US-031 actual-coverage evaluator.

## Preconditions

- The volunteer is authenticated, active, and assigned to a published, non-closed shift.
- The volunteer's current-season agreement is Confirmed.
- Current server time can be evaluated against the shift start.

## Acceptance criteria

1. **Given** the authenticated volunteer is assigned and the server time is from 15 minutes before through 30 minutes after shift start, **when** they select Check In, **then** an immutable check-in is recorded with current server time and the acting identity.
2. **Given** the volunteer is not assigned, is outside the check-in window, or the shift is CLOSED, **when** self check-in is attempted, **then** no attendance event is created and the denial is explained.
3. **Given** the volunteer has not confirmed the current agreement link for the shift's season, **when** self check-in is attempted, **then** it is rejected without changing attendance and the agreement requirement is explained.
4. **Given** the volunteer is a scout and fewer than two adults are checked in, **when** self check-in is attempted, **then** it is rejected by the actual local two-deep evaluator.
5. **Given** the volunteer is an adult and check-in improves coverage, **when** all other validations pass, **then** check-in is allowed even before actual coverage is fully safe.
6. **Given** check-in succeeds, **when** confirmation is shown, **then** it displays the server-recorded time and current time on shift.
7. **Given** the same successful check-in request is repeated, **when** it is processed, **then** no duplicate event is created.

## Business rules

- Users cannot choose, backdate, or future-date attendance time.
- Young Adult Scout access does not make a scout count as an adult.
- Actual checked-in attendance, not assignments, controls local operating coverage.
- The participant's agreement confirmation cannot be overridden by Committee or Admin.
- Raw check-in events are immutable.

## Dependencies

- US-015
- US-019
- US-023
- US-024
- US-031

## Out of scope

- Checking in another person
- Checkout
- Post-shift adjustments or no-show decisions
- Verifying national-policy leader eligibility
