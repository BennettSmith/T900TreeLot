# US-058 — Create Encrypted Completed-Season Archive

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-56](../../use-cases.md#use-case-56-admin-archives-and-deletes-a-completed-season)
- **Primary actor:** Admin

**As an** Admin, **I want** to create a verified encrypted archive of a completed season, **so that** its historical data can be restored after deliberate live deletion.

## Scope

Consistent export, manifest and checksums, passphrase-based age encryption, download, and successful-archive evidence.

## Preconditions

- The season is completed and inactive, no shift is in progress, and end-of-season reports are finalized.
- The Admin manually initiates archival.

## Acceptance criteria

1. **Given** an eligible season, **when** archive creation starts, **then** the Admin sees included categories and the system captures a consistent snapshot of all season-owned operational, agreement, message, delivery, audit, and reporting data plus minimum interpretive person/household/family snapshots.
2. **Given** snapshot generation succeeds, **when** the package is built, **then** the ZIP contains a format version, manifest, versioned JSON or CSV files, record counts, generation time, and SHA-256 checksums.
3. **Given** the Admin enters and confirms a passphrase, **when** encryption completes, **then** a passphrase-based age file named `season-{name}.zip.age` downloads and the application never stores or logs the passphrase.
4. **Given** archive verification succeeds, **when** completion is displayed, **then** the checksum is shown and retained as archive evidence while the live season remains visible and unchanged.
5. **Given** live season data changes afterward, **when** archive status is reviewed, **then** the prior archive is not silently updated and a new archive is required before deletion.

## Business rules

- No schedule, retention timer, or background policy may initiate archival.
- The archive contains the agreement URL but not external Google Doc contents.
- Losing the separately stored passphrase makes restoration impossible.

## Dependencies

- US-050

## Out of scope

- Season deletion (US-059)
- Automatic archival or archive custody
- Archive restoration
