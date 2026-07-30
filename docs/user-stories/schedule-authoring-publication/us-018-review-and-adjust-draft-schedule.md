# US-018 Review and Adjust Draft Schedule

- **Epic:** Schedule Authoring and Publication
- **Source use cases:** [UC-22](../../use-cases.md)
- **Primary actor:** Committee Member

**As a** Committee Member,  
**I want** to review and adjust a generated draft schedule,  
**so that** timing, staffing, and special-event details are correct before publication.

## Scope

Browse draft shifts by date and modify, add, or remove draft shifts without exposing them to regular users.

## Preconditions

- The Committee Member is authenticated and authorized.
- US-017 has generated a draft season schedule.

## Acceptance criteria

1. **Given** a draft schedule, **when** the Committee Member browses by date, **then** each shift's timing, adult and scout targets, minimum operating headcount, and special-event configuration are available for review.
2. **Given** a draft shift, **when** the Committee Member adjusts timing or volunteer requirements, **then** the changed values remain in draft mode.
3. **Given** a gap or unnecessary shift, **when** the Committee Member adds an extra draft shift or removes an existing draft shift, **then** the draft schedule reflects the change.
4. **Given** any draft adjustment, **when** the change is saved, **then** no regular user sees it and no notification is sent.

## Business rules

- Minimum operating headcount cannot be configured below two.
- Adjusting minimum headcount cannot disable the approved local two-deep rule.
- Draft shifts are visible and manageable by Committee and Admin only.

## Dependencies

- US-017 — generate the draft season schedule.

## Out of scope

- Editing reusable templates.
- Publishing the season.
- Changing already published shifts.
