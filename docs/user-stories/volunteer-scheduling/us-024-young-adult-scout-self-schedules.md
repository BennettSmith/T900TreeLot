# US-024 Young Adult Scout Self-Schedules

- **Epic:** Volunteer Scheduling
- **Source use cases:** [UC-12, UC-16, UC-18, UC-19, and UC-53](../../use-cases.md)
- **Primary actor:** Young Adult Scout

**As a** Young Adult Scout,  
**I want** to sign myself up for and cancel my own shifts,  
**so that** I can manage my participation without acting for anyone else.

## Scope

Create scout-slot assignments for the authenticated scout's linked profile and allow that scout to cancel their own assignments.

## Preconditions

- The scout is authenticated with active Young Adult Scout access.
- Their profile belongs to at least one active household.
- A published shift is available and accepting signups.

## Acceptance criteria

1. **Given** a Young Adult Scout starts signup, **when** the form is shown and submitted, **then** no family-member selector is offered and the server forces the target to the identity's linked scout profile.
2. **Given** the scout is currently confirmed, not already assigned, and a scout slot is available, **when** signup is confirmed, **then** one self-created assignment appears on their schedule and every linked household's manager schedule.
3. **Given** the scout is unconfirmed, has no active linked household, is already assigned, or no scout slot is available, **when** signup is attempted, **then** the server rejects it without creating an assignment.
4. **Given** any assignment for the scout, **when** the scout confirms cancellation, **then** the assignment is removed once, capacity is freed, and the audit trail identifies the scout as actor.

## Business rules

- Client-submitted person or household IDs cannot change the self-service target.
- A Young Adult Scout can cancel their own assignment even if a Family Manager originally created it.
- Self-created assignments are manageable by the scout and Family Managers in any linked household.
- Family Managers retain visibility into the scout's assignments and cancellation activity.

## Dependencies

- US-002 — authenticate and resolve the linked scout profile.
- US-006 — create the household containing the scout profile.
- US-007 — establish household membership.
- US-009 — establish the linked-household relationships.
- US-010 — grant active Young Adult Scout access.
- US-015 — enforce agreement eligibility.
- US-019 — publish the season schedule.
- US-021 — provide schedule navigation.

## Out of scope

- Scheduling another person.
- Granting or revoking Young Adult Scout access.
- Attendance actions.
