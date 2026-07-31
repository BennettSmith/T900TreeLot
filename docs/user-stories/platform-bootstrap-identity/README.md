# Platform Bootstrap and Identity

## Goal

Establish the first administrator and provide secure, personal, passwordless passkey access while preserving identity continuity when credentials or the claimed account email change or must be recovered. Separately, allow an Admin to grant or revoke Admin and Committee roles on existing adult login identities.

## Source use cases

- [UC-0: Creating the First Administrator](../../use-cases.md#use-case-0-creating-the-first-administrator)
- [UC-2: Authenticated Family Member Signs In](../../use-cases.md#use-case-2-authenticated-family-member-signs-in)
- [UC-2B: Authenticated Person Manages Credentials and Account Email](../../use-cases.md#use-case-2b-authenticated-person-manages-credentials-and-account-email)
- [UC-60: Admin Manages Privileged Roles](../../use-cases.md#use-case-60-admin-manages-privileged-roles)

## Actors

- Designated first Admin
- Technical setup person or designated technical operator
- Family Manager
- Young Adult Scout
- Committee Member
- Admin
- Authorized recovery actor

## Stories

1. [US-001: bootstrap first administrator](us-001-bootstrap-first-administrator.md)
2. [US-002: secure authenticated sign-in](us-002-secure-authenticated-sign-in.md)
3. [US-003: manage own passkeys and account email](us-003-change-own-phone-number.md)
4. [US-004: recover another person's passkey access](us-004-recover-another-persons-phone-access.md)
5. [US-060: manage Admin and Committee roles](us-060-manage-admin-and-committee-roles.md)

## Story dependency view

Arrows run from each hard prerequisite to the story that depends on it. Each story's **Dependencies** section is authoritative.

```mermaid
flowchart LR
    us001["US-001"] --> us002["US-002"]
    us002 --> us003["US-003"]
    us002 --> us004["US-004"]
    us008["US-008"] --> us004
    us010["US-010"] --> us004
    us001 --> us060["US-060"]
    us002 --> us060
    us006["US-006"] --> us060
```

## Cross-epic dependencies

- US-005 and US-006 establish the invitation and household path through which the first Family Manager receives authenticated access.
- US-006 also provides a later adult Family Manager login identity that US-060 can promote to Admin or Committee; Young Adult Scout identities are not eligible.
- US-008 establishes co-manager authority used in Family Manager recovery.
- US-010 establishes Young Adult Scout access and the Family Manager recovery relationship for that access.
- US-011, US-013, and US-015 configure, record, and enforce the participation gate shown after sign-in; authentication itself occurs before confirmation.
