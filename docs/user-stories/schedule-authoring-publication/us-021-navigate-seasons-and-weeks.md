# US-021 Navigate Seasons and Weeks

- **Epic:** Schedule Authoring and Publication
- **Source use cases:** [UC-44](../../use-cases.md)
- **Primary actor:** Any user

**As a** schedule viewer,  
**I want** the app to select a relevant season and week and let me navigate valid weeks,  
**so that** I can quickly reach the schedule content I need.

## Scope

Resolve the default season/week and provide adjacent-week and week-picker navigation.

## Preconditions

- The web app can determine season status and dates.
- US-019 has published a schedule, or a draft/latest completed season exists for the defined fallback states.

## Acceptance criteria

1. **Given** an active published season, **when** a user opens the app, **then** its current week loads by default.
2. **Given** a displayed season, **when** the user navigates, **then** adjacent-week controls and a picker offer only valid season weeks and show each week's shift count.
3. **Given** there is no active season, **when** the app loads, **then** the most recently completed season is selected.
4. **Given** only a draft season exists, **when** a regular user opens the app, **then** they see "Schedule not yet published" without draft details, while Committee can view the draft within its authority.

## Business rules

- Navigation is constrained to the selected season's valid dates.
- Family Managers and Young Adult Scouts see only published shifts.
- Committee and Admin may see draft schedules.

## Dependencies

- US-019 — publish the season schedule used by regular viewers.

## Out of scope

- Authoring or publishing shifts.
- Staffing-status calculations.
- Volunteer signup.
