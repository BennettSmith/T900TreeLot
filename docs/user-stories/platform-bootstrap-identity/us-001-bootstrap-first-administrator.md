# US-001: bootstrap first administrator

- **Epic:** [Platform Bootstrap and Identity](README.md)
- **Source use cases:** [UC-0](../../use-cases.md#use-case-0-creating-the-first-administrator)
- **Primary actor:** Designated first Admin

**As a** designated first administrator, **I want** to establish my administrator access using the configured bootstrap enrollment token and a passkey, **so that** I can begin authorized onboarding and delegate administrative roles.

## Scope

Create exactly one initial Admin through the deployment-configured bootstrap enrollment token, claimed email, and passkey registration, and permanently close the bootstrap path after success.

## Preconditions

- The web application is deployed and configured.
- No Admin exists.
- A one-time bootstrap enrollment token is configured before go-live.

## Acceptance criteria

1. **Given** no Admin exists and a valid bootstrap enrollment token is presented, **when** the designated person claims an email, registers a passkey, and completes their profile, **then** the system creates the first Admin identity and a secure session.
2. **Given** the first Admin has been established, **when** any person attempts to use the bootstrap mechanism, **then** the system rejects the attempt because bootstrap is permanently disabled.
3. **Given** an invalid or already-consumed bootstrap token is submitted, **when** the system responds, **then** the response does not reveal unrelated account details and remains subject to rate limiting.
4. **Given** the first Admin is active, **when** they use their granted permissions, **then** they can create new-household invitation links or QR codes and grant Admin or Committee roles.

## Business rules

- Exactly one first Admin may be created through the configured bootstrap enrollment token.
- Bootstrap is permanently disabled after that Admin is established.
- Every later authenticated identity requires an authorized invitation or role assignment.
- The claimed email is stored as the account identifier and is not mailbox-verified during bootstrap.
- Only Admin may grant or revoke Admin and Committee roles.
- The last active Admin cannot lose Admin access without another active Admin or the separately secured break-glass procedure.

## Dependencies

- No prior user story is required; deployment and bootstrap-token configuration are operational prerequisites.

## Out of scope

- Creating later Admins or Committee Members.
- Creating households or family-member profiles.
- Defining the secured break-glass recovery procedure.
- Configuring or confirming a seasonal agreement.
- Mailbox verification of the claimed email.
