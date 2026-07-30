# US-047 — View Family Leaderboard

- **Epic:** Hours, Leaderboards & Scout Bucks
- **Source use cases:** [UC-38](../../use-cases.md#use-case-38-family-manager-views-family-leaderboard)
- **Primary actor:** Family Manager

**As a** Family Manager, **I want** to compare family-unit participation, **so that** my family's combined contribution is visible without inflating shared members' hours.

## Scope

Family-unit rank, total corrected hours, shift count, own-family highlighting, and authorized member breakdown.

## Preconditions

- The actor is an active authenticated Family Manager.
- Family-unit relationships and corrected attendance are available.

## Acceptance criteria

1. **Given** reportable family units exist, **when** the actor selects the family leaderboard, **then** each row shows rank, family name, total corrected hours, and shift count.
2. **Given** the actor belongs to a ranked family unit, **when** results appear, **then** that family is clearly highlighted.
3. **Given** the actor opens their own family, **when** the breakdown is shown, **then** each contributing person's hours are visible within the actor's authorized relationships.
4. **Given** a scout or other person is linked through multiple households in one family unit, **when** totals are calculated, **then** that person's hours and shifts contribute exactly once.

## Business rules

- Family units are explicit reporting groupings; household links do not silently merge unrelated adults.
- Corrected attendance governs totals while source attendance events remain immutable.
- Committee/Admin can review family-unit composition; unrelated Family Managers cannot inspect private member details.

## Dependencies

- US-007
- US-009
- US-037

## Out of scope

- Changing family-unit membership
- Individual leaderboard behavior
- Scout Bucks adult-hour attribution
