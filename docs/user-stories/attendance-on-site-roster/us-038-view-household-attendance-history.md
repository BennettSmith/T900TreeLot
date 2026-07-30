# US-038 — View household attendance history

- **Epic:** Attendance & On-Site Roster
- **Source use cases:** [UC-32](../../use-cases.md#use-case-32-family-manager-views-attendance-history)
- **Primary actor:** Family Manager

As a Family Manager,
I want to view attendance history for my household,
so that I can verify family participation and credited hours.

## Scope

Provide household-level and person-level attendance summaries and detailed shift history within the manager's relationship-aware authorization boundary.

## Preconditions

- The Family Manager is authenticated and actively manages the requested household.
- Attendance records exist or the system can present an empty history.

## Acceptance criteria

1. **Given** an authorized Family Manager, **when** they open History, **then** the system shows the household's total shifts worked and total hours.
2. **Given** household members with attendance, **when** the summary is displayed, **then** each person has a shift count and hours total.
3. **Given** a person in the managed household, **when** the manager opens that person's details, **then** each worked shift shows date, scheduled time, credited hours, and who performed check-in and checkout.
4. **Given** an approved attendance adjustment, **when** history is calculated, **then** corrected hours are shown while the existence of the adjustment remains distinguishable from raw real-time events.
5. **Given** scheduled and walk-in attendance, **when** history is displayed, **then** both contribute to totals and each record's origin remains identifiable.
6. **Given** a person outside every household the actor manages, **when** the actor requests that person's attendance details, **then** access is denied without exposing private attendance.
7. **Given** a scout belongs to multiple households, **when** an authorized manager views that scout, **then** the scout has one attendance history rather than duplicated person records.

## Business rules

- Authorization is based on current household-management relationships, not submitted household or person IDs.
- Reports use approved corrected hours while preserving immutable raw events and adjustment audit history.
- Scheduled and walk-in hours both count.
- Shared scouts retain one profile and one attendance history across linked households.
- Attendance details do not grant access to unrelated household data.

## Dependencies

- US-002
- US-007
- US-037

## Out of scope

- Editing or correcting attendance
- Season-wide leaderboards or statistics
- Scout Bucks allocation or settlement
- Attendance data for unrelated households
