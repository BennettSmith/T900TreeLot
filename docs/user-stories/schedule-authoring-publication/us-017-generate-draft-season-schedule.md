# US-017 Generate Draft Season Schedule

- **Epic:** Schedule Authoring and Publication
- **Source use cases:** [UC-21](../../use-cases.md)
- **Primary actor:** Committee Chair

**As a** Committee Chair,  
**I want** to generate a complete draft season schedule from templates,  
**so that** the season can be prepared efficiently before anyone signs up.

## Scope

Configure a season, apply regular and special-event templates, add closed-day exceptions, preview, and generate draft shifts in bulk.

## Preconditions

- The Committee Chair is authenticated and authorized.
- US-016 has provided the templates selected for generation.

## Acceptance criteria

1. **Given** a new season schedule, **when** the Chair configures its name, date range, default location, template selection, special-event dates, and closed-day exceptions, **then** the system previews total shifts and volunteer slots.
2. **Given** a reviewed preview, **when** the Chair confirms generation, **then** all generated shifts are created in draft mode with values snapshotted from their templates.
3. **Given** a special-event template assigned to a date, **when** generation completes, **then** the corresponding draft shift carries its special-event configuration and higher volunteer requirements.
4. **Given** draft generation succeeds, **when** regular users browse schedules, **then** they cannot see the draft shifts and no notification has been sent.

## Business rules

- Regular templates apply according to configured weekdays; special-event templates apply to selected dates.
- Closed-day exceptions do not produce regular shifts.
- Bulk generation creates drafts only and sends no notifications.

## Dependencies

- US-016 — create reusable shift templates.

## Out of scope

- Adjusting generated shifts after generation.
- Publishing the season.
