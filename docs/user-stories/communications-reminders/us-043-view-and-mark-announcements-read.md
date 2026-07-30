# US-043 — View and Mark Announcements Read

- **Epic:** Communications & Reminders
- **Source use cases:** [UC-4A](../../use-cases.md#use-case-4a-authenticated-user-views-announcements)
- **Primary actor:** Authenticated user

**As an** authenticated user, **I want** to read current and historical announcements and manage my read state, **so that** I can distinguish new information from announcements I have handled.

## Scope

Announcement history, detail display, unread count, and identity-private read/unread actions.

## Preconditions

- The actor has an active authenticated identity.
- At least one canonical announcement may exist.

## Acceptance criteria

1. **Given** announcements exist, **when** the actor opens the list, **then** announcements appear newest first with title, priority, author, publication time, and the actor's read indicator.
2. **Given** unread announcements exist, **when** navigation is rendered, **then** it shows the current identity's accurate unread count.
3. **Given** the actor opens an announcement, **when** its complete body is displayed, **then** it is marked read for that identity with a server timestamp.
4. **Given** the actor changes read state, **when** they mark one announcement unread or all read, **then** only their private read records and count change.

## Business rules

- A new announcement starts unread for every active authenticated recipient except its author.
- History published before an identity gained access remains visible but does not increase that identity's initial unread count.
- SMS delivery, SMS-link clicks, Groups.io delivery, and delivery status do not themselves change web read state.
- Read state is private per authenticated identity and is not delivery confirmation.
- Authorization is identical for ordinary and HTMX requests.

## Dependencies

- US-002
- US-042

## Out of scope

- Publishing, delivery retries, or delivery-detail administration
- Message replies, reactions, or shared read receipts
