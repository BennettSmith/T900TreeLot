# Volunteer Scheduling

## Goal

Help households and Young Adult Scouts discover published shifts, create valid assignments, coordinate shared scouts, and cancel assignments only within their authority.

## Source use cases

- [UC-8 through UC-19, UC-25, UC-26, UC-43, and UC-53](../../use-cases.md)

## Actors

- Family Manager
- Young Adult Scout
- Managed Scout
- Adult family member
- Committee Member
- Admin
- System

## Stories

- [US-022 View household schedule](us-022-view-household-schedule.md)
- [US-023 Manager signs up household members](us-023-manager-signs-up-household-members.md)
- [US-024 Young Adult Scout self-schedules](us-024-young-adult-scout-self-schedules.md)
- [US-025 Cancel a household-owned adult/scout assignment](us-025-cancel-a-household-owned-adult-scout-assignment.md)
- [US-026 Coordinate a shared scout assignment](us-026-coordinate-a-shared-scout-assignment.md)
- [US-027 Discover available shifts in week view](us-027-discover-available-shifts-in-week-view.md)
- [US-028 Sign up for a special event](us-028-sign-up-for-a-special-event.md)

## Story dependency view

Arrows run from each hard prerequisite to the story that depends on it. Each story's **Dependencies** section is authoritative.

```mermaid
flowchart LR
    householdAuthority["US-002/006/007/009/010 Identity and household authority"]
    agreementEligibility["US-015 Agreement eligibility"]
    publishedNavigation["US-019/021 Published schedule navigation"]
    projectedStaffing["US-029 Projected staffing"]
    us022["US-022 View household schedule"]
    us023["US-023 Manager signup"]
    us024["US-024 Scout self-scheduling"]
    us025["US-025 Cancel household assignment"]
    us026["US-026 Coordinate shared scout"]
    us027["US-027 Discover available shifts"]
    us028["US-028 Special-event signup"]

    householdAuthority --> us022
    householdAuthority --> us023
    householdAuthority --> us024
    householdAuthority --> us025
    householdAuthority --> us026
    householdAuthority --> us027
    householdAuthority --> us028
    agreementEligibility --> us023
    agreementEligibility --> us024
    agreementEligibility --> us028
    publishedNavigation --> us022
    publishedNavigation --> us023
    publishedNavigation --> us024
    publishedNavigation --> us025
    publishedNavigation --> us026
    publishedNavigation --> us027
    publishedNavigation --> us028
    projectedStaffing --> us027
    us022 --> us025
    us022 --> us026
    us023 --> us028
```

## Cross-epic dependencies

- US-002 provides authenticated Family Manager and Young Adult Scout identities.
- US-006 and US-008 establish manager authority; US-007 and US-009 provide household membership and shared-scout relationships.
- US-015 provides person-level agreement eligibility.
- US-019 and US-021 provide the published schedule and its season/week navigation.
