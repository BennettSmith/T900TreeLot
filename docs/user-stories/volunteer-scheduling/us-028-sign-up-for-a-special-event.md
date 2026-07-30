# US-028 Sign Up for a Special Event

- **Epic:** Volunteer Scheduling
- **Source use cases:** [UC-25](../../use-cases.md)
- **Primary actor:** Family Manager

**As a** Family Manager,  
**I want** to recognize and sign eligible household members up for a special event,  
**so that** high-priority tree-lot work receives the help it needs.

## Scope

Discover a published special-event shift, review its distinctive details, and use household-member signup.

## Preconditions

- The Family Manager is authenticated and manages an active household.
- The special-event shift is published and accepting signups.
- The selected participant meets current agreement eligibility.

## Acceptance criteria

1. **Given** a published special-event shift, **when** it appears in schedule views, **then** it has a star, an "ALL HANDS NEEDED" badge, higher volunteer requirements, and any special notes.
2. **Given** a schedule-publication notification, **when** special events exist and notification was selected, **then** the summary highlights those events.
3. **Given** the manager opens a special event, **when** they review availability, **then** they can select themselves or another eligible household adult or scout for a matching open slot.
4. **Given** a valid selection, **when** signup is confirmed, **then** the selected person is assigned using the same authority, agreement, duplicate, role-slot, and capacity rules as ordinary shifts.

## Business rules

- Special-event presentation does not bypass ordinary signup validation.
- The selected participant, not merely the acting manager, must be Confirmed for the shift's current agreement.
- Special events are prominently marked in the web app and publication notification.

## Dependencies

- US-002 — authenticate the Family Manager.
- US-006 — create the household.
- US-007 — establish household membership.
- US-015 — enforce agreement eligibility.
- US-019 — publish and highlight the special event.
- US-021 — provide schedule navigation.
- US-023 — sign up the selected household member.

## Out of scope

- Creating or publishing the special-event shift.
- Special-event attendance or reminders.
- Staffing-dashboard monitoring.
