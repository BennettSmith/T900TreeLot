# US-054 — Edit Display Name

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-46](../../use-cases.md#use-case-46-user-edits-display-name)
- **Primary actor:** Authenticated user

**As an** authenticated user, **I want** to edit my display name, **so that** the web app uses the name by which I should currently be known.

## Scope

Editing and validating the actor's first, last, or preferred display name and propagating it to current views.

## Preconditions

- The actor has an active identity linked to a person profile.

## Acceptance criteria

1. **Given** the actor opens profile settings, **when** they edit a supported name field and save valid content, **then** the linked person profile is updated.
2. **Given** a name change succeeds, **when** current rosters, leaderboards, family views, and profile pages render, **then** they use the updated display name.
3. **Given** a historical audit entry predates the change, **when** history is reviewed, **then** its captured name snapshot remains unchanged and attributable.
4. **Given** the actor submits another person's profile identifier, **when** authorization is evaluated, **then** the self-service command still targets only the profile linked to the actor's identity.

## Business rules

- Authentication remains tied to the verified phone identity, not the display name.
- Name changes do not replace the person profile, identity, roles, memberships, assignments, or history.
- Server-side authorization does not trust submitted profile IDs.

## Dependencies

- US-002

## Out of scope

- Phone-number change or recovery
- Family Manager editing of another managed profile
- Historical audit snapshot rewriting
