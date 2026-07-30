# US-004: recover another person's passkey access

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-2B](../../use-cases.md#use-case-2b-authenticated-person-manages-credentials-and-account-email)
- **Primary actor:** Authorized recovery actor

**As an** authorized recovery actor, **I want** to restore another person's access after they lose all passkeys, **so that** they can register a new passkey without creating a second identity or losing their records.

## Scope

Cover assisted recovery for a Young Adult Scout, Family Manager, Committee Member, or Admin when no usable passkey remains, using the recovery authority defined for that person's role and a reissued enrollment invitation.

## Preconditions

- The person has an existing authenticated identity but cannot complete a passkey assertion.
- An authorized recovery actor is available, or the separately secured break-glass path is required because no active Admin can recover an Admin.

## Acceptance criteria

1. **Given** a Young Adult Scout has lost all passkeys, **when** a Family Manager who manages that profile confirms the scout's identity and reissues a Young Adult Scout invitation, **then** the scout may register a new passkey on the same identity.
2. **Given** a Family Manager has an active co-manager, **when** that co-manager or an Admin confirms the person's identity, **then** they may revoke remaining credentials and reissue a recovery or co-manager enrollment invitation.
3. **Given** a sole Family Manager or Committee Member needs recovery, **when** an Admin completes assisted identity verification, **then** the Admin may revoke credentials and reissue an enrollment invitation.
4. **Given** an Admin needs recovery, **when** another Admin completes recovery, **then** the existing Admin identity may accept a recovery invitation and register a new passkey.
5. **Given** no active Admin can recover an Admin, **when** the designated technical operator uses the separately secured break-glass process, **then** the recovery action is recorded.
6. **Given** an authorized recovery is approved, **when** credentials are changed, **then** old passkeys and every existing session are revoked before the recovery invitation can create a replacement passkey.
7. **Given** the claimed email remains unverified, **when** recovery is performed, **then** recovery remains invitation-based rather than email-based.
8. **Given** recovery succeeds, **when** the person signs in with the new passkey, **then** the same internal identity, profile, roles, memberships, assignments, and history are preserved and the recovery is present in the audit trail.

## Business rules

- Possession of an email address alone is insufficient to take over an identity.
- Recovery authority depends on the subject: managing Family Manager for a Young Adult Scout; co-manager or Admin for a Family Manager; Admin for a sole Family Manager or Committee Member; another Admin for an Admin.
- The secured break-glass process applies only when no active Admin can perform Admin recovery.
- Prior passkeys and all sessions are revoked before re-enrollment.
- Recovery changes credentials on the existing identity and is audited.

## Dependencies

- US-002 provides identities, passkey sign-in, and revocable sessions.
- US-008 provides the active co-manager relationship used for co-manager-assisted Family Manager recovery.
- US-010 provides Young Adult Scout access and its link to a manager-controlled scout profile.

## Out of scope

- Specifying the technical controls of the break-glass procedure.
- Self-service passkey or email changes while a passkey remains available.
- Creating a replacement profile or merging duplicate identities.
- Mailbox-verified email recovery before the later verification workflow exists.
- Removing roles, memberships, assignments, or history.
