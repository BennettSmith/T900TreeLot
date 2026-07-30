# US-007: establish household members and explicit adult-scout relationships

- **Epic:** [Household Onboarding and Family Graph](README.md)
- **Source use cases:** [UC-1](../../use-cases.md#use-case-1-new-family-joins)
- **Primary actor:** Family Manager

**As a** Family Manager, **I want** to add the people in my household and explicitly record each adult's relationship to its scouts, **so that** participation and family reporting use the correct people and relationships.

## Scope

Create one person profile for each adult or scout in the household and record parent, step-parent, or guardian links explicitly between the managed adult and scout profiles.

## Preconditions

- The household has been created and is active.
- The actor is an active Family Manager for that household.

## Acceptance criteria

1. **Given** an active manager of the household, **when** they add an adult or scout family member, **then** the system creates or associates a person profile with that household under the manager's authority.
2. **Given** a scout is added during household onboarding, **when** the profile is established, **then** the scout is a managed profile and does not require an authenticated login.
3. **Given** adult and scout profiles managed by the actor, **when** the manager records a parent, step-parent, or guardian relationship, **then** the system stores an explicit link between those two person profiles rather than inferring it from household membership.
4. **Given** a relationship is created, corrected, or removed, **when** the change is saved, **then** it is audited and provisional Scout Bucks attribution is recalculated.
5. **Given** relationship information exists, **when** an unrelated volunteer views a shift roster, **then** those relationship details are not exposed.
6. **Given** household members have been established, **when** onboarding continues, **then** each person retains a separate current-season agreement status and only individually Confirmed people become eligible for participation.

## Business rules

- A person profile owns that person's assignments, attendance, hours, and history.
- One person has one profile even across multiple households or authenticated roles.
- Household membership does not itself prove a parent, step-parent, or guardian relationship.
- Family Managers record relationships for profiles they manage; Admin may correct disputed or duplicate relationships.
- Explicit person-to-person relationships continue across linked households.
- Relationship creation, correction, and removal are audited and recalculate provisional Scout Bucks attribution.
- Managed scouts need no authenticated access.

## Dependencies

- US-006 establishes the active household and first Family Manager.

## Out of scope

- Inferring relationships from household membership.
- Granting authenticated access to a scout.
- Linking a scout to another household.
- Correcting disputed or duplicate relationships as an Admin workflow.
- Defining Scout Bucks settlement calculations.
- Confirming the seasonal agreement.
