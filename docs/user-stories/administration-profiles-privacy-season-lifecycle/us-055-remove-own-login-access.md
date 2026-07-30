# US-055 — Remove Own Login Access

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-47](../../use-cases.md#use-case-47-authenticated-person-removes-own-login)
- **Primary actor:** Authenticated person

**As an** authenticated person, **I want** to remove my own login access, **so that** I can stop authenticating without erasing my profile or volunteer history.

## Scope

Continuity checks, consequence review, passkey step-up re-authentication, identity-access removal, session revocation, and audit.

## Preconditions

- The actor is authenticated.
- Another active manager exists for an active household if the actor is its only manager.
- Another active Admin exists if the actor is an Admin.

## Acceptance criteria

1. **Given** removal would leave an active household managerless or remove the last active Admin, **when** requested, **then** it is blocked with the required continuity options.
2. **Given** continuity is satisfied, **when** the actor reviews all affected roles and re-authenticates with a passkey step-up, **then** explicit confirmation removes their authenticated access.
3. **Given** removal commits, **when** effects are inspected, **then** all sessions are revoked, passkeys and claimed email authentication are removed, the person is removed from future in-app inbox delivery and any future verified-email notification preference, authenticated roles are revoked as confirmed, and an audit record is appended.
4. **Given** a Young Adult Scout removes access, **when** removal succeeds, **then** the existing scout profile returns to manager-controlled status and its schedule remains manageable by Family Managers.
5. **Given** any actor removes access, **when** retained data is inspected, **then** profile, display name, photo, assignments, attendance, user ID, and history remain preserved.

## Business rules

- Login removal is not personal-data deletion, export, or anonymization.
- Multi-role consequences must be listed before removing the entire identity.
- The actor is signed out after the atomic removal.

## Dependencies

- US-002
- US-006
- US-009

## Out of scope

- Permanent data removal (US-057)
- Household deactivation or transfer
- Removing another person's login
