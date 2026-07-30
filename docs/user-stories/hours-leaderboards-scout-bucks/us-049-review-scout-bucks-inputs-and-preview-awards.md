# US-049 — Review Scout Bucks Inputs and Preview Awards

- **Epic:** Hours, Leaderboards & Scout Bucks
- **Source use cases:** [UC-40](../../use-cases.md#use-case-40-treasurer-finalizes-scout-bucks-awards)
- **Primary actor:** Treasurer

**As a** Treasurer, **I want** to review credited-hour inputs and preview exact awards, **so that** errors are found before an immutable settlement is created.

## Scope

Finalized input review, adult-hour allocation detail, distributable-pool entry, and deterministic award preview.

## Preconditions

- The season is completed.
- Attendance corrections and adult-to-scout relationships have been reviewed and finalized.

## Acceptance criteria

1. **Given** finalized inputs, **when** the report opens, **then** each scout shows own hours and shifts, allocated adult hours by adult and eligible-scout count, and total credited hours.
2. **Given** an adult has eligible scouts, **when** allocation is reviewed, **then** the adult's corrected hours are split equally at full precision among distinct eligible scouts; adults with none contribute no Scout Bucks hours.
3. **Given** a pool in dollars and cents, **when** preview runs, **then** it shows total credited hours, an informational effective rate, every proposed award, and confirmation that awards equal the pool exactly.
4. **Given** fractional cents occur, **when** awards are previewed, **then** integer cents are assigned by deterministic largest remainder with stable person-ID tie-breaking.
5. **Given** total credited hours are zero, **when** a nonzero pool is entered, **then** preview cannot be approved for finalization.

## Business rules

- Scout and adult hours are deduplicated across households and roles.
- Corrected attendance, including walk-ins, is used; superseded events remain immutable.
- No dollar estimate appears before the Treasurer enters a pool.
- Preview creates no settlement, balance, or spending ledger.

## Dependencies

- US-007
- US-009
- US-037

## Out of scope

- Editing attendance or relationships
- Persisting or exporting a settlement
- Managing Scout Bucks balances or redemptions
