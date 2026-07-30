# US-022 View Household Schedule

- **Epic:** Volunteer Scheduling
- **Source use cases:** [UC-8](../../use-cases.md)
- **Primary actor:** Family Manager

**As a** Family Manager,  
**I want** to view one combined schedule for household members,  
**so that** I can coordinate their upcoming volunteer commitments.

## Scope

Display upcoming assignments for active adults and scouts in a managed household, with person filtering and authorized assignment details.

## Preconditions

- The Family Manager is authenticated.
- The manager has authority over the selected household.
- US-019 has published the relevant season schedule.

## Acceptance criteria

1. **Given** a managed household, **when** the manager opens its schedule, **then** all active household members and their upcoming assignments are shown.
2. **Given** an assignment, **when** it is displayed, **then** the manager sees the assigned person, adult/scout slot type, date, time, and location.
3. **Given** a selected household member, **when** the manager filters the schedule, **then** only that person's assignments are shown.
4. **Given** an assignment is opened, **when** the manager views its actions, **then** cancellation is offered only when household ownership rules authorize it.

## Business rules

- Shared scouts retain one profile and schedule; assignments are visible to all linked households.
- Visibility does not grant cancellation authority over another household's assignment.
- Young Adult Scout assignments remain visible and manageable as allowed by origin rules.

## Dependencies

- US-002 — authenticate the Family Manager.
- US-006 — create the household.
- US-007 — establish household membership.
- US-009 — link a shared scout whose assignments must appear across households.
- US-019 — publish the season schedule.
- US-021 — provide season and week navigation.

## Out of scope

- Creating or cancelling an assignment.
- Attendance and hours history.
