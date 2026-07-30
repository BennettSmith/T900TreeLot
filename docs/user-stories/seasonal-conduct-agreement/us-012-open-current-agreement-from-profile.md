# US-012 Open Current Agreement from Profile

- **Epic:** Seasonal Conduct Agreement
- **Source use cases:** [UC-55](../../use-cases.md)
- **Primary actor:** Authenticated user

**As an** authenticated user with profile access,  
**I want** to open the current season's agreement from an authorized profile,  
**so that** I or the participant can revisit the rules of conduct.

## Scope

Show current confirmation information on an authorized profile and open the externally maintained agreement.

## Preconditions

- The user is authenticated.
- US-011 has configured the current season's agreement link.
- The user is authorized to view the relevant profile or participant view.

## Acceptance criteria

1. **Given** an authorized profile, **when** the user opens it, **then** the profile shows the person's current-season confirmation status and confirmation date when applicable.
2. **Given** the current agreement is configured, **when** the user selects "View Agreement," **then** the browser opens the public Google Doc link.
3. **Given** a Family Manager, **when** they request a profile outside a household they manage, **then** the system denies access to that profile's agreement entry point.
4. **Given** a Young Adult Scout, **when** they use this capability, **then** it is limited to their own profile; Committee and Admin use their authorized season or participant views.

## Business rules

- Managed Scouts access the agreement on a Family Manager's device.
- Opening the link does not confirm the agreement.
- The system does not proxy, copy, render, cache, or preserve the external document contents.

## Dependencies

- US-002 — authenticate the user.
- US-007 — establish household membership.
- US-008 — establish co-manager authority where profile access is delegated.
- US-009 — link a shared scout whose profile is visible from multiple households.
- US-011 — configure the current agreement link.

## Out of scope

- Editing the external agreement.
- Recording confirmation.
