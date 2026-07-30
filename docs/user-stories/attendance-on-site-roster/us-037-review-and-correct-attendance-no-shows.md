# US-037 — Review and correct attendance/no-shows

- **Epic:** Attendance & On-Site Roster
- **Source use cases:** [UC-31](../../use-cases.md#use-case-31-committee-reviews-shift-attendance)
- **Primary actor:** Committee Member

As a Committee Member,
I want to review completed-shift attendance and record corrections or no-shows,
so that credited hours are accurate without rewriting real-time history.

## Scope

Review attendance outcomes and create separate, reasoned, audited adjustments or no-show records after real-time windows have closed.

## Preconditions

- The shift is complete or its applicable real-time attendance window has closed.
- Real-time check-in and checkout events, if any, are available.
- The actor has Committee or Admin attendance-adjustment authority.

## Acceptance criteria

1. **Given** a completed shift, **when** Committee opens attendance review, **then** the summary shows checked in, checked out, still checked in, and pending/no-show counts.
2. **Given** a volunteer record, **when** it is reviewed, **then** it shows immutable check-in and checkout times, acting users, calculated hours, adjustments, and relevant coverage transitions or closure.
3. **Given** a volunteer forgot to check out, **when** Committee enters an approved departure time or hours plus a required reason, **then** a separate audited adjustment determines corrected hours without creating or backdating checkout.
4. **Given** later evidence verifies attendance without a real-time check-in, **when** Committee records corrected attendance with a required explanation, **then** a separate adjustment is created and no normal check-in is fabricated.
5. **Given** an assigned volunteer did not arrive and the check-in window is closed, **when** Committee marks No Show and adds a follow-up note, **then** the no-show is recorded while the original assignment remains in history.
6. **Given** a correction is saved, **when** hours are reported, **then** corrected hours are used while original events, the adjustment, reason, actor, and creation time remain auditable.

## Business rules

- Raw real-time attendance events are immutable.
- Adjustments are separate records and never backdate or replace normal events.
- Every adjustment requires a reason and records actor and creation time.
- No-show marking is allowed only after the check-in window closes.
- Corrected hours cannot be negative, and checkout or adjusted departure must be later than arrival.
- Coverage transitions and shift closure history remain visible during review.

## Dependencies

- US-034
- US-035
- US-036

## Out of scope

- Real-time attendance entry
- Deleting raw events
- Changing assignments
- Calculating Scout Bucks settlements
