# US-032 — Send critical coverage alert

- **Epic:** Staffing Intelligence & Shift Operations
- **Source use cases:** [UC-57](../../use-cases.md#use-case-57-committee-sends-a-critical-coverage-alert)
- **Primary actor:** Committee Member or Admin

As a Committee Member or Admin,
I want to send a troop-wide critical coverage alert,
so that volunteers can resolve an unsafe projected shift before it must close.

## Scope

Preview, confirm, publish, deliver, audit, and resolve a critical alert for one unresolved projected coverage condition.

## Preconditions

- The Staffing Alerts dashboard identifies a published shift as CRITICAL.
- The sender has Committee or Admin authority.
- The shift has not started and is not CLOSED.

## Acceptance criteria

1. **Given** a CRITICAL shift, **when** the sender starts an alert, **then** the preview shows date, time, location, adult, scout, and total counts, the unresolved rule, closure warning, and direct signup link.
2. **Given** the preview, **when** the sender enters a response deadline, **then** the system accepts only a deadline before the shift starts.
3. **Given** a valid preview, **when** the sender confirms, **then** one canonical high-priority announcement is placed in every active Family Manager and Young Adult Scout in-app inbox.
4. **Given** Groups.io is enabled, **when** the alert is confirmed, **then** the same alert is queued for that channel without making inbox success depend on it.
5. **Given** the same unresolved condition was already alerted, **when** another send is requested, **then** explicit duplicate confirmation is required and the send is audited.
6. **Given** later signups satisfy every projected operating rule, **when** coverage is reevaluated, **then** the status updates immediately and one deduplicated coverage-secured update is sent to the original recipient set.

## Business rules

- A critical alert is troop-wide, not a targeted staffing reminder, and is not suppressed by shift-reminder preferences.
- The canonical announcement and each delivery attempt have independent, idempotent status.
- External provider calls do not occur inside the state-changing transaction; delivery uses the transactional outbox.
- Alert content identifies whether minimum headcount, scheduled local two-deep coverage, or both are unresolved.
- Resolving projected coverage does not prove safe actual attendance at shift time.

## Dependencies

- US-030

## Out of scope

- Targeted staffing reminders
- Recording attendance
- Automatically or manually closing the shift
- Verifying recipients read the alert
