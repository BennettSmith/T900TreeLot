# US-030 — Review staffing alerts dashboard

- **Epic:** Staffing Intelligence & Shift Operations
- **Source use cases:** [UC-42](../../use-cases.md#use-case-42-committee-reviews-staffing-alerts-dashboard)
- **Primary actor:** Committee Member

As a Committee Member,
I want to review prioritized staffing alerts,
so that I can address the most urgent projected shortfalls first.

## Scope

Provide an automatically refreshed dashboard of published shifts whose projected staffing is LOW or CRITICAL.

## Preconditions

- A season schedule has been published.
- Active assignments and each shift's staffing requirements are available.

## Acceptance criteria

1. **Given** one or more shifts need attention, **when** the Committee Member views navigation, **then** an alert badge indicates that staffing alerts exist.
2. **Given** LOW and CRITICAL shifts, **when** the dashboard opens, **then** shifts are grouped by severity and ordered soonest first within each group.
3. **Given** an alerted shift, **when** its entry is displayed, **then** the exact adult, scout, and total-person shortfall is shown and the system identifies whether minimum headcount, scheduled local two-deep coverage, or both are unresolved.
4. **Given** an alerted shift, **when** the Committee Member chooses an available action, **then** they can open shift details, share its signup link, send a targeted staffing reminder, or begin a critical coverage alert when closure is possible.
5. **Given** an assignment is added or removed, **when** projected coverage changes, **then** the dashboard updates the alert's severity or removes the resolved alert without treating the change as actual attendance.

## Business rules

- Dashboard severity is based on projected coverage from active assignments.
- Projected minimum headcount and scheduled local two-deep coverage are evaluated separately.
- Targeted staffing reminders go only to volunteers eligible for that shift.
- A critical coverage alert is troop-wide and distinct from a targeted reminder.
- Actual checked-in coverage may create CLOSURE REQUIRED even when the earlier projection was safe.

## Dependencies

- US-019
- US-023
- US-024
- US-029

## Out of scope

- Recording attendance
- Sending the critical alert itself
- Recording closure or reopening
- Determining national-policy eligibility of scheduled adults
