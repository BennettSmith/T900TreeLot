# US-053 — Update Profile Photo

- **Epic:** Administration, Profiles, Privacy & Season Lifecycle
- **Source use cases:** [UC-45](../../use-cases.md#use-case-45-user-updates-profile-photo)
- **Primary actor:** Authenticated user

**As an** authenticated user, **I want** to add, change, or remove my profile photo, **so that** my current image appears consistently in authorized troop views.

## Scope

Browser image selection, adjustment, upload, replacement/removal, storage, and authorized display.

## Preconditions

- The actor has an active identity linked to a person profile.

## Acceptance criteria

1. **Given** the actor opens profile settings, **when** no photo exists, **then** a placeholder and an add-photo action are shown.
2. **Given** the actor selects or captures a supported image, **when** they crop or adjust and confirm it, **then** the validated image is stored through the profile BlobStore and linked to their person profile.
3. **Given** an update succeeds, **when** rosters, leaderboards, or family views render that profile, **then** the new photo appears only to authorized authenticated viewers.
4. **Given** the actor removes their photo, **when** the action commits, **then** the stored image is deleted and the placeholder returns.

## Business rules

- Server authorization and validation apply equally to full-page and enhanced requests.
- Removing login access does not remove a preserved profile photo; photo removal and verified privacy deletion are separate actions.
- Image data is not exposed through public or unrelated-person access.

## Dependencies

- US-002

## Out of scope

- Display-name editing
- Login removal
- Privacy-request fulfillment
