# US-036 — Committee checks in arriving volunteer

- **Epic:** Attendance & On-Site Roster
- **Source use cases:** [UC-29](../../use-cases.md#use-case-29-committee-checks-in-an-arriving-volunteer), [UC-53](../../use-cases.md#use-case-53-agreement-confirmation-blocks-participation), [UC-58](../../use-cases.md#use-case-58-system-enforces-the-local-two-deep-coverage-rule-during-a-shift)
- **Primary actor:** Committee Member

As an on-site Committee Member,
I want to check in an arriving assigned volunteer,
so that attendance can be recorded when no eligible working adult is available.

## Scope

Permit a Committee Member to record an assigned volunteer's immutable check-in under the same agreement, timing, duplicate, and actual-coverage rules as other real-time check-ins.

## Preconditions

- The Committee Member is authenticated and on site.
- The target is active and assigned to a published, non-closed shift.
- Current server time is available.

## Acceptance criteria

1. **Given** an assigned volunteer arrives from 15 minutes before through 30 minutes after shift start, has confirmed the current agreement, and is not checked in, **when** Committee selects Check In, **then** the system records current server time and identifies Committee as the actor.
2. **Given** Committee is not assigned or checked in to the shift, **when** an otherwise-valid real-time check-in is performed, **then** Committee authority permits it.
3. **Given** the target is outside the check-in window, **when** Committee attempts check-in, **then** the action is rejected and no normal attendance event is created.
4. **Given** the target's current agreement is not Confirmed, **when** Committee attempts check-in, **then** the action is rejected and cannot be overridden.
5. **Given** the target is a scout and fewer than two adults are checked in, **when** Committee attempts check-in, **then** the actual local two-deep evaluator rejects it.
6. **Given** the target is already checked in, **when** Committee repeats the request, **then** no duplicate attendance event is created.
7. **Given** check-in succeeds, **when** attendance is viewed, **then** relevant Family Managers can see the status and the audit trail identifies the Committee actor.

## Business rules

- Committee authority does not permit prospective or retroactive normal attendance events.
- Committee/Admin cannot override the participant's agreement requirement.
- Scout attendance requires actual local two-deep coverage.
- Attendance time is current server time and raw events are immutable.
- Post-shift verified attendance is represented by an adjustment, not a backdated check-in.

## Dependencies

- US-015
- US-019
- US-023
- US-024
- US-031
- US-034

## Out of scope

- Walk-in creation
- Backdated check-in
- Post-shift no-show marking or adjustments
- Proving national-policy adult-leader eligibility
