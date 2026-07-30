# US-019 Publish Season Schedule

- **Epic:** Schedule Authoring and Publication
- **Source use cases:** [UC-23](../../use-cases.md)
- **Primary actor:** Committee Chair

**As a** Committee Chair,  
**I want** to publish the reviewed season schedule in one operation,  
**so that** authenticated users can see shifts and begin signing up.

## Scope

Review the final summary, select publication-notification options, and atomically publish all draft shifts for the season.

## Preconditions

- The Committee Chair is authenticated and authorized.
- US-018 has produced a reviewed draft schedule ready for publication.

## Acceptance criteria

1. **Given** the reviewed draft, **when** the Chair starts publication, **then** the system shows total shifts, volunteer slots, and special events plus the option to send an SMS summary.
2. **Given** the Chair confirms publication, **when** the operation succeeds, **then** all season shifts become published simultaneously, visible, and available for signup.
3. **Given** SMS notification is selected, **when** publication succeeds, **then** each active Family Manager and Young Adult Scout phone number receives one summary containing the shift count, date range, and special-event highlights.
4. **Given** SMS notification is not selected, **when** publication succeeds, **then** shifts still become visible and no publication SMS is required.

## Business rules

- Publication sends at most one comprehensive SMS per active recipient, not one per shift.
- Special events are prominently marked in the web app and highlighted in the notification.
- External provider calls do not run inside the publication transaction; delivery failures do not undo web publication.

## Dependencies

- US-018 — review and adjust the draft schedule.

## Out of scope

- Creating templates or generating the draft.
- Volunteer signup.
- Adding an unplanned shift after publication.
