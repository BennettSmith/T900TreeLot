# US-045 — View Scout Hours and Stats

- **Epic:** Hours, Leaderboards & Scout Bucks
- **Source use cases:** [UC-36](../../use-cases.md#use-case-36-scout-hours-and-stats-are-viewed)
- **Primary actor:** Family Manager or Young Adult Scout

**As a** Family Manager or Young Adult Scout, **I want** to view authorized scout participation statistics, **so that** I can track contribution and recent work.

## Scope

Scout totals, rank, recent shift history, and an authorized Family Manager's family summary.

## Preconditions

- The actor is authenticated.
- Corrected attendance and reporting relationships are available for the selected season.

## Acceptance criteria

1. **Given** a Family Manager selects a scout in a household they manage, **when** statistics load, **then** total hours, completed shifts, average hours per shift, volunteer rank, and recent shift hours are shown.
2. **Given** a Young Adult Scout opens My Stats, **when** statistics load, **then** only their linked profile's detailed statistics are shown.
3. **Given** a Family Manager views the summary, **when** the family unit is resolved, **then** combined family hours and family rank are shown without double-counting a person linked through multiple households.
4. **Given** an attendance correction exists, **when** any metric is calculated, **then** corrected hours replace superseded raw totals while the immutable events remain unchanged.

## Business rules

- Scheduled and walk-in hours count; negative durations are impossible.
- Detailed family-member data is limited to relationships the actor is authorized to manage.
- A person contributes at most once to a family-unit total.

## Dependencies

- US-002
- US-007
- US-009
- US-010
- US-037

## Out of scope

- Editing attendance
- Season-wide committee reports
- Scout Bucks dollar awards
