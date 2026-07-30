# US-020 Add an Individual Published Shift

- **Epic:** Schedule Authoring and Publication
- **Source use cases:** [UC-3 and UC-24](../../use-cases.md)
- **Primary actor:** Committee Member

**As a** Committee Member,  
**I want** to add one published shift after season publication,  
**so that** an unexpected staffing need can be offered immediately.

## Scope

Create and immediately publish an individual shift in an already published season.

## Preconditions

- The Committee Member is authenticated and authorized to manage shifts.
- US-019 has published the season.

## Acceptance criteria

1. **Given** a published season, **when** the Committee Member creates a shift, **then** they provide date, time, adult and scout targets, minimum operating headcount, location, and optional notes.
2. **Given** a minimum operating headcount below two, **when** the shift is submitted, **then** the system rejects it without creating the shift.
3. **Given** valid shift details, **when** the Committee Member confirms creation for the published season, **then** the shift is published immediately and available to eligible users.
4. **Given** the shift is published, **when** notification delivery is queued, **then** active Family Managers and Young Adult Scouts eligible to fill it receive an in-app inbox notice without making publication depend on any external provider.

## Business rules

- The shift defines separate adult and scout targets and a minimum operating headcount of at least two.
- The local two-deep rule cannot be configured away.
- This post-publication path publishes immediately and generates the individual-addition in-app notification required by UC-24.
- Because recipients are a targeted eligible set, Groups.io is not used.

## Dependencies

- US-019 — publish the season schedule.

## Out of scope

- Bulk schedule generation.
- Adding a draft shift before season publication.
- Assigning volunteers to the new shift.
