# US-009: link a scout across households

- **Epic:** [Household Onboarding and Family Graph](README.md)
- **Source use cases:** [UC-26](../../use-cases.md#use-case-26-divorced-parents-scout-in-two-households)
- **Primary actor:** Family Manager

**As a** Family Manager in another household, **I want** to use a scout-specific household link code, **so that** the same scout profile can belong to both households without duplication.

## Scope

Generate, share outside the system, and redeem a cryptographically random code bound to one existing scout profile to add that profile to an additional household while preserving independent management boundaries.

## Preconditions

- The scout has one existing profile in a source household.
- Both source and destination households are established.
- A Family Manager of the source household can communicate the code directly to a Family Manager of the destination household.

## Acceptance criteria

1. **Given** a Family Manager manages the source household and its scout profile, **when** they request a household link code, **then** the system generates a cryptographically random, expiring, single-use code bound only to that scout profile and purpose.
2. **Given** the source manager regenerates the scout's code, **when** the new code is created, **then** the prior code is invalidated.
3. **Given** a destination Family Manager has a valid code, **when** they redeem it for their household, **then** the existing scout profile is linked to the destination household and no duplicate person profile is created.
4. **Given** the scout is linked to both households, **when** either household views the scout's schedule, **then** all of the scout's assignments are visible in both households.
5. **Given** a household-owned assignment was created through one household, **when** a manager of the other household attempts cancellation, **then** the server denies it; each household controls only its own household-owned assignments.
6. **Given** a scout profile is linked to another household, **when** the link is established, **then** adults from either household are not automatically linked to the other household and existing explicit adult-to-scout relationships remain person-level links.
7. **Given** a person lacks a valid scout-specific code, **when** they try to locate a scout to link, **then** no scout search capability exposes profiles.

## Business rules

- A scout may belong to multiple households while retaining one profile and one schedule.
- Each scout has a separate cryptographically random link code.
- Link codes are single-use, expire after a configured short period, are purpose-bound, and are bound to one scout profile.
- Regenerating a code invalidates the previous code.
- There is no scout search; parents communicate codes directly.
- Linking a scout does not link adults between households.
- Assignments are visible across linked households, but household-owned cancellation authority remains with the originating household.
- A Young Adult Scout-created assignment is instead manageable by the scout and managers in any linked household.

## Dependencies

- US-006 establishes both households and their Family Managers.
- US-007 establishes the single scout profile and explicit adult-to-scout relationships.

## Out of scope

- Creating either household.
- Creating a second scout profile.
- Searching for scouts by name or other personal data.
- Automatically linking adults or siblings.
- Shift signup and cancellation workflows beyond preserving their ownership boundaries.
- Resolving custody disputes.
