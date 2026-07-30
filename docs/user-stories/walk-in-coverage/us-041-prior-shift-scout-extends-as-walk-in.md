# US-041 — Prior-shift scout extends as walk-in

- **Epic:** Walk-In Coverage
- **Source use cases:** [UC-35](../../use-cases.md#use-case-35-scout-from-prior-shift-extends-as-walk-in), [UC-53](../../use-cases.md#use-case-53-agreement-confirmation-blocks-participation), [UC-58](../../use-cases.md#use-case-58-system-enforces-the-local-two-deep-coverage-rule-during-a-shift)
- **Primary actor:** Checked-In Authenticated Adult on the next shift

As a checked-in authenticated adult on the next shift,
I want to add a scout who finished the prior shift as a walk-in,
so that the scout can extend their help while each shift's hours remain separate.

## Scope

Transition a physically present scout from completed scheduled attendance on the prior shift to a distinct real-time walk-in record on an in-progress next shift.

## Preconditions

- The scout has checked out of the prior scheduled shift.
- The next shift has started, has not ended, and is not CLOSED.
- An authenticated adult is checked in to the next shift.

## Acceptance criteria

1. **Given** the scout has checked out of the prior shift and remains on site, **when** a checked-in authenticated adult on the next shift adds them, **then** a distinct next-shift walk-in and immediate check-in are recorded at current server time.
2. **Given** the scout still has an open attendance record on the prior shift, **when** extension is attempted, **then** the system rejects it until the prior attendance is checked out.
3. **Given** the scout has not confirmed the agreement in effect for the next shift's season, **when** extension is attempted, **then** it is rejected without override.
4. **Given** fewer than two adults are checked in to the next shift, **when** extension is attempted, **then** the actual local two-deep evaluator rejects the scout walk-in.
5. **Given** the next shift is not in progress or is CLOSED, **when** extension is attempted, **then** no walk-in or check-in is created.
6. **Given** extension succeeds, **when** hours are calculated, **then** prior scheduled hours and next-shift walk-in hours are credited separately and neither interval is duplicated.
7. **Given** the scout belongs to multiple households, **when** attendance history is viewed by an authorized manager, **then** both work intervals are visible against the scout's single profile.
8. **Given** the scout later departs during the next shift's checkout window, **when** an eligible checked-in adult checks them out, **then** checkout follows the standard same-shift actor and server-time rules.

## Business rules

- A prior-shift scout is not automatically carried into the next shift.
- The prior attendance record must close before the next walk-in record opens.
- The next-shift walk-in uses current server time and cannot be backdated to bridge a gap.
- The actor must be checked in to the next shift; preceding-shift handoff authority is insufficient for walk-in creation.
- Scheduled and walk-in hours remain separate but both count toward totals.
- Scout walk-in creation requires actual local two-deep coverage.

## Dependencies

- US-015
- US-019
- US-031
- US-035
- US-040

## Out of scope

- Automatically extending attendance across shifts
- Merging scheduled and walk-in hours
- Adding an adult extension under scout walk-in authority
- Post-shift reconstruction of an extension
