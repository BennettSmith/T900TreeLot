# US-040 — Checked-in adult adds walk-in scout

- **Epic:** Walk-In Coverage
- **Source use cases:** [UC-34](../../use-cases.md#use-case-34-checked-in-authenticated-adult-adds-walk-in-scout), [UC-53](../../use-cases.md#use-case-53-agreement-confirmation-blocks-participation), [UC-58](../../use-cases.md#use-case-58-system-enforces-the-local-two-deep-coverage-rule-during-a-shift)
- **Primary actor:** Checked-In Authenticated Adult

As a checked-in authenticated adult,
I want to add an eligible unscheduled scout to my current shift,
so that available on-site help can be recorded without separate Committee approval.

## Scope

Allow an authenticated adult checked in to the current in-progress shift to create and immediately check in a scout walk-in.

## Preconditions

- The actor is an authenticated adult checked in to the target shift.
- The target shift is in progress and not CLOSED.
- The scout is physically present and has a person profile.

## Acceptance criteria

1. **Given** an eligible scout is present during the actor's in-progress shift, **when** the actor selects the scout and confirms Add Walk-In, **then** a walk-in assignment and immutable check-in are recorded at current server time with the actor identified.
2. **Given** the actor is a checked-in authenticated adult on that shift, **when** the valid addition is submitted, **then** no separate Committee approval is required.
3. **Given** the actor is not checked in to the target shift or only worked the immediately preceding shift, **when** walk-in creation is attempted, **then** it is rejected.
4. **Given** the scout has not confirmed the current agreement, **when** walk-in creation is attempted, **then** it is rejected without override.
5. **Given** fewer than two adults are checked in, **when** the scout walk-in is submitted, **then** the actual local two-deep evaluator rejects it without creating either record.
6. **Given** the shift is before start, after end, or CLOSED, **when** the actor attempts the addition, **then** it is rejected.
7. **Given** the scout is already scheduled, checked in, or recorded as a walk-in for that shift, **when** another addition is attempted, **then** duplicate-person validation rejects it.
8. **Given** planned scout capacity is full, **when** every walk-in rule passes, **then** the walk-in may exceed capacity and the roster records the resulting state.

## Business rules

- This authority belongs to authenticated adults checked in to the same in-progress shift; handoff authority does not apply.
- The target must be a scout; only Committee/Admin can add an adult walk-in.
- Agreement confirmation is checked for the selected scout.
- Scout addition must preserve actual local two-deep coverage.
- Walk-in and check-in creation are atomic, use server time, and remain distinguishable from scheduled assignments.
- Checkout follows normal real-time checkout rules.

## Dependencies

- US-015
- US-019
- US-031
- US-034
- US-035

## Out of scope

- Adding adult walk-ins
- Pre-shift walk-in reservation
- Committee approval
- Post-shift or backdated creation
