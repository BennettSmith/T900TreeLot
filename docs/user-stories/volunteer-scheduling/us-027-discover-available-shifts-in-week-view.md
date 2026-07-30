# US-027 Discover Available Shifts in Week View

- **Epic:** Volunteer Scheduling
- **Source use cases:** [UC-43](../../use-cases.md)
- **Primary actor:** Family Manager

**As a** Family Manager,  
**I want** to compare staffing needs in week view,  
**so that** I can find shifts where my household's help is most valuable.

## Scope

Show published shifts with staffing indicators, highlight critical need, and provide a path to shift details and signup.

## Preconditions

- The Family Manager is authenticated.
- US-019 has published the season schedule.
- US-021 has selected a valid season and week.

## Acceptance criteria

1. **Given** a published week, **when** the manager opens week view, **then** each shift shows a clear fully staffed, needs-help, critical/closure-required, or closed indicator.
2. **Given** multiple shifts, **when** the manager scans the week, **then** high-need shifts are visually distinguishable from fully staffed shifts.
3. **Given** a critical shift, **when** the manager opens it, **then** the details explain the staffing need and offer signup while the shift remains eligible to accept assignments.
4. **Given** a closed shift, **when** it is displayed, **then** it is marked as no longer operating and does not offer signup.

## Business rules

- Indicators reflect the shift's current staffing and operational status.
- CRITICAL status does not itself close signup; additional eligible volunteers may resolve the deficiency.
- CLOSED shifts disable participation actions.

## Dependencies

- US-002 — authenticate the Family Manager.
- US-006 — create the household.
- US-007 — establish household membership.
- US-019 — publish the season schedule.
- US-021 — provide season and week navigation.
- US-029 — calculate the projected staffing indicators shown in week view.

## Out of scope

- Calculating or administering staffing alerts.
- Closing or reopening shifts.
- Completing the signup transaction.
