# US-057 — Fulfill Permanent Data Removal

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-49](../../use-cases.md#use-case-49-user-requests-permanent-data-removal)
- **Primary actor:** Admin or Privacy Contact

**As an** Admin or Privacy Contact, **I want** to fulfill a separately verified permanent-removal request, **so that** personal data is deleted or anonymized to the fullest permitted extent.

## Scope

Verification gate, irreversible-impact review, coordinated deletion/anonymization, report impact, and non-identifying fulfillment evidence.

## Preconditions

- The request came through the published privacy channel.
- The requester and any parent/guardian authority were separately verified.
- The requester has been warned that removal is irreversible and can change reports and Scout Bucks inputs.

## Acceptance criteria

1. **Given** only an authenticated session or login-removal request, **when** permanent removal is attempted, **then** it is blocked until separate privacy verification is recorded.
2. **Given** a verified, confirmed request, **when** fulfillment commits, **then** profile data and photo are deleted, identifying historical fields are removed or anonymized, household links are removed, and assignments, attendance, messages, and agreement confirmations are deleted or anonymized as applicable.
3. **Given** the subject contributed hours or adult allocations, **when** removal completes, **then** affected reports and provisional Scout Bucks calculations no longer include that personal contribution.
4. **Given** fulfillment succeeds, **when** evidence is retained, **then** it proves a request was fulfilled without retaining identifying information about the deleted person.
5. **Given** any step in the coordinated state change fails, **when** the transaction rolls back, **then** no partial deletion is presented as fulfilled.

## Business rules

- Permanent removal is distinct from US-055 login removal, which preserves profile and history.
- Agreement confirmations have no special retention exemption.
- Timing after season-end reporting is preferred but not stated as a mandatory denial rule.

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

- Providing legal advice or defining statutory retention exceptions
- Personal-data export
- Deleting an entire season
