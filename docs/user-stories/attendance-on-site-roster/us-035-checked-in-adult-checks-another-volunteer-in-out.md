# US-035 — Checked-in adult checks another volunteer in/out

- **Epic:** Attendance & On-Site Roster
- **Source use cases:** [UC-28](../../use-cases.md#use-case-28-working-adult-checks-in-another-volunteer), [UC-30](../../use-cases.md#use-case-30-working-adult-checks-out-another-volunteer), [UC-58](../../use-cases.md#use-case-58-system-enforces-the-local-two-deep-coverage-rule-during-a-shift)
- **Primary actor:** Authenticated Adult Volunteer

As an authenticated adult working on site,
I want to check another arriving or departing volunteer in or out,
so that the roster remains accurate for volunteers who cannot operate it themselves.

## Scope

Allow an eligible authenticated adult to record immutable, real-time attendance for assigned volunteers from any family.

## Preconditions

- The actor is an authenticated adult with a recorded adult person classification.
- The target is assigned to the shift.
- Check-in or checkout occurs in its applicable server-time window.

## Acceptance criteria

1. **Given** the actor is checked in to the target shift, **when** an assigned, agreement-confirmed volunteer arrives during the check-in window, **then** the actor can record that volunteer's check-in at current server time.
2. **Given** the actor was checked in to the immediately preceding shift at the same location, both shifts are in their respective handoff windows, and the target shift is in its check-in window, **when** an assigned volunteer arrives, **then** the actor can check that volunteer in.
3. **Given** the target is a scout and fewer than two adults would be checked in, **when** check-in is attempted, **then** it is rejected and no attendance event is created.
4. **Given** the target has not confirmed the current-season agreement, **when** check-in is attempted, **then** it is rejected without an override.
5. **Given** the actor is checked in to the same shift, the target has an open check-in, and the shift is from 15 minutes before through 30 minutes after scheduled end, **when** the actor chooses Check Out, **then** checkout is recorded at current server time and worked duration is shown.
6. **Given** the actor only has preceding-shift handoff authority, **when** they attempt to check someone out of the target shift, **then** the action is rejected.
7. **Given** an adult checkout reduces actual adult coverage below two, **when** checkout is recorded, **then** it succeeds and triggers noncompliant coverage, urgent operational alerting, and the required stop-work response.
8. **Given** the target belongs to another family, **when** every roster rule passes, **then** family relationship does not prevent the action and the actor is recorded.

## Business rules

- Check-in uses the target shift's window; checkout uses the target shift's checkout window.
- Preceding-shift authority applies only to check-in during overlapping handoff windows.
- The actor's authentication role alone is insufficient; the actor must have an adult person classification.
- Young Adult Scouts can act only on themselves.
- All events use server time, are immutable, and are idempotent.
- Adult checkout is never blocked by a resulting coverage violation.

## Dependencies

- US-015
- US-019
- US-023
- US-024
- US-031
- US-034

## Out of scope

- Walk-in creation
- Committee attendance authority
- Backdated events or post-shift corrections
- Household relationship authorization for roster actions
