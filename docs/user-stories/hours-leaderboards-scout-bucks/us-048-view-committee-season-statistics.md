# US-048 — View Committee Season Statistics

- **Epic:** Hours, Leaderboards & Scout Bucks
- **Source use cases:** [UC-39](../../use-cases.md#use-case-39-committee-views-season-statistics)
- **Primary actor:** Committee Member

**As a** Committee Member, **I want** to review season-wide participation and coverage statistics, **so that** the troop can evaluate operations and recognize contributors.

## Scope

Season summary, coverage measures, top contributors, and authorized report drill-downs.

## Preconditions

- The actor has active Committee or Admin authority.
- Corrected attendance and reporting relationships are available for the season.

## Acceptance criteria

1. **Given** the actor opens Season Stats, **when** the report loads, **then** it shows total distinct volunteers, family units, corrected hours, and completed shifts.
2. **Given** scheduled and walk-in activity exists, **when** coverage statistics are calculated, **then** the report shows their proportions, no-show rate, critical coverage alerts, prevented openings, and closed shifts.
3. **Given** ranked participation exists, **when** summaries display, **then** top contributors in each supported category are shown without duplicate people.
4. **Given** the actor selects a detail, **when** drill-down opens, **then** authorization is rechecked and the underlying corrected records reconcile to the summary.

## Business rules

- Corrected attendance drives reports; immutable events and adjustment audit history remain available.
- People linked through multiple households or roles are counted once.
- Family reporting uses explicit family units.

## Dependencies

- US-007
- US-009
- US-037

## Out of scope

- Attendance correction entry
- Operational shift closure decisions
- Scout Bucks finalization
