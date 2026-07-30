# US-056 — Fulfill Personal Data Export

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-48](../../use-cases.md#use-case-48-user-requests-data-export)
- **Primary actor:** Admin or Privacy Contact

**As an** Admin or Privacy Contact, **I want** to fulfill a separately verified personal-data export, **so that** a current or former person receives a complete portable copy of data held about them.

## Scope

Verified-request intake, subject resolution, complete compilation, portable export, secure delivery, and fulfillment audit.

## Preconditions

- The requester used the published privacy channel.
- The Privacy Contact separately verified the requester and, for a managed minor, the parent/guardian's authority.

## Acceptance criteria

1. **Given** no separate identity verification exists, **when** fulfillment is attempted from an authenticated session alone, **then** export is blocked.
2. **Given** a verified request, **when** the authorized actor compiles it, **then** it includes the subject's profile and retained account email/photo, memberships, assignments, immutable attendance and corrections, walk-ins, messages, account activity, and agreement confirmations.
3. **Given** the subject is linked through multiple households or roles, **when** data is gathered, **then** all matching records are included once without exposing unrelated people's private data.
4. **Given** compilation succeeds, **when** delivery occurs, **then** JSON or CSV is securely delivered only to the verified requester and fulfillment is audited without logging sensitive contents.

## Business rules

- Current and former authenticated people may request exports; a verified parent/guardian may request one for a managed minor.
- Export does not alter source records or remove login access.
- Account emails, message bodies, and export contents are redacted from operational logs.

## Dependencies

- US-002
- US-006
- US-007
- US-009
- US-023
- US-024
- US-034
- US-035
- US-036
- US-037
- US-042

## Out of scope

- Permanent deletion
- Legal determination of requester eligibility
- Season archive creation
