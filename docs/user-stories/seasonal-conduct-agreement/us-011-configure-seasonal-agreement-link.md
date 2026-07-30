# US-011 Configure Seasonal Agreement Link

- **Epic:** Seasonal Conduct Agreement
- **Source use cases:** [UC-50](../../use-cases.md)
- **Primary actor:** Admin

**As an** Admin,  
**I want** to configure the agreement title and public Google Doc link for a season,  
**so that** participants review the troop-approved rules that are currently in effect.

## Scope

Configure or replace the single external agreement link associated with a season.

## Preconditions

- The Admin is authenticated and authorized to configure the season.
- Troop leadership has published the intended rules as a publicly readable Google Doc.

## Acceptance criteria

1. **Given** a season and a public HTTPS Google Doc link, **when** the Admin enters a display title and link, reviews the destination, and confirms, **then** the season stores that title and link as its current agreement.
2. **Given** an invalid or unsupported link, **when** the Admin submits it, **then** the system rejects it without changing the current agreement.
3. **Given** existing confirmations for the season, **when** the Admin attempts to replace the link, **then** the system warns that all confirmations will reset and requires explicit confirmation.
4. **Given** the Admin confirms a replacement, **when** the change is committed, **then** the new link becomes current and every participant's confirmation for that season is reset atomically.

## Business rules

- A season has one current display title and one public HTTPS Google Doc link.
- The system does not copy, upload, proxy, render, cache, version, or retain the document contents.
- The system cannot detect document edits made without changing the URL; Admin must replace the link when applicable rules change.

## Dependencies

- US-002 — authenticate the Admin.

## Out of scope

- Creating, editing, publishing, storing, or versioning the Google Doc.
- Confirming the agreement for any participant.
