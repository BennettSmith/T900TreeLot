# US-044 — Send Automated Shift Reminders

- **Epic:** Communications & Reminders
- **Source use cases:** [UC-6](../../use-cases.md#use-case-6-shift-reminders-automated)
- **Primary actor:** System

**As the** scheduling system, **I want** to send correctly routed reminders for upcoming assignments, **so that** volunteers and responsible managers are less likely to miss shifts.

## Scope

Hourly discovery, recipient routing, preference enforcement, SMS content, and idempotent delivery for shifts approximately 24 hours away.

## Preconditions

- A published shift has a confirmed, active assignment.
- Assignment origin and current active household relationships are available.

## Acceptance criteria

1. **Given** the hourly process runs, **when** a shift enters the configured approximately-24-hour window, **then** each confirmed assignment is considered once using the configured tree-lot time zone and injected clock.
2. **Given** a household-owned assignment, **when** recipients are resolved, **then** active Family Managers of the originating household are selected.
3. **Given** a Young Adult Scout-created assignment, **when** recipients are resolved, **then** active Family Managers in every linked household are selected; if the volunteer is a Young Adult Scout, that scout is also selected directly.
4. **Given** a selected recipient has reminders enabled, **when** delivery is processed, **then** the SMS names the assigned person, states shift date and time, and links to shift details; opted-out recipients receive no reminder.
5. **Given** the process repeats or a delivery is retried, **when** the same reminder key is encountered, **then** successful SMS is not duplicated and failures can be retried safely.

## Business rules

- Reminder preference applies only to nonessential operational reminders, not authentication or security messages.
- Routing uses active relationships and assignment origin, never client-supplied household IDs.
- External SMS runs outside database transactions through the transactional outbox.
- Times are stored in UTC and presented using `TREE_LOT_TIME_ZONE`.

## Dependencies

- US-002
- US-009
- US-010
- US-019
- US-023
- US-024

## Out of scope

- Troop announcements, critical coverage alerts, and targeted staffing reminders
- Assignment creation, cancellation, or preference editing
