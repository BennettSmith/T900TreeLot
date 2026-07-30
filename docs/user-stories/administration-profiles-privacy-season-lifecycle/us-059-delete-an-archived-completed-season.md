# US-059 — Delete an Archived Completed Season

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-56](../../use-cases.md#use-case-56-admin-archives-and-deletes-a-completed-season)
- **Primary actor:** Admin

**As an** Admin, **I want** to deliberately delete a successfully archived completed season, **so that** season-owned live data is removed while shared cross-season records remain.

## Scope

Eligibility checks, destructive-impact review, SMS re-authentication, explicit confirmations, atomic season deletion, and minimal receipt.

## Preconditions

- A successful current archive and checksum exist from US-058.
- The season is completed and inactive, no shift is in progress, and live data has not changed since archival.

## Acceptance criteria

1. **Given** no current successful archive exists or the season is active, draft, or in progress, **when** deletion is requested, **then** the action remains disabled.
2. **Given** an eligible archive, **when** the Admin starts deletion, **then** exact categories and counts are shown with the warning that restoration requires both the external archive and passphrase.
3. **Given** the Admin re-authenticates by SMS, confirms separate archive/passphrase storage, and types the season name, **when** deletion is confirmed, **then** all exclusively season-owned data is removed atomically.
4. **Given** deletion succeeds, **when** shared data is inspected, **then** identities, person profiles, households, family units, reusable templates, and roles remain.
5. **Given** deletion succeeds, **when** evidence is inspected, **then** only a minimal non-personal receipt remains with actor, deletion time, former season ID, archive checksum, and deleted counts; the season is absent from live views and search.
6. **Given** the deletion transaction fails, **when** rollback completes, **then** the complete season remains intact.

## Business rules

- Deletion is always a separately initiated Admin action; archival never triggers it.
- No age, schedule, retention timer, or background process may delete a season.
- Restoration is a separately controlled technical operation and is never automatic.

## Dependencies

- US-058

## Out of scope

- Creating or storing the archive
- Shared profile or identity deletion
- Archive restoration
