# US-046 — View Individual Leaderboard

- **Epic:** Hours, Leaderboards & Scout Bucks
- **Source use cases:** [UC-37](../../use-cases.md#use-case-37-family-manager-views-individual-leaderboard)
- **Primary actor:** Family Manager

**As a** Family Manager, **I want** to view volunteers ranked by worked hours, **so that** troop contributions can be recognized.

## Scope

An individual leaderboard showing rank, display name, corrected hours, shift count, and the current actor's position.

## Preconditions

- The actor is an active authenticated Family Manager.
- Corrected season attendance is available.

## Acceptance criteria

1. **Given** reportable participation exists, **when** the actor opens the individual leaderboard, **then** volunteers are ranked by corrected worked hours with rank, name, hours, and shift count.
2. **Given** more volunteers exist than the initial display, **when** the leaderboard loads, **then** at least the top ten are available and additional entries can be reached.
3. **Given** the actor has reportable hours, **when** their row is displayed or located, **then** it is highlighted with a clear “YOU” indicator.
4. **Given** corrected attendance changes before settlement finalization, **when** the view is refreshed, **then** ranks and totals reflect the approved correction.

## Business rules

- Reports use approved corrected hours while retaining immutable real-time events.
- Each person profile appears once even if it has multiple roles or household links.
- Phone numbers and private relationship details are never displayed.

## Dependencies

- US-007
- US-009
- US-037

## Out of scope

- Family-unit ranking
- Editing profiles or attendance
- Scout Bucks allocation and settlement
