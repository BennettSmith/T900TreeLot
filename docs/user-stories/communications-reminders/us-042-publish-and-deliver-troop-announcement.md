# US-042 — Publish and Deliver Troop Announcement

- **Epic:** Communications & Reminders
- **Source use cases:** [UC-4](../../use-cases.md#use-case-4-committee-sends-announcement)
- **Primary actor:** Committee Member

**As a** Committee Member, **I want** to publish one troop announcement through the in-app inbox and configured troop-wide channels, **so that** active recipients receive consistent information without one channel's failure hiding it elsewhere.

## Scope

Compose, review, publish, deliver, track, and selectively retry a troop announcement.

## Preconditions

- The actor is authenticated and authorized to send troop announcements.
- Recipient identities and optional Groups.io configuration can be resolved.

## Acceptance criteria

1. **Given** a title, body, and priority, **when** the actor reviews the announcement, **then** the system shows active Family Manager and Young Adult Scout inbox recipient counts and the Groups.io destination only when enabled.
2. **Given** the actor confirms publication, **when** the command succeeds, **then** one canonical announcement records its content, author, priority, and server publication time and is immediately available in each recipient's in-app inbox.
3. **Given** publication succeeds, **when** delivery is queued, **then** every active Family Manager and Young Adult Scout receives an inbox message and enabled Groups.io receives a separate troop-wide post attempt.
4. **Given** an external channel fails, **when** statuses are viewed or retried, **then** each result remains independently visible and only failed external deliveries are retried idempotently without duplicating successful ones.

## Business rules

- In-app inbox publication and each optional external channel have independent delivery status.
- External provider calls occur outside the publishing transaction through an idempotent transactional outbox.
- Groups.io is disabled by default; its failure cannot roll back in-app publication, and inbox success does not depend on Groups.io.
- Priority does not alter the required recipient set, and announcement delivery is not suppressed by shift-reminder preferences.
- Message bodies and provider secrets are not exposed in logs or recipient views of other people's delivery details.
- SMS and phone-number delivery are not used.

## Dependencies

- US-002
- US-006
- US-010

## Out of scope

- Personal read/unread state (US-043)
- Shift reminders and targeted staffing reminders
- Recipient replies or discussion threads
- Verified-email notification delivery
