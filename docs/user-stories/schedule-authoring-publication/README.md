# Schedule Authoring and Publication

## Goal

Let authorized committee leaders define reusable shift patterns, generate and refine a private draft, publish the season once ready, and navigate or extend the published schedule.

## Source use cases

- [UC-3, UC-20, UC-21, UC-22, UC-23, UC-24, and UC-44](../../use-cases.md)

## Actors

- Committee Member
- Committee Chair
- Admin
- Any authenticated user

## Stories

- [US-016 Create shift templates](us-016-create-shift-templates.md)
- [US-017 Generate draft season schedule](us-017-generate-draft-season-schedule.md)
- [US-018 Review and adjust draft schedule](us-018-review-and-adjust-draft-schedule.md)
- [US-019 Publish season schedule](us-019-publish-season-schedule.md)
- [US-020 Add an individual published shift](us-020-add-an-individual-published-shift.md)
- [US-021 Navigate seasons and weeks](us-021-navigate-seasons-and-weeks.md)

## Story dependency view

Arrows run from each hard prerequisite to the story that depends on it. Each story's **Dependencies** section is authoritative.

```mermaid
flowchart LR
    authenticatedCommittee["US-002 Authenticated Committee authority"]
    us016["US-016 Create templates"]
    us017["US-017 Generate draft"]
    us018["US-018 Adjust draft"]
    us019["US-019 Publish schedule"]
    us020["US-020 Add published shift"]
    us021["US-021 Navigate seasons and weeks"]

    authenticatedCommittee --> us016
    us016 --> us017
    us017 --> us018
    us018 --> us019
    us019 --> us020
    us019 --> us021
```

## Cross-epic dependencies

- US-002 provides authenticated Committee, Admin, and schedule-viewer identities.
- Published schedule stories provide the schedule required by the volunteer-scheduling epic.
- The authoring flow is US-016 → US-017 → US-018 → US-019; US-020 and US-021 depend on US-019.
