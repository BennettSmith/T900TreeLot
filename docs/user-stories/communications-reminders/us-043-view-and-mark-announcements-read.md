# US-043 — View and Mark Inbox Messages Read

- **Epic:** Communications & Reminders
- **Source use cases:** [UC-4A](../../use-cases.md#use-case-4a-authenticated-user-views-in-app-inbox)
- **Primary actor:** Authenticated user

**As an** authenticated user, **I want** to read current and historical inbox messages and manage my read state, **so that** I can distinguish new information from messages I have handled.

## Scope

Inbox history, detail display, unread count, and identity-private read/unread actions.

## Preconditions

- The actor has an active authenticated identity.
- At least one inbox message may exist.

## Acceptance criteria

1. **Given** inbox messages exist, **when** the actor opens the list, **then** messages appear newest first with title or subject, type, priority when applicable, author or system source, publication time, and the actor's read indicator.
2. **Given** unread messages exist, **when** navigation is rendered, **then** it shows the current identity's accurate unread count.
3. **Given** the actor opens a message, **when** its complete body is displayed, **then** it is marked read for that identity with a server timestamp.
4. **Given** the actor changes read state, **when** they mark one message unread or all read, **then** only their private read records and count change.

## Business rules

- A new inbox message starts unread for every active authenticated recipient except its human author.
- History published before an identity gained access remains visible when retained for that audience but does not increase that identity's initial unread count.
- Groups.io delivery, future opted-in email delivery, and external delivery status do not themselves change in-app read state.
- Direct and family-scoped messages appear only in intended recipients' inboxes.
- Read state is private per authenticated identity and is not delivery confirmation.
- Authorization is identical for ordinary and HTMX requests.

## Dependencies

- US-002
- US-042

## Out of scope

- Publishing, delivery retries, or delivery-detail administration
- Message replies, reactions, or shared read receipts
