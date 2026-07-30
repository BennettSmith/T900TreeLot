# US-051 — Issue Corrected Scout Bucks Revision

- **Epic:** Hours, Leaderboards & Scout Bucks
- **Source use cases:** [UC-40](../../use-cases.md#use-case-40-treasurer-finalizes-scout-bucks-awards)
- **Primary actor:** Treasurer or Admin

**As a** Treasurer or Admin, **I want** to issue a reasoned corrected settlement revision, **so that** an error can be fixed without rewriting financial history.

## Scope

Reason capture, renewed input and pool confirmation, immutable successor revision, and revision-aware export/history.

## Preconditions

- A finalized Scout Bucks settlement already exists.
- The actor is authorized to correct settlements.

## Acceptance criteria

1. **Given** a finalized settlement, **when** the actor starts a correction, **then** the system requires a nonempty reason and displays the prior revision for comparison.
2. **Given** corrected attendance, relationships, or pool information, **when** the actor reviews the correction, **then** the credited-hour snapshot, distributable pool, deterministic cent allocation, and exact-sum confirmation are presented again.
3. **Given** the actor confirms a valid correction, **when** it is stored, **then** a new immutable revision links to the prior revision and records reason, actor, and server timestamp without modifying either prior snapshot or awards.
4. **Given** multiple revisions exist, **when** reports or CSV exports are requested, **then** the current revision is unmistakable and every prior revision remains available in audit history.

## Business rules

- Corrections never edit or delete a finalized settlement in place.
- Every revision independently allocates integer cents exactly equal to its recorded pool.
- Full-precision credited hours and deterministic largest-remainder tie-breaking are retained.

## Dependencies

- US-007
- US-009
- US-037
- US-050

## Out of scope

- Altering raw attendance events
- Deleting obsolete revisions
- Maintaining Scout Bucks balances or spending
