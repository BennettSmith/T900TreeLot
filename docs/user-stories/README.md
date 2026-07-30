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

The generated [requirements traceability report](../traceability.md) maps every
current use-case revision to its user-story revision, increment, delivery
status, implementation PR, and merge SHA. `traceability/manifest.yaml` is the
machine-readable source for that mapping; do not maintain a second mapping in
this README.

UC-52 remains intentionally undefined in the authoritative use-case document
and has no story mapping.
