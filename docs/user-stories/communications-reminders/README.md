# Communications & Reminders

## Goal

Keep active troop participants informed through a canonical web record, reliable independent delivery channels, private read state, and correctly routed shift reminders.

## Source use cases

- [UC-4 — Committee Sends Announcement](../../use-cases.md#use-case-4-committee-sends-announcement)
- [UC-4A — Authenticated User Views Announcements](../../use-cases.md#use-case-4a-authenticated-user-views-announcements)
- [UC-6 — Shift Reminders (Automated)](../../use-cases.md#use-case-6-shift-reminders-automated)

## Actors

- Committee Member and Admin
- Family Manager
- Young Adult Scout
- Automated reminder worker

## Stories

- [US-042 — Publish and deliver troop announcement](us-042-publish-and-deliver-troop-announcement.md)
- [US-043 — View and mark announcements read](us-043-view-and-mark-announcements-read.md)
- [US-044 — Send automated shift reminders](us-044-send-automated-shift-reminders.md)

## Dependencies

- US-002 provides authenticated identities.
- US-006 establishes active household managers, and US-010 establishes Young Adult Scout access used for recipient selection.
- US-019 provides published shifts.
- US-023 and US-024 provide household-owned and Young Adult Scout-created assignments used by reminder routing.
- US-043 requires US-042's canonical announcements.

## Story dependency view

Arrows run from each hard prerequisite to the story that depends on it. Each story's **Dependencies** section is authoritative.

```mermaid
flowchart LR
    authenticatedIdentity["Authenticated identity (US-002)"] --> publishAnnouncement["US-042 Publish announcement"]
    authenticatedIdentity --> viewAnnouncements["US-043 View announcements"]
    authenticatedIdentity --> sendReminders["US-044 Send shift reminders"]
    announcementRecipients["Active manager and Young Adult Scout access (US-006, US-010)"] --> publishAnnouncement
    reminderRelationships["Household links and Young Adult Scout access (US-009, US-010)"] --> sendReminders
    reminderAssignments["Published shifts and assignments (US-019, US-023, US-024)"] --> sendReminders
    publishAnnouncement --> viewAnnouncements
```
