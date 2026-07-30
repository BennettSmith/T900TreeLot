# User stories

## Authority and decomposition

[`docs/use-cases.md`](../use-cases.md) remains the authoritative source for user workflows, permissions, and business rules. These user stories are product-slice decompositions of that source: each story packages a small, testable outcome that can be delivered through the real application boundary. A story must not invent policy or override a use case; any conflict is resolved in favor of the use case and documented before behavior changes.

Story identifiers (`US-001` through `US-059`) are stable traceability keys. They remain attached to the same product outcome even if titles, implementation boundaries, or delivery order evolve. Story numbering is organizational, not a statement that stories must be implemented strictly in numeric order.

See [hard story dependencies](dependencies.md) and the [recommended incremental roadmap](roadmap.md).

## Epics

1. [Platform Bootstrap and Identity](platform-bootstrap-identity/)
2. [Household Onboarding and Family Graph](household-onboarding-family-graph/)
3. [Young Adult Scout Access](young-adult-scout-access/)
4. [Seasonal Conduct Agreement](seasonal-conduct-agreement/)
5. [Schedule Authoring and Publication](schedule-authoring-publication/)
6. [Volunteer Scheduling](volunteer-scheduling/)
7. [Staffing Intelligence and Shift Operations](staffing-intelligence-shift-operations/)
8. [Attendance and On-Site Roster](attendance-on-site-roster/)
9. [Walk-In Coverage](walk-in-coverage/)
10. [Communications and Reminders](communications-reminders/)
11. [Hours, Leaderboards, and Scout Bucks](hours-leaderboards-scout-bucks/)
12. [Administration, Profiles, Privacy, and Season Lifecycle](administration-profiles-privacy-season-lifecycle/)

## Use-case traceability

The lists below map every defined use case to the stories that explicitly cover it. One use case can require several product slices, and one story can cover related rules from several use cases.

- **UC-0** → [US-001](platform-bootstrap-identity/us-001-bootstrap-first-administrator.md)
- **UC-1** → [US-005](household-onboarding-family-graph/us-005-invite-a-new-household.md), [US-006](household-onboarding-family-graph/us-006-redeem-household-invitation-and-create-household.md), [US-007](household-onboarding-family-graph/us-007-establish-household-members-and-explicit-adult-scout-relationships.md)
- **UC-2** → [US-002](platform-bootstrap-identity/us-002-secure-authenticated-sign-in.md)
- **UC-2A** → [US-010](young-adult-scout-access/us-010-grant-young-adult-scout-access-to-an-existing-scout-profile.md)
- **UC-2B** → [US-003](platform-bootstrap-identity/us-003-change-own-phone-number.md), [US-004](platform-bootstrap-identity/us-004-recover-another-persons-phone-access.md)
- **UC-3** → [US-020](schedule-authoring-publication/us-020-add-an-individual-published-shift.md)
- **UC-4** → [US-042](communications-reminders/us-042-publish-and-deliver-troop-announcement.md)
- **UC-4A** → [US-043](communications-reminders/us-043-view-and-mark-announcements-read.md)
- **UC-5** → [US-052](administration-profiles-privacy-season-lifecycle/us-052-deactivate-a-household.md)
- **UC-6** → [US-044](communications-reminders/us-044-send-automated-shift-reminders.md)
- **UC-7** → [US-008](household-onboarding-family-graph/us-008-add-a-co-manager.md)
- **UC-8** → [US-022](volunteer-scheduling/us-022-view-household-schedule.md)
- **UC-9, UC-10, and UC-11** → [US-023](volunteer-scheduling/us-023-manager-signs-up-household-members.md)
- **UC-12** → [US-024](volunteer-scheduling/us-024-young-adult-scout-self-schedules.md)
- **UC-13 and UC-14** → [US-025](volunteer-scheduling/us-025-cancel-a-household-owned-adult-scout-assignment.md)
- **UC-15** → [US-025](volunteer-scheduling/us-025-cancel-a-household-owned-adult-scout-assignment.md), [US-026](volunteer-scheduling/us-026-coordinate-a-shared-scout-assignment.md)
- **UC-16** → [US-024](volunteer-scheduling/us-024-young-adult-scout-self-schedules.md)
- **UC-17** → [US-023](volunteer-scheduling/us-023-manager-signs-up-household-members.md)
- **UC-18 and UC-19** → [US-023](volunteer-scheduling/us-023-manager-signs-up-household-members.md), [US-024](volunteer-scheduling/us-024-young-adult-scout-self-schedules.md)
- **UC-20** → [US-016](schedule-authoring-publication/us-016-create-shift-templates.md)
- **UC-21** → [US-017](schedule-authoring-publication/us-017-generate-draft-season-schedule.md)
- **UC-22** → [US-018](schedule-authoring-publication/us-018-review-and-adjust-draft-schedule.md)
- **UC-23** → [US-019](schedule-authoring-publication/us-019-publish-season-schedule.md)
- **UC-24** → [US-020](schedule-authoring-publication/us-020-add-an-individual-published-shift.md)
- **UC-25** → [US-028](volunteer-scheduling/us-028-sign-up-for-a-special-event.md)
- **UC-26** → [US-009](household-onboarding-family-graph/us-009-link-a-scout-across-households.md), [US-026](volunteer-scheduling/us-026-coordinate-a-shared-scout-assignment.md)
- **UC-27** → [US-034](attendance-on-site-roster/us-034-check-self-in.md)
- **UC-28** → [US-035](attendance-on-site-roster/us-035-checked-in-adult-checks-another-volunteer-in-out.md)
- **UC-29** → [US-036](attendance-on-site-roster/us-036-committee-checks-in-arriving-volunteer.md)
- **UC-30** → [US-035](attendance-on-site-roster/us-035-checked-in-adult-checks-another-volunteer-in-out.md)
- **UC-31** → [US-037](attendance-on-site-roster/us-037-review-and-correct-attendance-no-shows.md)
- **UC-32** → [US-038](attendance-on-site-roster/us-038-view-household-attendance-history.md)
- **UC-33** → [US-039](walk-in-coverage/us-039-committee-records-walk-in-covering-no-show.md)
- **UC-34** → [US-040](walk-in-coverage/us-040-checked-in-adult-adds-walk-in-scout.md)
- **UC-35** → [US-041](walk-in-coverage/us-041-prior-shift-scout-extends-as-walk-in.md)
- **UC-36** → [US-045](hours-leaderboards-scout-bucks/us-045-view-scout-hours-and-stats.md)
- **UC-37** → [US-046](hours-leaderboards-scout-bucks/us-046-view-individual-leaderboard.md)
- **UC-38** → [US-047](hours-leaderboards-scout-bucks/us-047-view-family-leaderboard.md)
- **UC-39** → [US-048](hours-leaderboards-scout-bucks/us-048-view-committee-season-statistics.md)
- **UC-40** → [US-049](hours-leaderboards-scout-bucks/us-049-review-scout-bucks-inputs-and-preview-awards.md), [US-050](hours-leaderboards-scout-bucks/us-050-finalize-and-export-scout-bucks-settlement.md), [US-051](hours-leaderboards-scout-bucks/us-051-issue-corrected-scout-bucks-revision.md)
- **UC-41** → [US-029](staffing-intelligence-shift-operations/us-029-view-weekly-staffing-levels.md)
- **UC-42** → [US-030](staffing-intelligence-shift-operations/us-030-review-staffing-alerts-dashboard.md)
- **UC-43** → [US-027](volunteer-scheduling/us-027-discover-available-shifts-in-week-view.md)
- **UC-44** → [US-021](schedule-authoring-publication/us-021-navigate-seasons-and-weeks.md)
- **UC-45** → [US-053](administration-profiles-privacy-season-lifecycle/us-053-update-profile-photo.md)
- **UC-46** → [US-054](administration-profiles-privacy-season-lifecycle/us-054-edit-display-name.md)
- **UC-47** → [US-055](administration-profiles-privacy-season-lifecycle/us-055-remove-own-login-access.md)
- **UC-48** → [US-056](administration-profiles-privacy-season-lifecycle/us-056-fulfill-personal-data-export.md)
- **UC-49** → [US-057](administration-profiles-privacy-season-lifecycle/us-057-fulfill-permanent-data-removal.md)
- **UC-50** → [US-011](seasonal-conduct-agreement/us-011-configure-seasonal-agreement-link.md)
- **UC-51** → [US-013](seasonal-conduct-agreement/us-013-confirm-agreement-for-a-participant.md)
- **UC-52 is intentionally undefined** in the authoritative use-case document and has no story mapping.
- **UC-53** → [US-015](seasonal-conduct-agreement/us-015-enforce-participation-eligibility.md), [US-023](volunteer-scheduling/us-023-manager-signs-up-household-members.md), [US-024](volunteer-scheduling/us-024-young-adult-scout-self-schedules.md), [US-034](attendance-on-site-roster/us-034-check-self-in.md), [US-036](attendance-on-site-roster/us-036-committee-checks-in-arriving-volunteer.md), [US-039](walk-in-coverage/us-039-committee-records-walk-in-covering-no-show.md), [US-040](walk-in-coverage/us-040-checked-in-adult-adds-walk-in-scout.md), [US-041](walk-in-coverage/us-041-prior-shift-scout-extends-as-walk-in.md)
- **UC-54** → [US-014](seasonal-conduct-agreement/us-014-review-confirmation-status.md)
- **UC-55** → [US-012](seasonal-conduct-agreement/us-012-open-current-agreement-from-profile.md)
- **UC-56** → [US-058](administration-profiles-privacy-season-lifecycle/us-058-create-encrypted-completed-season-archive.md), [US-059](administration-profiles-privacy-season-lifecycle/us-059-delete-an-archived-completed-season.md)
- **UC-57** → [US-032](staffing-intelligence-shift-operations/us-032-send-critical-coverage-alert.md)
- **UC-58** → [US-031](staffing-intelligence-shift-operations/us-031-enforce-actual-local-two-deep-coverage.md), [US-034](attendance-on-site-roster/us-034-check-self-in.md), [US-035](attendance-on-site-roster/us-035-checked-in-adult-checks-another-volunteer-in-out.md), [US-036](attendance-on-site-roster/us-036-committee-checks-in-arriving-volunteer.md), [US-039](walk-in-coverage/us-039-committee-records-walk-in-covering-no-show.md), [US-040](walk-in-coverage/us-040-checked-in-adult-adds-walk-in-scout.md), [US-041](walk-in-coverage/us-041-prior-shift-scout-extends-as-walk-in.md)
- **UC-59** → [US-033](staffing-intelligence-shift-operations/us-033-close-or-reopen-unsafe-shift.md)
