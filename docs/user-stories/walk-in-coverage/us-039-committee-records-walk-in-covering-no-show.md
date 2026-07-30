# US-039 — Committee records walk-in covering no-show

- **Epic:** Walk-In Coverage
- **Source use cases:** [UC-33](../../use-cases.md#use-case-33-walk-in-covers-for-no-show), [UC-53](../../use-cases.md#use-case-53-agreement-confirmation-blocks-participation), [UC-58](../../use-cases.md#use-case-58-system-enforces-the-local-two-deep-coverage-rule-during-a-shift)
- **Primary actor:** Committee Member

As a Committee Member,
I want to record an eligible walk-in who covers a no-show,
so that the shift's actual staffing and each volunteer's attendance remain accurate.

## Scope

Create a walk-in record and immediate check-in for an adult or scout during an in-progress shift, and link the operational context to a separately retained no-show assignment.

## Preconditions

- The shift has started, has not ended, and is not CLOSED.
- The Committee Member is authenticated and authorized.
- A scheduled volunteer has not arrived and another known person is physically present to cover.

## Acceptance criteria

1. **Given** an in-progress shift, **when** Committee searches for and selects an eligible adult or scout, adds the coverage note, and confirms, **then** a distinct walk-in assignment and immutable check-in are recorded at current server time.
2. **Given** the selected person has not confirmed the current agreement, **when** addition is attempted, **then** no walk-in or attendance record is created and Committee cannot override the denial.
3. **Given** the selected person is a scout and fewer than two adults are checked in, **when** addition is attempted, **then** the actual local two-deep evaluator rejects it.
4. **Given** the selected person is an adult who improves deficient coverage, **when** all other validations pass, **then** the walk-in may be added.
5. **Given** the original assigned volunteer did not arrive and the check-in window has closed, **when** Committee records the replacement, **then** the original assignment is retained, marked No Show with an explanation, and not converted into the walk-in.
6. **Given** the shift has not started, has ended, or is CLOSED, **when** Committee attempts to add the walk-in, **then** the action is rejected.
7. **Given** the person is already participating in the shift or has an ineligible role classification, **when** addition is attempted, **then** duplicate-person and role checks reject it.

## Business rules

- Walk-ins are real-time exceptions allowed only while a shift is in progress.
- Creation and automatic check-in use current server time and cannot be backdated.
- Walk-ins may exceed planned capacity, but the over-capacity state and actor are recorded.
- Walk-in and scheduled attendance remain distinguishable in reporting.
- The no-show assignment remains in the audit trail.
- Walk-in checkout follows normal real-time checkout actor and timing rules.

## Dependencies

- US-015
- US-019
- US-031
- US-036

## Out of scope

- Adding a walk-in before or after the shift
- Replacing or deleting the original assignment
- Backdated attendance
- Checkout implementation
