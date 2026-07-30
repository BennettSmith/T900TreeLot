# US-052 — Deactivate a Household

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-5](../../use-cases.md#use-case-5-family-becomes-inactive)
- **Primary actor:** Admin

**As an** Admin, **I want** to deactivate one household safely, **so that** its management and future scheduling stop without erasing people or unrelated access.

## Scope

Warnings, confirmation, household status, authority suspension, origin-aware assignment cancellation, and audit.

## Preconditions

- The actor is an active authenticated Admin.
- The target household is active.

## Acceptance criteria

1. **Given** an active household, **when** deactivation begins, **then** the Admin sees warnings and affected future assignments before confirming.
2. **Given** confirmation, **when** deactivation commits, **then** the household becomes inactive, its Family Manager authority is suspended, and its future household-owned assignments are cancelled once with slots freed.
3. **Given** a person also has another active household or Committee/Admin role, **when** deactivation completes, **then** that other authorized access, shared profile, and supported assignments remain active.
4. **Given** a scout has no active linked household afterward, **when** effects are applied, **then** Young Adult Scout access and future self-created assignments are suspended.
5. **Given** deactivation succeeds, **when** history is viewed, **then** profiles, past assignments, attendance, and an audit fact remain; later reactivation does not recreate cancelled assignments.

## Business rules

- Deactivation targets a household, not every linked identity or person.
- Assignment cancellation follows recorded origin ownership.
- The entire state change, audit append, and resulting outbox work are atomic.

## Dependencies

- US-002
- US-006
- US-007
- US-009
- US-023
- US-024

## Out of scope

- Permanent personal-data deletion
- Automatic assignment recreation on reactivation
- Removing unrelated roles or household access
