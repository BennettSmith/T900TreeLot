# US-001: bootstrap first administrator

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-0](../../use-cases.md#use-case-0-creating-the-first-administrator)
- **Primary actor:** Designated first Admin

**As a** designated first administrator, **I want** to establish my administrator access using the configured bootstrap phone number, **so that** I can begin authorized onboarding and delegate administrative roles.

## Scope

Create exactly one initial Admin through the deployment-configured bootstrap phone number, verified by the standard SMS one-time-code or magic-link flow, and permanently close the bootstrap path after success.

## Preconditions

- The web application and SMS authentication provider are deployed and configured.
- No Admin exists.
- The designated first Admin's mobile number is configured as the bootstrap number before go-live.

## Acceptance criteria

1. **Given** no Admin exists and the entered number matches the configured bootstrap number, **when** the designated person successfully verifies a short-lived, single-use SMS code or magic link, **then** the system creates the first Admin identity and allows the person to complete their profile.
2. **Given** the first Admin has been established, **when** any person attempts to use the bootstrap mechanism, **then** the system rejects the attempt because bootstrap is permanently disabled.
3. **Given** a number that is not eligible for bootstrap is submitted, **when** the system responds to the authentication request, **then** the response does not reveal whether that number is registered and remains subject to rate limiting.
4. **Given** the first Admin is active, **when** they use their granted permissions, **then** they can create new-household invitations and grant Admin or Committee roles.

## Business rules

- Exactly one first Admin may be created through the configured bootstrap phone number.
- Bootstrap is permanently disabled after that Admin is established.
- Every later authenticated identity requires an authorized invitation or role assignment.
- Verification codes and links are short-lived, single-use, and rate-limited.
- Only Admin may grant or revoke Admin and Committee roles.
- The last active Admin cannot lose Admin access without another active Admin or the separately secured break-glass procedure.

## Dependencies

- No prior user story is required; deployment and SMS-provider configuration are operational prerequisites.

## Out of scope

- Creating later Admins or Committee Members.
- Creating households or family-member profiles.
- Defining the secured break-glass recovery procedure.
- Configuring or confirming a seasonal agreement.
