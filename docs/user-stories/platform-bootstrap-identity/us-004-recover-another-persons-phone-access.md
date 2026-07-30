# US-004: recover another person's phone access

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-2B](../../use-cases.md#use-case-2b-authenticated-person-changes-phone-number)
- **Primary actor:** Authorized recovery actor

**As an** authorized recovery actor, **I want** to restore another person's access after they lose their old phone number, **so that** they can use a newly verified number without creating a second identity or losing their records.

## Scope

Cover assisted recovery for a Young Adult Scout, Family Manager, Committee Member, or Admin when the old number is unavailable, using the recovery authority defined for that person's role.

## Preconditions

- The person has an existing authenticated identity but cannot use its current phone number.
- An authorized recovery actor is available, or the separately secured break-glass path is required because no active Admin can recover an Admin.
- The proposed new number can receive SMS verification.

## Acceptance criteria

1. **Given** a Young Adult Scout has lost the old number, **when** a Family Manager who manages that profile confirms the scout's identity and reissues access, **then** recovery may proceed for the scout's new number.
2. **Given** a Family Manager has an active co-manager, **when** that co-manager or an Admin confirms the person's identity, **then** they may revoke and reissue the manager's access.
3. **Given** a sole Family Manager or Committee Member needs recovery, **when** an Admin completes assisted identity verification, **then** the Admin may revoke and reissue access.
4. **Given** an Admin needs recovery, **when** another Admin completes recovery, **then** the existing Admin identity may be linked to the new verified number.
5. **Given** no active Admin can recover an Admin, **when** the designated technical operator uses the separately secured break-glass process, **then** the recovery action is recorded.
6. **Given** an authorized recovery is approved, **when** the credential is changed, **then** the old phone link and every existing session are revoked before the unique new number is linked.
7. **Given** the new number fails system-wide uniqueness or SMS verification, **when** recovery is attempted, **then** it is rejected and the number is not linked.
8. **Given** recovery succeeds, **when** the person signs in with the new number, **then** the same internal identity, profile, roles, memberships, assignments, and history are preserved and the recovery is present in the audit trail.

## Business rules

- Possession of a new phone number alone is insufficient to take over an identity.
- Recovery authority depends on the subject: managing Family Manager for a Young Adult Scout; co-manager or Admin for a Family Manager; Admin for a sole Family Manager or Committee Member; another Admin for an Admin.
- The secured break-glass process applies only when no active Admin can perform Admin recovery.
- The prior phone link and all sessions are revoked before reassignment.
- The new normalized number must be unique and verified.
- Recovery changes a credential on the existing identity and is audited.

## Dependencies

- US-002 provides identities, verification, and revocable sessions.
- US-008 provides the active co-manager relationship used for co-manager-assisted Family Manager recovery.
- US-010 provides Young Adult Scout access and its link to a manager-controlled scout profile.

## Out of scope

- Specifying the technical controls of the break-glass procedure.
- Self-service change while the old number remains available.
- Creating a replacement profile or merging duplicate identities.
- Silently taking a number away from another active identity.
- Removing roles, memberships, assignments, or history.
