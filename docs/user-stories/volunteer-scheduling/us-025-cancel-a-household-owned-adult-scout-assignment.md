# US-025 Cancel a Household-Owned Adult/Scout Assignment

- **Epic:** Volunteer Scheduling
- **Source use cases:** [UC-13, UC-14, and UC-15](../../use-cases.md)
- **Primary actor:** Family Manager

**As a** Family Manager,  
**I want** to cancel an adult or scout assignment owned by my household,  
**so that** capacity is released without changing another household's commitments.

## Scope

Cancel household-owned assignments under origin-based authority and prevent unauthorized cross-household cancellation.

## Preconditions

- The Family Manager is authenticated and manages the selected household.
- The assignment is visible from the household schedule.

## Acceptance criteria

1. **Given** an adult or scout assignment owned by the manager's household, **when** the manager reviews the warning and confirms cancellation, **then** the assignment is cancelled and its slot is freed once.
2. **Given** a shared scout assignment owned by another household, **when** this manager views it, **then** no cancellation control is offered and the originating household is identified.
3. **Given** an unauthorized direct cancellation request for another household's assignment, **when** the server validates ownership, **then** it rejects the request without changing the assignment or capacity.
4. **Given** cancellation makes projected coverage CRITICAL, **when** cancellation succeeds, **then** it remains allowed, the manager receives a prominent warning, and Committee/Admin receives a staffing alert.

## Business rules

- Household-owned assignments may be cancelled only by managers of the originating household, except for audited Committee/Admin override.
- A Young Adult Scout-created assignment is not household-owned and may be managed by any linked household.
- Repeated cancellation requests are idempotent and free capacity only once.
- Every successful cancellation records the acting user.

## Dependencies

- US-002 — authenticate the Family Manager.
- US-006 — create the household.
- US-007 — establish household and shared-scout membership.
- US-009 — link the shared scout used to exercise cross-household ownership rules.
- US-019 — provide the published assignment's schedule.
- US-021 — provide season and week navigation.
- US-022 — expose the household schedule and assignment details.

## Out of scope

- Young Adult Scout self-cancellation.
- Committee/Admin override cancellation.
- Reassigning a replacement volunteer.
