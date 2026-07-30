# US-026 Coordinate a Shared Scout Assignment

- **Epic:** Volunteer Scheduling
- **Source use cases:** [UC-26](../../use-cases.md)
- **Primary actor:** Family Manager

**As a** Family Manager of a scout in multiple households,  
**I want** all linked households to see the scout's assignments while preserving origin ownership,  
**so that** we can avoid double booking and coordinate changes safely.

## Scope

Share one scout profile and schedule across linked households, expose assignment origin, and enforce the corresponding management boundary.

## Preconditions

- The scout's single profile is linked to multiple households.
- The viewing manager is authenticated and authorized for one linked household.
- A published shift assignment exists for the scout.

## Acceptance criteria

1. **Given** either household creates an assignment for the shared scout, **when** managers in any linked household view their schedules, **then** they all see the same assignment on the scout's one profile.
2. **Given** an assignment created through a household, **when** another linked household views it, **then** the originating household is identified and the other household cannot cancel it.
3. **Given** the scout is already assigned to a shift through any household, **when** another household attempts signup for that shift, **then** the duplicate assignment is rejected.
4. **Given** an assignment created directly by a Young Adult Scout, **when** a manager from any linked household views it, **then** that manager may manage it under the Young Adult Scout exception.

## Business rules

- A shared scout has one profile, one schedule, and assignments visible to all linked households.
- Household ownership governs cancellation of manager-created assignments.
- Young Adult Scout-created assignments have no originating household and are manageable by the scout and all linked-household managers.
- Linking a scout does not automatically link adults from the households.

## Dependencies

- US-002 — authenticate the Family Manager.
- US-006 — create the linked households.
- US-007 — establish the scout's household memberships.
- US-009 — link the scout across those households without duplicating the profile.
- US-019 — publish the shift schedule.
- US-021 — provide season and week navigation.
- US-022 — display household and shared-scout schedules.

## Out of scope

- Creating household link codes or linking the scout profile.
- Resolving custody disputes.
- Committee/Admin overrides.
