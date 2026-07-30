# US-050 — Finalize and Export Scout Bucks Settlement

- **Epic:** Hours, Leaderboards & Scout Bucks
- **Source use cases:** [UC-40](../../use-cases.md#use-case-40-treasurer-finalizes-scout-bucks-awards)
- **Primary actor:** Treasurer

**As a** Treasurer, **I want** to finalize and export the reviewed Scout Bucks awards, **so that** the troop's separate account process receives an exact, auditable settlement.

## Scope

Confirmation, immutable revision creation, current-revision display, and CSV export.

## Preconditions

- The actor is an authorized Treasurer or Admin.
- A valid US-049 preview matches finalized inputs and the entered distributable pool.

## Acceptance criteria

1. **Given** a valid preview, **when** the actor confirms finalization, **then** one transaction stores an immutable settlement revision containing the pool, full-precision credited-hour snapshot, integer-cent awards, rounding allocation, actor, and server timestamp.
2. **Given** finalization succeeds, **when** totals are validated, **then** all awards sum exactly to the entered pool and the same input cannot create an accidental duplicate settlement.
3. **Given** a finalized settlement, **when** the actor exports it, **then** the CSV identifies season and revision and contains values that reconcile exactly to the stored snapshot.
4. **Given** the settlement is viewed later, **when** source attendance or relationships differ, **then** the finalized revision remains unchanged.

## Business rules

- Integer cents and deterministic largest remainder are authoritative; credited-hour ratios retain full precision.
- Finalized settlements are immutable and revisioned.
- Export does not create an ongoing balance, redemption, transfer, or adjustment ledger.
- Audit facts are appended in the state-changing transaction.

## Dependencies

- US-007
- US-009
- US-037
- US-049

## Out of scope

- Correcting a finalized settlement (US-051)
- Scout Bucks spending or account management
- Changing finalized attendance inputs
