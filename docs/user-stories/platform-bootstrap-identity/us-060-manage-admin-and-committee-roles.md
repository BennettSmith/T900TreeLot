# US-060: manage Admin and Committee roles

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-60](../../use-cases.md#use-case-60-admin-manages-privileged-roles)
- **Primary actor:** Admin

**As an** Admin, **I want** to grant or revoke Admin and Committee roles on an existing adult login identity, **so that** privileged authority can be delegated or reduced without creating a second person or reopening bootstrap.

## Scope

Allow a signed-in Admin to grant or revoke Admin and/or Committee roles on an existing adult login identity after a recent passkey step-up, while preserving identity continuity, auditing the change, and enforcing last-active-Admin continuity. The target must already have a login identity on file and does not need to be signed in during the change.

## Preconditions

- The first Admin has been established and bootstrap is closed.
- The acting Admin is signed in and active.
- The target is an adult person profile that already has a login identity (claimed email and at least one registered passkey), typically a Family Manager, established through an authorized invitation path.
- The target is not a Young Adult Scout identity and is not a managed scout without login access.

## Acceptance criteria

1. **Given** an active Admin and an eligible adult login identity, **when** the Admin grants the Admin role, the Committee role, or both after a recent passkey step-up, **then** the target identity receives exactly those roles without a new person profile, claimed email, or passkey set, whether or not the target is currently signed in.
2. **Given** an active Admin and a target adult login identity that holds Admin and/or Committee roles, **when** the Admin revokes one or both roles after a recent passkey step-up, **then** those privileged permissions no longer authorize privileged actions.
3. **Given** the acting Admin is the last active Admin, **when** they attempt to revoke their own Admin role or otherwise leave the system with no active Admin, **then** the change is rejected unless another Admin has already been appointed or the separately secured break-glass procedure applies.
4. **Given** the selected target is a Young Adult Scout identity or a managed scout without login access, **when** the Admin attempts to grant Admin or Committee roles, **then** the server rejects the change.
5. **Given** a Committee Member, Family Manager, Young Adult Scout, or unauthenticated person attempts to change Admin or Committee roles, **when** the server evaluates authorization, **then** no role change occurs.
6. **Given** a privileged-role change succeeds, **when** audit evidence is inspected, **then** it records the acting Admin, target identity, roles granted or revoked, and server timestamp without exposing secrets or passkey material.

## Business rules

- Only Admin may grant or revoke Admin and Committee roles.
- Eligible targets are adult person profiles that already have a login identity, typically Family Managers; the target need not be signed in.
- Young Adult Scout identities and managed scouts without login cannot receive Admin or Committee roles.
- Privileged-role changes require a recent passkey step-up by the acting Admin and are audited.
- Role changes update the existing adult login identity; they never create a replacement person profile or discard memberships, assignments, or history.
- The last active Admin cannot remove or lose the Admin role without first appointing another Admin or using the separately secured break-glass procedure.
- This story does not enroll a new identity, create a household invitation, or grant Young Adult Scout access.

## Dependencies

- US-001 establishes the first Admin and closes bootstrap.
- US-002 provides authenticated sessions and passkey step-up.
- US-006 establishes a later adult Family Manager login identity that can receive a privileged role.

## Out of scope

- Creating the first Admin through bootstrap.
- Creating new-household invitations or redeeming them.
- Granting or revoking Young Adult Scout access.
- Defining the secured break-glass recovery procedure.
- Changing claimed email or managing passkeys.
- Household membership or family-graph changes.
