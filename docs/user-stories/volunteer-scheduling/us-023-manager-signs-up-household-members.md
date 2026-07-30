# US-023 Manager Signs Up Household Members

- **Epic:** Volunteer Scheduling
- **Source use cases:** [UC-9, UC-10, UC-11, UC-17, UC-18, UC-19, and UC-53](../../use-cases.md)
- **Primary actor:** Family Manager

**As a** Family Manager,  
**I want** to select an eligible household adult or scout for an available shift,  
**so that** the intended person—not merely my login—is assigned.

## Scope

Create a household-owned adult or scout assignment while enforcing authority, activity, agreement, publication, duplicate, role-slot, and capacity rules.

## Preconditions

- The manager is authenticated and manages an active household.
- The selected person is an active member of that household.
- The shift belongs to a published schedule and is accepting signups.

## Acceptance criteria

1. **Given** an available published shift, **when** the manager starts signup, **then** eligible household adults and scouts are grouped by role and ineligible members are disabled with an explanation.
2. **Given** an eligible adult or scout with current agreement confirmation and a matching open slot, **when** the manager confirms, **then** one assignment records the selected volunteer, acting manager, and originating household.
3. **Given** the person is unconfirmed, already assigned, the household is inactive, the shift is full or unavailable, or no matching role slot remains, **when** signup is submitted, **then** the server rejects it and creates no assignment.
4. **Given** concurrent requests for the last matching slot, **when** they are processed, **then** capacity and duplicate invariants are enforced atomically and no overbooking occurs.

## Business rules

- The selected volunteer's agreement status controls eligibility; the manager need not be personally confirmed to schedule someone who is.
- A person may have at most one assignment per shift and cannot occupy both slot types.
- Assignments target person profiles; the volunteer need not have login access.
- Signup may remain open while projected coverage is CRITICAL, provided a matching slot is available.

## Dependencies

- US-002 — authenticate the Family Manager.
- US-006 — create the household.
- US-007 — establish household membership.
- US-015 — enforce agreement eligibility.
- US-019 — publish the season schedule.
- US-021 — provide schedule navigation.

## Out of scope

- Young Adult Scout self-signup.
- Assignment cancellation.
- Committee/Admin override assignment.
