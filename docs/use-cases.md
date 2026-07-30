# Troop 900 Tree Lot Shift Scheduler
## Use Cases Document

**Version:** 3.13  
**Date:** July 2026  
**Purpose:** This document describes what the Troop 900 Tree Lot Shift Scheduler does from a user's perspective. It covers all functional use cases organized by category.

The document version identifies an edition of the complete requirements set.
Per-use-case revisions, delivery status, and implementation evidence are
tracked separately in the generated
[`docs/traceability.md`](traceability.md). Follow the
[traceability process](traceability-process.md) for semantic requirement
changes; editorial changes do not increment per-use-case revisions.

---

## Table of Contents

1. [Overview](#overview)
2. [System Bootstrap](#system-bootstrap)
3. [Family Onboarding](#family-onboarding)
4. [Seasonal Conduct Agreement](#seasonal-conduct-agreement)
5. [Shift Template Management & Schedule Creation](#shift-template-management--schedule-creation)
6. [Committee Operations](#committee-operations)
7. [Shift Scheduling](#shift-scheduling)
8. [Multi-Household Support](#multi-household-support)
9. [Attendance Tracking](#attendance-tracking)
10. [Walk-In Coverage](#walk-in-coverage)
11. [Hours Tracking & Leaderboards](#hours-tracking--leaderboards)
12. [End-of-Season Reports](#end-of-season-reports)
13. [Schedule Management & Staffing Views](#schedule-management--staffing-views)
14. [Profile Management](#profile-management)
15. [Privacy & Data Requests](#privacy--data-requests)
16. [Permission Summary](#permission-summary)
17. [Business Rules](#business-rules)

---

## Overview

The Troop 900 Tree Lot Shift Scheduler is a responsive web application designed to manage volunteer scheduling for Boy Scout Troop 900's annual Christmas tree lot fundraiser. The system enables family managers to sign up parents and scouts for shifts, track attendance, view hours worked, and coordinate schedules across the 4-6 week fundraising season.

### Key Capabilities

- Passwordless authentication using passkeys (WebAuthn), with a claimed email address as the account identifier
- Closed enrollment through authorized invitation links or QR codes (for example at troop meetings or on a printed enrollment form), not open self-registration
- Family accounts managed by one or more parents/guardians, without requiring scouts to have login credentials
- Optional Young Adult Scout access for older scouts ready to manage their own shift participation, limited to their own schedule and shift actions
- Multi-household family structures supporting divorced parents, joint custody, and blended families
- Role-based access for Admins, Committee Members, Family Managers, and Young Adult Scouts
- Season-specific confirmation that every participating adult and scout has read and agrees to the linked rules of conduct
- Shift template management for efficient schedule creation
- Critical staffing alerts, shift closure workflows, and enforcement of the troop's approved local two-deep coverage rule
- Real-time shift scheduling with permission boundaries
- Web-based attendance tracking replacing paper sign-in sheets
- Walk-in coverage for handling no-shows
- Individual and family hours tracking with leaderboards
- Scout Bucks credited-hour tracking, end-of-season dollar finalization, and award export
- Per-user in-app message inbox with private read/unread state for announcements, reminders, and operational notices
- Optional Groups.io posting for troop-wide announcements when that deployment integration is enabled

### Delivery and Technology Constraints

- All user and administrative functionality is delivered through a web browser; there are no native iOS or Android applications
- The backend is implemented in Go and renders HTML on the server
- HTMX is used for progressive enhancement and partial-page updates
- Browser JavaScript is required for passkey registration and assertion and may be required for other browser APIs such as profile-photo capture
- The interface is responsive and fully usable on phone, tablet, and desktop form factors
- Heavy client-side frameworks such as Angular and React are not used
- Authentication does not use SMS one-time codes, magic links, or social identity providers
- Operational notifications do not use SMS or phone-number delivery
- Every notification for an authenticated recipient is recorded in that user's in-app inbox with private read/unread state
- Troop-wide announcements may additionally post to Groups.io when that optional deployment integration is enabled
- Direct or family-scoped messages remain in-app only until a later verified-email notification channel exists
- Claimed account email remains unverified at enrollment; a future mailbox-verification workflow may allow opted-in delivery to verified email addresses without replacing the in-app inbox

### User Base

Approximately 40-50 family members (scouts and parents) from 15-20 families/households. Family managers, administrators, committee members, and optionally Young Adult Scouts have authenticated access.

---

## System Bootstrap

### Use Case 0: Creating the First Administrator

**Actors:** Troop Leadership (designated first admin), Technical Setup Person

**Description:** Before the system can be used, the very first administrator login must be established. This is a one-time setup that occurs when the system is initially deployed. Because all later accounts require authorized invitations and authentication uses passkeys, a special bootstrap process creates that first admin.

**Preconditions:**
- The Go web application has been deployed and configured
- No administrator accounts exist in the system yet
- A one-time bootstrap enrollment token has been configured as a production secret

**What Happens:**
1. During initial system deployment, a one-time bootstrap enrollment token is configured for the designated first admin
2. The designated person opens the web app with that bootstrap token (link or equivalent enrollment entry)
3. The person claims an email address as their account identifier, completes their profile, and registers a passkey
4. The system creates the first Admin identity, links the passkey, and establishes a secure session
5. The bootstrap mechanism is disabled—no additional accounts can be created this way
6. The admin can now generate invitation links or QR codes for households and assign committee roles to other users

**Security Considerations:**
- The bootstrap enrollment token must be configured before the system goes live and is single-use
- Only ONE account can be created through bootstrap
- Once the first admin exists, all subsequent users must use an authorized invitation
- The claimed email is stored as the account identifier but is not mailbox-verified during bootstrap
- Passkey registration requires a supported browser with JavaScript and WebAuthn
- The first admin can promote other users to admin or committee roles as needed

**Outcome:** 
- The system has its first administrator who can begin onboarding families
- The bootstrap path is closed; all future accounts require authorized invitations
- The admin can delegate by creating additional admins or committee members

---

## Family Onboarding

### Use Case 1: New Family Joins

**Actors:** Admin, Primary Family Manager

**Description:** A new family joins the system using an authorized household invitation provided by the Admin, typically as a link or QR code handed out at a troop meeting or on a printed enrollment form.

**What Happens:**
1. Admin generates a unique, expiring new-household invitation and provides its link or QR code to the primary parent out of band
2. The parent opens the invitation in a supported browser
3. The parent claims an email address as their account identifier, completes their profile, and registers a passkey
4. The parent creates the household account and becomes its first Family Manager
5. The parent adds family members and records each adult's parent, step-parent, or guardian relationships to scouts
6. Scouts are created as managed family-member profiles and do not need login credentials; an eligible older scout can later be granted Young Adult Scout access
7. The family manager enters restricted agreement-first onboarding and facilitates current-season confirmation for each family member
8. Scheduling unlocks individually as each person's agreement status becomes Confirmed

**Outcome:** The family account is established with the primary parent as a Family Manager. The family can review announcements and confirm the agreement, but only family members whose current-season agreement status is Confirmed can participate in scheduling.

---

### Use Case 2: Authenticated Family Member Signs In

**Actors:** Family Manager, Young Adult Scout

**Description:** A family manager or Young Adult Scout accesses their permitted family-account functions with a passkey, without a password or third-party social login.

**What Happens:**
1. The person opens the web app on a supported browser
2. They sign in with a passkey; when the browser needs an account hint, they may enter their claimed email address first
3. The system verifies the WebAuthn assertion against a passkey registered to an active authenticated identity
4. The system creates a secure session and applies permissions for their role
5. A Family Manager opens the family dashboard; a Young Adult Scout opens their personal schedule
6. Unless the person's own current-season agreement status is Confirmed, the web app prominently directs them to the Agreement Center and disables their participation actions

**Security and Recovery:**
- Discoverable passkeys are preferred so repeat sign-in does not require typing an email
- Sessions use secure, HTTP-only cookies and can be revoked by the user, a family manager with authority over the profile, or an administrator
- Losing all registered passkeys requires assisted recovery through Use Case 2B; email is not a sign-in factor while unverified
- Authentication responses are rate-limited and do not disclose whether an email belongs to an account

**Outcome:** The person gains passwordless access without using a Google or Apple account. Most scouts remain managed profiles and never need to authenticate.

---

### Use Case 2A: Older Scout Becomes a Young Adult Scout

**Actors:** Family Manager or Admin, Older Scout

**Description:** A Family Manager or Admin grants limited authenticated access to an older scout who is ready to manage their own shift participation.

**Terminology:** "Young Adult Scout" is an application access level, not a statement that the scout has reached the legal age of majority.

**What Happens:**
1. An authorized Family Manager or Admin opens the scout's profile and selects "Grant Young Adult Access"
2. The system creates a unique, expiring Young Adult Scout invitation bound to that existing scout profile
3. The authorized user provides the invitation link or QR code to the scout out of band
4. The scout opens the invitation, claims an email address as their account identifier, registers a passkey, and accepts access
5. The existing scout profile is linked to the new authenticated identity; no duplicate family member is created
6. The scout can:
   - View their own upcoming and completed shifts
   - Sign themselves up for an available scout slot
   - Cancel their own signup
   - Check themselves in and out
   - View their own hours and statistics
7. The scout cannot:
   - View private details or schedules for other family members
   - Sign up, cancel, or check in another person
   - Add or edit family members
   - Invite managers or change family settings
8. Family managers retain visibility and management authority over the scout's assignments
9. Shift signup and attendance remain disabled until the scout confirms the current-season agreement

**How the Scout Is Associated with the Family:**
- The authorized Family Manager or Admin starts from an existing scout profile that is already a member of the family
- The invitation contains a short-lived reference to that specific scout profile and family; it is not a general family join code
- Accepting the invitation creates an authenticated identity for the scout, binds their claimed email and passkey, and links that identity to the existing scout profile
- The scout's family membership continues to come from the profile's family/household relationships, not from the email address
- For a scout in multiple linked households, the same authenticated identity and scout profile are used across all households
- A claimed email can belong to only one authenticated person in the entire system
- The invitation is rejected if the claimed email is already linked to a Family Manager, Young Adult Scout, Committee Member, or Admin
- An invitation cannot silently create or select a different profile; conflicts require administrator resolution

**Outcome:** The older scout gains age-appropriate self-service access while remaining part of the family account and subject to family-manager oversight.

---

### Use Case 2B: Authenticated Person Manages Credentials and Account Email

**Actors:** Family Manager, Young Adult Scout, Committee Member, Admin

**Description:** An authenticated person manages passkeys and their claimed account email without creating a new identity or losing family relationships, roles, assignments, or history. Email mailbox verification is deferred until notifications or email-based recovery require it.

**Manage Passkeys (at least one passkey still available):**
1. The person opens account security settings
2. The system requires a recent passkey assertion (step-up)
3. The person may register an additional passkey on the current device or remove a passkey they still control
4. The identity must retain at least one passkey unless access is being removed through Use Case 47
5. Existing sessions may remain; newly registered passkeys can be used on subsequent sign-ins

**Change Account Email (self-service):**
1. The person opens account security settings and selects "Change Email"
2. The system requires a recent passkey assertion
3. The person enters the new email address
4. The system normalizes the new address and confirms that it is not linked to another active identity
5. The system atomically replaces the claimed email on the existing authenticated identity and marks it unverified
6. All existing browser sessions are revoked, and the person signs in again with a passkey
7. Until a later verification workflow succeeds, the new email cannot be used for notifications or email-based recovery

**Assisted Recovery (no remaining passkey):**
1. The person contacts an authorized recovery actor without attempting to create a second account
2. For a Young Adult Scout, a Family Manager may revoke remaining credentials and reissue a Young Adult Scout invitation for passkey re-enrollment
3. For a Family Manager with an active co-manager, the co-manager or an Admin may revoke credentials and reissue a co-manager or recovery enrollment invitation after confirming the person's identity
4. A sole Family Manager or Committee Member requires Admin-assisted identity verification; an Admin requires recovery by another Admin
5. If no active Admin can perform recovery, the designated technical operator uses a separately secured break-glass recovery procedure and records the action
6. Old passkeys and sessions are revoked before the recovery invitation can create a replacement passkey on the same identity
7. While the account email remains unverified, recovery is invitation-based rather than email-based

**Failure and Security Rules:**
- If the new email is already linked to another person, the change is rejected and the old email remains unchanged
- Possession of an email address alone is not sufficient to take over an existing identity
- Claiming or changing an email does not prove mailbox ownership
- The existing internal user ID is preserved; credential and email changes are not account migrations
- Recovery and credential-management actions are recorded in the audit trail

**Outcome:** The person continues to sign in with passkeys on the same identity while retaining the same profile, family memberships, permissions, assignments, and historical records.

---

### Use Case 7: Family Manager Adds Co-Manager

**Actors:** Existing Family Manager, Co-Parent/Guardian

**Description:** A second parent or guardian is invited to share family management responsibilities for the same household through their own passkey-authenticated login.

**What Happens:**
1. The existing manager adds the co-parent/guardian as an adult family member, or selects an existing adult profile in the household
2. The manager creates a unique, expiring co-manager invitation and optionally records the invitee's claimed email
3. The manager provides the invitation link or QR code out of band (for example in person, by showing the QR code, or by copying the link)
4. The co-manager opens the invitation in a supported browser, claims an email address if needed, registers a passkey, and accepts access to the household
5. If the invitee already has an authenticated identity, they prove it with a passkey and the household manager role is linked to that existing identity rather than creating a second person
6. The co-manager gains family management permissions through their own login

**Outcome:** One or both parents can independently access the same family account, manage family members, and select parents or scouts when filling shifts.

---

## Seasonal Conduct Agreement

### Use Case 50: Admin Sets the Seasonal Agreement Link

**Actors:** Admin, Troop Leadership

**Description:** Admin associates a season with the public Google Doc containing the troop-approved tree-lot rules of conduct.

**What Happens:**
1. Troop leadership creates, maintains, and publishes the rules of conduct as a publicly readable Google Doc outside the scheduling system
2. Admin opens the season configuration and enters:
   - A display title
   - The public HTTPS Google Doc link
3. The system validates the link format and lets the Admin open it for review
4. Admin confirms that this is the agreement link in effect for the season
5. The system stores the link but does not copy, upload, render, or retain the document contents
6. If the Admin replaces a link after people have confirmed:
   - The system warns that every confirmation for the season will be reset
   - The Admin must explicitly confirm the replacement
   - All participants must open the new link and confirm again

**External-Document Limitation:**
- The system cannot detect edits made to a Google Doc while its URL remains unchanged
- Troop leadership and Admin are responsible for using the intended published document
- If the applicable rules change, Admin must replace the configured link so prior confirmations are reset

**Outcome:** The season has one public agreement link in effect, and no agreement document is stored by the scheduling system.

---

### Use Case 51: Family Confirms the Seasonal Agreement

**Actors:** Family Manager, Adult Family Member, Managed Scout, Young Adult Scout

**Description:** Each parent, adult volunteer, and scout confirms that they have read and agree to the rules linked for the current season before participating.

**What Happens:**
1. The Family Manager opens the Agreement Center during restricted onboarding
2. The web app shows each family member as Confirmed or Not Confirmed for the active season
3. A participant opens the public Google Doc link
4. After reading it, the participant selects the checkbox "I have read and agree to the tree-lot rules of conduct" and submits the confirmation
5. An authenticated adult or Young Adult Scout confirms through their own session
6. A Managed Scout or adult without login access may confirm on a Family Manager's device; the system records both the selected participant and the authenticated manager who facilitated the confirmation
7. The system stores the person's boolean confirmation, server timestamp, acting identity, season, and current agreement-link identifier
8. Confirming one person does not confirm any other family member
9. Each confirmed person becomes eligible for scheduling and tree-lot participation

**Outcome:** Agreement confirmation is tracked separately for every participating parent, adult, and scout.

---

### Use Case 53: Agreement Confirmation Blocks Participation

**Actors:** Family Manager, Young Adult Scout, Committee Member, Admin, System

**Description:** The system prevents a person who has not confirmed the current season's agreement from signing up, being assigned, checking in, or being added as a walk-in.

**What Happens:**
1. An actor attempts a participation action for a person
2. The server checks that person's confirmation for the shift's season and current agreement link
3. If not confirmed:
   - The action is rejected
   - No shift capacity or attendance record is changed
   - The web app explains that the person must read and confirm the current agreement
   - Family Managers receive a link to the Agreement Center
4. Committee and Admin cannot override the requirement or mark a person confirmed without that person's explicit checkbox submission

**Outcome:** No person works the tree lot without confirming the current season's linked rules of conduct.

---

### Use Case 54: Admin Reviews Agreement Confirmation

**Actors:** Admin, Committee Member

**Description:** Authorized leaders review participation readiness.

**What Happens:**
1. Admin or Committee opens the season agreement-status view
2. The view shows each person as Confirmed or Not Confirmed
3. Committee can see confirmation status but not private account or session metadata
4. Admin can additionally see the confirmation time and acting identity when a Family Manager facilitated another person's confirmation
5. The view can be filtered by household, person role, and confirmation status

**Outcome:** Leaders can identify people who still need to read and confirm the rules before participating.

---

### Use Case 55: User Opens the Agreement Link from a Profile

**Actors:** Family Manager, Young Adult Scout, Committee Member, Admin

**Description:** An authenticated user opens the public rules-of-conduct document associated with the current season.

**What Happens:**
1. The user opens an authorized person profile
2. The profile shows the person's current-season confirmation status and confirmation date when applicable
3. The user selects "View Agreement"
4. The browser opens the season's public Google Doc link
5. The scheduling system does not proxy, copy, render, cache, or preserve the document contents

**Authorization:**
- A Family Manager can access the link from profiles in households they manage
- A Young Adult Scout can access it from their own profile
- Committee and Admin can access it from the season and participant views
- A Managed Scout accesses it on a Family Manager's device

**Outcome:** Participants can revisit the current rules at the externally maintained public link.

---

## Shift Template Management & Schedule Creation

### Use Case 20: Committee Creates Shift Templates

**Actors:** Committee Member

**Description:** Committee creates reusable shift templates that define standard shift patterns for different days and occasions.

**What Happens:**
1. Committee member accesses template management
2. Reviews existing templates from prior seasons (if available)
3. Creates new templates or updates existing ones with:
   - Template name and type (weekday, weekend, special event)
   - Which days of the week the template applies to
   - Shift times for the day
   - Number of scouts and parents required per shift
   - Minimum total people required to operate the lot during the shift, which cannot be fewer than two
4. Templates can be deactivated but remain available for historical reference

**Template Types:**
- **Weekday Standard:** Monday-Thursday pattern with after-school and evening shifts
- **Friday Extended:** Friday pattern potentially with later hours
- **Weekend Full Day:** Saturday-Sunday pattern with multiple shifts throughout the day
- **Special Events:** Lot Setup Day, Tree Delivery Day, Christmas Eve, etc.

**Outcome:** Templates are saved for reuse across seasons. Changes to templates do not affect previously generated schedules—only new schedules created with the template.

---

### Use Case 21: Committee Generates Season Schedule (Bulk Creation)

**Actors:** Committee Chair

**Description:** Committee generates the entire season's shifts in one operation using templates.

**What Happens:**
1. Committee member starts a new season schedule
2. Configures the season with:
   - Season name (e.g., "2024 Tree Lot")
   - Start and end dates (e.g., Nov 29 - Dec 24)
   - Default location
3. Selects which templates to apply:
   - Regular templates automatically apply based on day of week
   - Special event templates are assigned to specific dates
4. Adds exceptions (closed days like Thanksgiving)
5. Reviews the preview showing total shifts and volunteer slots needed
6. Confirms generation

**Outcome:** 
- All shifts are created in "draft" mode (typically 70+ shifts in seconds)
- NO notifications are sent to users
- Committee can review and adjust before publishing
- Special event shifts (Lot Setup, Tree Delivery) are created with higher volunteer requirements

---

### Use Case 22: Committee Reviews and Adjusts Draft Schedule

**Actors:** Committee Member

**Description:** Committee reviews the generated schedule and makes adjustments before publishing.

**What Happens:**
1. Committee member browses the draft schedule by date
2. Reviews shifts for accuracy:
   - Check volunteer requirements
   - Verify minimum operating headcount
   - Verify timing
   - Review special event configurations
3. Makes adjustments as needed:
   - Modify volunteer requirements for specific shifts (e.g., reduce for days with known conflicts)
   - Add extra shifts not covered by templates
   - Remove or adjust shifts as needed
4. Changes remain in draft mode

**Outcome:** Schedule is refined and ready for publication. No notifications sent during this phase.

---

### Use Case 23: Committee Publishes Schedule

**Actors:** Committee Chair

**Description:** Committee publishes the entire season schedule, making it visible to all authenticated users.

**What Happens:**
1. Committee reviews final summary (total shifts, volunteer slots, special events)
2. Confirms publication with options to:
   - Place one summary notice in every active Family Manager and Young Adult Scout in-app inbox
   - Highlight special events in the notice
   - Post the same summary to Groups.io when that optional integration is enabled
3. System publishes all shifts simultaneously
4. ONE in-app inbox message is recorded for each active Family Manager and Young Adult Scout with:
   - Total shift count and date range
   - Highlights of special events (e.g., "Lot Setup Nov 27, Tree Delivery Dec 3 & 10")
5. If Groups.io is enabled and the Committee chose that channel, the system posts the same summary to the configured troop group
6. All shifts become visible and available for signup

**Outcome:** 
- Family Managers and Young Adult Scouts receive a single comprehensive in-app notification (not individual notifications per shift)
- Signup period officially begins
- Special event shifts are prominently marked in the web app

---

### Use Case 24: Committee Adds Individual Shift After Publishing

**Actors:** Committee Member

**Description:** Committee adds a single shift after the season has been published.

**What Happens:**
1. Unexpected need arises (weather changes, special event, etc.)
2. Committee member creates a single new shift
3. Shift is published immediately (not draft)
4. An in-app inbox message is recorded for eligible Family Managers and Young Adult Scouts
5. Because the recipients are a targeted eligible set rather than the whole troop, Groups.io is not used for this notice

**Outcome:** New shift is added with an in-app notification for eligible recipients. This is suitable for unplanned additions during the season.

---

### Use Case 25: Family Manager Signs Up for Special Event Shift

**Actors:** Family Manager

**Description:** A family manager discovers and signs up eligible family members for a high-priority special event shift.

**What Happens:**
1. Family manager sees an in-app inbox notice about schedule publication highlighting special events
2. Opens the web app in a browser and browses shifts
3. Special event shifts display distinctive visual indicators:
   - Star icon
   - "ALL HANDS NEEDED" badge
   - Higher volunteer requirements
   - Special notes (e.g., "Wear work clothes and gloves. Lunch provided.")
4. Parent reviews which family members are available
5. Signs up self and/or scouts for the special event

**Outcome:** Family is registered for high-priority shifts. Committee can monitor fill rates for critical days.

---

## Committee Operations

### Use Case 3: Committee Creates Shifts (Individual)

**Actors:** Committee Member

**Description:** Committee member creates individual shifts outside of bulk generation.

**What Happens:**
1. Committee member accesses shift management
2. Creates a new shift with:
   - Date and time
   - Required number of scouts and parents
   - Minimum operating headcount, which cannot be fewer than two
   - Location
   - Notes or special instructions
3. Chooses whether to publish immediately or save as draft
4. If published, an in-app inbox message is recorded for eligible Family Managers and Young Adult Scouts
5. Targeted eligible-recipient notices remain in-app only; Groups.io is not used

**Outcome:** New shift is available in the system. Eligible Family Managers and Young Adult Scouts are notified in-app if the shift is published.

---

### Use Case 4: Committee Sends Announcement

**Actors:** Committee Member

**Description:** Committee publishes a troop announcement into every active Family Manager and Young Adult Scout in-app inbox. If the optional Groups.io deployment integration is enabled, the same announcement is also posted there. Announcements are not sent by SMS.

**What Happens:**
1. Committee member composes a message with:
   - Title
   - Body text
   - Priority level
2. The committee member reviews the delivery summary showing:
   - Number of active Family Manager inbox recipients
   - Number of active Young Adult Scout inbox recipients
   - The configured troop Groups.io destination, only when that integration is enabled
3. The committee member sends the announcement
4. The system stores one canonical announcement and places a copy in each recipient's in-app inbox
5. Family Managers and Young Adult Scouts can open the announcement from the Inbox view with its title, body, priority, author, and publication time
6. If Groups.io is enabled, the system posts the same title and message to the configured troop group
7. The system records delivery status separately for in-app publication, each enabled optional channel, and any future opted-in email channel
8. If one external channel fails, the announcement remains visible in successful channels and the committee member can retry only the failed deliveries without duplicating the others

**Outcome:** Family Managers and Young Adult Scouts can read the announcement in their in-app inbox. When configured, Groups.io receives it as an additional best-effort troop-wide channel. The web app retains the canonical copy, per-user read state, and delivery status.

---

### Use Case 4A: Authenticated User Views In-App Inbox

**Actors:** Family Manager, Young Adult Scout, Committee Member, Admin

**Description:** An authenticated user views their personal in-app inbox of troop announcements, reminders, staffing notices, and other operational messages, and can distinguish unread messages from those they have already read.

**What Happens:**
1. User opens the Inbox view
2. System lists that user's messages in reverse chronological order with:
   - Title or subject
   - Message type (for example announcement, reminder, staffing notice, closure notice)
   - Priority when applicable
   - Author or system source and publication time
   - Unread/read indicator for the current user
3. Navigation shows the current user's unread inbox count
4. User opens a message to read its complete body
5. When the complete message is displayed, the system records it as read for that authenticated identity with a server timestamp
6. User may explicitly mark a message unread or mark all inbox messages read
7. Read/unread changes made by one Family Manager do not affect a co-manager, Young Adult Scout, or any other user

**Read-State Rules:**
- Each inbox message begins unread for every active authenticated recipient except its human author
- Messages published before a person received authenticated access remain available as history when retained for that audience, but do not increase that person's initial unread count
- Delivery through an enabled Groups.io integration, or a future opted-in email send, does not by itself mark the in-app message read
- Opening the linked message while authenticated marks it read when the complete body is displayed
- Read state is private to the authenticated user and is not included in other recipients' views
- Delivery status and read state are separate: delivered does not mean read
- Direct or family-scoped messages appear only in the intended recipients' inboxes and are never posted to Groups.io

**Outcome:** Each authenticated user has a personal in-app inbox with a synchronized unread count and private read history.

---

### Use Case 5: Family Becomes Inactive

**Actors:** Admin

**Description:** Admin deactivates a household account that is leaving the troop or no longer participating.

**What Happens:**
1. Admin locates the family in the management interface
2. Initiates deactivation, seeing any warnings about upcoming assignments
3. Confirms deactivation
4. System processes:
   - Household account is marked inactive
   - Family Manager authority for that household is suspended
   - Future assignments owned by that household are cancelled and their slots are freed
   - Person profiles and historical records are preserved
   - Shared scouts and their assignments through other active households remain active
   - If a scout has no remaining active household, Young Adult Scout access and future self-created assignments are suspended
   - An authenticated identity remains active if it has another active household or Committee/Admin role

**Outcome:** 
- Family Managers cannot schedule through the inactive household
- Users with another active household or privileged role retain only that other authorized access
- Cancelled household-owned shift slots are available to others
- The household can be reactivated if it returns

---

### Use Case 6: Shift Reminders (Automated)

**Actors:** System (automated)

**Description:** System automatically places Family Manager and Young Adult Scout reminders for upcoming assignments into each recipient's in-app inbox. Shift reminders are direct/family-scoped and are not posted to Groups.io.

**What Happens:**
1. Automated process runs hourly
2. Finds shifts starting in approximately 24 hours
3. For each upcoming shift:
   - Retrieves all confirmed assignments
   - For a household-owned assignment, finds active Family Managers in the originating household
   - For a Young Adult Scout-created assignment, finds active Family Managers in every linked household
   - If the volunteer is a Young Adult Scout, includes that scout as a direct recipient
   - Checks each recipient's shift-reminder preference
   - Records an in-app inbox reminder for recipients who have reminders enabled
4. Reminder includes:
   - Assigned family member's name
   - Shift date and time
   - Link to view shift details in the web app

**Outcome:** Volunteers are reminded in-app of their upcoming shifts, reducing no-shows.

---

## Shift Scheduling

### Use Case 8: Family Manager Views Family Schedule

**Actors:** Family Manager

**Description:** A family manager views the combined schedule for all parents, guardians, and scouts in the family account.

**What Happens:**
1. The manager opens the responsive web app and navigates to the family schedule
2. The web app shows all active family members
3. The manager views all upcoming family shifts, including:
   - Which family member is assigned
   - Whether the assignment fills a parent or scout slot
   - Date, time, and location
4. The manager can filter the schedule by family member
5. The manager can open an assignment to view or cancel it, subject to household rules

**Outcome:** The family has one coordinated schedule. Family managers can see and manage assignments for both managed scouts and Young Adult Scouts.

---

### Use Case 9: Family Manager Signs Up an Adult for a Shift

**Actors:** Family Manager

**Description:** A family manager fills an available parent/adult slot with themselves or another adult in the family.

**What Happens:**
1. The manager browses available shifts
2. They select a shift to view its time, location, staffing needs, current signups, and notes
3. They select "Sign Up"
4. The web app requires them to choose who will fill the shift from the family's eligible adults
5. They select themselves, a co-manager, or another eligible adult and confirm
6. The system validates:
   - The acting manager and family account are active
   - The selected adult is active and belongs to the family
   - The selected adult's current-season agreement status is Confirmed
   - The shift has an open parent/adult slot
   - The selected adult is not already assigned to the shift
7. The assignment records both the selected volunteer and the manager who created it

**Outcome:** The selected adult—not merely the logged-in manager—is registered for the shift, and the family calendar is updated.

---

### Use Case 10: Family Manager Signs Up a Scout for a Shift

**Actors:** Family Manager, Scout

**Description:** A family manager fills an available scout slot for either a managed scout or a Young Adult Scout.

**What Happens:**
1. The manager browses available shifts
2. They select a shift and choose "Sign Up"
3. The web app requires them to choose who will fill the shift
4. It shows eligible family members grouped by role:
   - Adults eligible for parent/adult slots
   - Scouts eligible for scout slots
   - Ineligible or already-assigned members with an explanation
5. The manager selects the scout and confirms the signup
6. The system validates the manager's family-management permission, the scout's membership, a Confirmed current-season agreement status, and available scout capacity
7. The assignment records the scout as the volunteer and the manager as its creator

**Outcome:** The selected scout is registered for the shift. Any authorized family manager can manage the assignment, even when the scout has Young Adult Scout access, subject to multi-household rules.

---

### Use Case 11: Signup Prevented for Wrong Slot Type

**Actors:** Family Manager

**Description:** A family manager attempts to select a family member who is not eligible for the remaining slot type.

**What Happens:**
1. The manager opens signup for a shift
2. The family-member selector shows which parent/adult and scout slots remain
3. Members who cannot fill an available slot are disabled with a clear explanation
4. The server repeats the eligibility validation when the form is submitted

**Outcome:** No invalid assignment is created, and the manager can select another eligible family member.

---

### Use Case 12: Young Adult Scout Signs Self Up

**Actors:** Young Adult Scout

**Description:** An older scout with limited authenticated access signs themselves up for an available scout slot.

**What Happens:**
1. The scout signs in with their passkey
2. They browse available shifts and open the details for a shift
3. They select "Sign Up"
4. The secure session identifies the authenticated identity and its linked scout profile
5. The web app does not show a family-member selector, and the server forces the assignment target to that linked scout profile
6. The server validates that the profile has active Young Adult Scout access, belongs to at least one active family, has a Confirmed current-season agreement status, and that a scout slot is available
7. The assignment records the linked scout profile as the volunteer and the authenticated identity as the acting user
8. The assignment appears on the scout's personal schedule and each linked household's family-manager schedule

**Outcome:** The scout can manage their own participation, while family managers retain visibility and authority to manage the assignment.

---

### Use Case 13: Family Manager Cancels Adult Assignment

**Actors:** Family Manager

**Description:** A family manager cancels an adult family member's shift assignment.

**What Happens:**
1. The manager views the family schedule
2. Selects the adult's assignment that needs to be cancelled
3. Selects "Cancel Signup"
4. Sees warning that the shift will need another volunteer
5. Confirms cancellation
6. Assignment is deleted
7. Shift slot becomes available

**Outcome:** The selected adult's assignment is removed. Committee may be notified to find a replacement.

---

### Use Case 14: Family Manager Cancels Scout Assignment

**Actors:** Family Manager, Scout

**Description:** A family manager cancels a managed or Young Adult Scout's shift assignment.

**What Happens:**
1. The manager views the family schedule
2. Opens the scout's assignment
3. System confirms that the assignment is owned by the manager's household or was created by a Young Adult Scout linked to that household
4. Manager selects "Cancel Assignment"
5. Confirms the cancellation
6. Assignment is deleted

**Outcome:** The scout's assignment is cancelled and the slot becomes available.

---

### Use Case 15: Manager Cannot Cancel Another Household's Assignment

**Actors:** Family Manager, Scout in Multiple Households

**Description:** A manager attempts to cancel a scout assignment created through another household.

**What Happens:**
1. The manager views the shared scout's schedule
2. The assignment identifies the household that created it
3. No cancellation control is offered to the other household
4. The server rejects any direct cancellation attempt without the required household permission
5. The manager coordinates with the creating household or a committee member

**Outcome:** Cross-household visibility prevents double booking while each household retains control over assignments it created.

**Young Adult Scout Exception:** An assignment created directly by a Young Adult Scout is manageable by that scout and by family managers in any of the scout's linked households.

---

### Use Case 16: Young Adult Scout Cancels Own Assignment

**Actors:** Young Adult Scout

**Description:** A Young Adult Scout cancels one of their own shift assignments, including an assignment originally created by a family manager.

**What Happens:**
1. The scout opens their personal schedule
2. They select the assignment and choose "Cancel Signup"
3. The web app warns that the shift will need another volunteer
4. The scout confirms the cancellation
5. The assignment is removed and the audit trail identifies the scout as the acting user
6. Family managers can see the cancellation in family activity

**Outcome:** The scout can manage their own schedule. Family managers retain the ability to cancel the scout's assignments as well.

---

### Use Case 17: Signup Prevented - Shift Full

**Actors:** Any User

**Description:** A user attempts to sign up for a shift that has no available slots.

**What Happens:**
1. User views a shift that shows as "Full"
2. Sees complete roster of who's assigned
3. "Sign Up" button is disabled or not visible
4. Message indicates the shift is at capacity

**Outcome:** No assignment created. User can browse other available shifts.

---

### Use Case 18: Signup Prevented - Already Signed Up

**Actors:** Any User

**Description:** A user attempts to sign up for a shift they're already on.

**What Happens:**
1. User views a shift they're already assigned to
2. Instead of "Sign Up" button, sees "You're signed up!"
3. Option to "Cancel Signup" is available
4. No duplicate signup is possible

**Outcome:** System prevents duplicate assignments.

---

### Use Case 19: Signup Prevented - Inactive Household

**Actors:** Family Manager or Young Adult Scout linked to a deactivated household

**Description:** A user attempts to schedule through a deactivated household account.

**What Happens:**
1. User opens the web app
2. System detects that the selected household is inactive
3. User sees "Household Inactive" status
4. Scheduling controls for that household are unavailable
5. If the person has another active household or Committee/Admin role, those unrelated permissions remain available
6. User may sign out or contact Admin about reactivation

**Outcome:** The inactive household cannot create scheduling actions, without disabling valid access through another household or role.

---

## Multi-Household Support

### Use Case 26: Divorced Parents - Scout in Two Households

**Actors:** Mom, Dad, Step-Mom, Scout

**Description:** A scout whose parents are divorced belongs to two separate households in the system.

**Setup Process:**
1. Admin creates separate new-household invitation links or QR codes for each parent
2. Mom enrolls with a passkey, sets up "Smith-Johnson Household", and adds scouts
3. System generates household link codes (or equivalent QR codes) for each scout
4. Mom shares those scout link codes with Dad through direct communication; this does not make Dad a co-manager of Mom's household
5. Dad enrolls with his own household invitation and passkey, sets up "Smith Household", and may invite Step-Mom as a co-manager on his household
6. Dad uses the scout link codes to add the existing scout profiles to his household
7. Step-Mom accepts a co-manager invitation on Dad's household and gains management permissions there only

**How Scheduling Works:**
- Scout appears in BOTH households
- Either household can sign the scout up for shifts
- ALL assignments are visible to BOTH households
- This prevents double-booking: if Mom signs up scout for Dec 15, Dad sees it
- Permission boundary: Each household can only cancel assignments they created
- If Mom signs up the scout, Dad cannot cancel it (and vice versa)

**Outcome:**
- Scout has one profile visible in both households
- Coordination happens naturally through shared visibility
- Each household maintains independence in what they can manage
- Step-parents have full management rights within their household

---

## Attendance Tracking

### Use Case 27: Authenticated Volunteer Checks Self In

**Actors:** Family Manager, Young Adult Scout

**Description:** An authenticated family manager or Young Adult Scout arrives at their shift and checks in using the responsive web app.

**What Happens:**
1. The assigned volunteer arrives during the shift's check-in window
2. Opens the web app and sees the current shift with "CHECK IN NOW" active
3. Selects "Check In"
4. System validates that the person is assigned, has a Confirmed current-season agreement status, and that the current server time is within 15 minutes before through 30 minutes after the scheduled start
5. Check-in is recorded using the current server timestamp; the user cannot choose or edit the time
6. The volunteer sees confirmation with their check-in time
7. Timer shows time on shift

**Outcome:** 
- The volunteer is marked as checked in
- Committee can see they're present
- An authenticated adult working the shift may now check in other arriving volunteers

---

### Use Case 28: Working Adult Checks In Another Volunteer

**Actors:** Authenticated Adult Volunteer, Arriving Volunteer

**Description:** An authenticated adult working the current shift—or finishing the immediately preceding shift during the handoff—checks in an arriving adult, managed scout, or Young Adult Scout.

**What Happens:**
1. The volunteer arrives during the target shift's check-in window
2. The adult opens the target shift roster
3. Sees list of assigned volunteers with pending/checked-in status
4. Selects the arriving volunteer's name
5. Selects "Check In"
6. System validates:
   - The actor is an authenticated adult
   - The actor is checked in to the target shift, or was checked in to the immediately preceding shift and is still within that shift's checkout window
   - The target is assigned to the target shift
   - The target's current-season agreement status is Confirmed
   - The current server time is within 15 minutes before through 30 minutes after the target shift's scheduled start
   - The target is not already checked in
7. The target's check-in is recorded using the current server timestamp, with the adult identified as the actor

**Outcome:**
- The arriving adult, managed scout, or Young Adult Scout is marked as checked in
- Relevant family managers can see the attendance status
- Audit trail shows which authenticated adult performed the check-in
- The adult may check in volunteers from any family; family relationship is not required

---

### Use Case 29: Committee Checks In an Arriving Volunteer

**Actors:** Committee Member, Arriving Volunteer

**Description:** An on-site committee member checks in a scheduled volunteer when no eligible authenticated adult volunteer is available.

**What Happens:**
1. The volunteer arrives during the check-in window
2. The committee member opens the shift roster in the web app
3. They locate the volunteer and select "Check In"
4. The system validates that the volunteer is assigned, has a Confirmed current-season agreement status, is not already checked in, and the current server time is within the check-in window
5. The check-in is recorded at the current server time with the committee member identified as the actor

**Outcome:**
- The volunteer is marked as checked in without needing their own phone or credentials
- Relevant family managers can see the check-in
- The audit trail identifies who performed the check-in
- Committee status does not allow prospective or retroactive check-in outside the time window

---

### Use Case 30: Working Adult Checks Out Another Volunteer

**Actors:** Authenticated Adult Volunteer, Volunteer Ending Shift

**Description:** Near the end of a shift, an authenticated adult working that shift checks out another adult, managed scout, or Young Adult Scout.

**What Happens:**
1. The shift enters its checkout window, from 15 minutes before through 30 minutes after its scheduled end
2. The adult opens the shift roster showing current attendance status
3. The adult selects a checked-in volunteer and chooses "Check Out"
4. The system validates that the actor is an authenticated adult checked in to the same shift
5. The system records checkout using the current server timestamp and identifies the adult as the actor
6. Confirmation shows the hours worked by the checked-out volunteer

**Outcome:**
- Hours worked are calculated and recorded for the checked-out volunteer
- Complete attendance record for the shift
- Relevant family managers can see the completed shift with hours

---

### Use Case 31: Committee Reviews Shift Attendance

**Actors:** Committee Member

**Description:** Committee reviews attendance for a completed shift and records audited corrections without creating retroactive check-in or checkout events.

**What Happens:**
1. Committee opens shift attendance view
2. Sees summary: how many checked in, checked out, still checked in, pending/no-show
3. Reviews individual records showing:
   - Who checked in and when
   - Who checked out and when
   - Who checked them in (if done by another person)
   - Hours worked
   - Any local two-deep coverage transitions or closure event during the shift
4. For volunteers who forgot to check out:
   - Creates an attendance adjustment with the approved hours or departure time
   - Adds a required note and supporting reason
   - Does not create or backdate a normal checkout event
5. For volunteers who didn't show:
   - Marks as "No Show"
   - Adds follow-up note
6. A missing check-in cannot be backdated; if attendance is verified later, Committee records a separate attendance adjustment with a required explanation

**Outcome:**
- All attendance records are complete
- Hours are calculated for everyone
- No-shows are flagged for follow-up
- Normal check-in/out events always retain their actual server time
- Adjustments are clearly distinguished from real-time events and preserve a complete audit trail

---

### Use Case 32: Family Manager Views Attendance History

**Actors:** Family Manager

**Description:** A Family Manager views the attendance history for their household.

**What Happens:**
1. Family Manager opens "History"
2. Sees household summary:
   - Total shifts worked
   - Total hours across the family
3. Sees breakdown by person:
   - Each family member's shift count and hours
4. Can tap on any person to see detailed shift history:
   - Each shift with date, time, hours worked
   - Who performed check-in/check-out

**Outcome:** Family Manager has full visibility into household participation and can verify hours for troop requirements.

---

## Walk-In Coverage

### Use Case 33: Walk-In Covers for No-Show

**Actors:** Committee Member, Volunteer covering for no-show

**Description:** When a scheduled volunteer doesn't show up, someone else steps in.

**What Happens:**
1. Shift is in progress and scheduled volunteer hasn't arrived
2. Another volunteer (e.g., just finished prior shift) offers to stay and cover
3. Committee member adds the volunteer as a "walk-in":
   - Searches for the volunteer
   - Confirms the volunteer's current-season agreement status is Confirmed
   - Adds them with a note explaining the situation
4. Walk-in is automatically checked in at time of addition
5. Original no-show is marked as such with explanation
6. Walk-in works and checks out normally

**Outcome:**
- Walk-in volunteer gets credit for hours worked
- No-show is documented for follow-up
- Shift maintains adequate staffing
- Clear audit trail of what happened

---

### Use Case 34: Checked-In Authenticated Adult Adds Walk-In Scout

**Actors:** Checked-In Authenticated Adult, Unscheduled Scout

**Description:** An authenticated adult who is already working a shift adds an unscheduled scout who offers to help.

**What Happens:**
1. Shift is in progress and the authenticated adult is checked in
2. An unscheduled scout arrives (e.g., picking up a sibling) and offers to help
3. Adult can add them as a walk-in without separate committee approval:
   - Taps "Add Walk-In" on shift roster
   - Searches for and selects the scout
   - System confirms the scout's current-season agreement status is Confirmed
   - Adds a note (optional)
4. Scout is immediately checked in as a walk-in
5. Scout works and is checked out by an eligible authenticated adult

**Key Point:** Checked-in authenticated adults can add scout walk-ins to their own shift without separate committee approval. This empowers on-the-ground decision making when extra help is available.

**Outcome:**
- Scout gets credit for hours worked
- Walk-in is clearly marked in records
- Authenticated adult who added the walk-in is recorded
- No committee approval needed

---

### Use Case 35: Scout from Prior Shift Extends as Walk-In

**Actors:** Scout who just finished their shift

**Description:** A scout who just finished their scheduled shift stays to help with the next shift.

**What Happens:**
1. Scout finishes and checks out of their scheduled shift
2. Next shift needs help
3. Scout offers to stay longer
4. A checked-in authenticated adult on the next shift adds the scout as a walk-in
5. Scout works additional hours
6. A checked-in family manager or committee member checks the scout out when done

**Outcome:**
- Scout credited for both shifts separately
- Clear distinction between scheduled time and walk-in time
- Both households (if applicable) see all hours

---

## Hours Tracking & Leaderboards

### Use Case 36: Scout Hours and Stats Are Viewed

**Actors:** Family Manager, Young Adult Scout

**Description:** A family manager views statistics for any scout in the family, or a Young Adult Scout views their own statistics.

**What Happens:**
1. A family manager opens family statistics and selects a scout, or a Young Adult Scout opens "My Stats"
2. The web app shows the selected scout's summary:
   - Total hours worked
   - Number of shifts completed
   - Average hours per shift
   - Rank among all volunteers
3. A family manager may also see the family summary:
   - Combined family hours
   - Family rank among all families
4. Can view list of recent shifts with hours for each

**Outcome:** Family managers can track every scout's participation, and Young Adult Scouts can track their own.

---

### Use Case 37: Family Manager Views Individual Leaderboard

**Actors:** Family Manager

**Description:** A Family Manager views the ranking of all volunteers by hours worked.

**What Happens:**
1. Family Manager opens "Leaderboard"
2. Sees individual leaderboard showing:
   - Rank, name, hours, shift count
   - Top 10 or more volunteers
   - Their own position highlighted (with "YOU" indicator)
3. Can identify top contributors in the troop

**Outcome:** Creates friendly competition and recognition for participation.

---

### Use Case 38: Family Manager Views Family Leaderboard

**Actors:** Family Manager

**Description:** A Family Manager views the ranking of family units by combined hours.

**What Happens:**
1. Family Manager switches to "Family" view on leaderboard
2. Sees family leaderboard showing:
   - Rank, family name, total hours, shift count
   - Their own family highlighted
3. Can tap on their family to see breakdown by member:
   - Each person's contribution to the total
   - Hours are deduplicated for scouts in multiple households

**Important Note:** For scouts in multiple households, their hours count once toward the family total, not doubled.

**Outcome:** Families can see their combined contribution and how they compare to others.

---

### Use Case 39: Committee Views Season Statistics

**Actors:** Committee Member

**Description:** Committee views overall participation metrics for the season.

**What Happens:**
1. Committee opens "Season Stats"
2. Sees overall summary:
   - Total volunteers
   - Total families
   - Total hours worked
   - Total shifts completed
3. Sees shift coverage stats:
   - Percentage of scheduled vs. walk-in assignments
   - No-show rate
   - Critical coverage alerts, prevented openings, and closed shifts
4. Sees top contributors in each category
5. Can drill down into detailed reports

**Outcome:** Committee has visibility into troop participation and can identify top contributors for recognition.

---

## End-of-Season Reports

### Use Case 40: Treasurer Finalizes Scout Bucks Awards

**Actors:** Treasurer (Committee Member), Admin

**Description:** After the season's profit is known, the Treasurer converts finalized credited hours into dollar-denominated Scout Bucks awards and exports the result for the troop's separate account-management process.

**Context:** Scout Bucks are dollars that scouts can later use toward troop dues or trips. During the season, the dollar value is unknown. The scheduler therefore tracks only provisional credited hours based on:
- Hours the scout personally worked
- An equal share of hours worked by each parent, step-parent, or guardian associated with that scout

**Before Dollar Finalization:**
1. Family and committee views label Scout Bucks credited hours as Provisional
2. No estimated dollar balance or dollars-per-hour promise is displayed
3. Committee completes attendance adjustments and reviews the season's adult-to-scout relationships
4. The system calculates each scout's finalized credited hours

**Finalization Process:**
1. Treasurer opens "Reports"
2. Selects "Scout Bucks Report" for the completed season
3. Reviews the credited-hours report showing for each scout:
   - Scout's own hours and shift count
   - Allocated adult hours, broken down by adult and showing the number of eligible scouts sharing those hours
   - Total credited hours
4. Reviews each adult's allocation:
   - Total hours worked
   - Eligible scouts
   - Equal share allocated to each scout
5. Treasurer enters the distributable profit pool, in dollars and cents, after tree-sale expenses and troop hold-backs
6. System previews:
   - Total finalized credited hours
   - Informational effective dollars-per-credit-hour rate
   - Each scout's exact proposed dollar award
   - Confirmation that awards sum exactly to the entered pool
7. Treasurer confirms finalization
8. System stores an immutable settlement revision with the pool, credited-hour snapshot, calculated awards, rounding allocation, actor, and timestamp
9. Treasurer exports the final award report as CSV for the separate system or process that maintains Scout Bucks balances and redemptions

**Important Calculation Rules:**
- Each adult's hours are allocated once and split equally among that adult's eligible scouts
- A parent or guardian's eligible scouts include their active scouts across linked households
- A step-parent's eligible scouts are the active scouts in the household(s) where the step-parent participates
- A scout is included only once in an adult's allocation, even when linked through multiple households
- Scout's own hours are counted once even if in multiple households
- If an adult has no eligible scouts, their hours remain in volunteer statistics but are not allocated to Scout Bucks
- Credited-hour allocation uses full precision
- Each unrounded award is proportional to the scout's share of total finalized credited hours
- Dollar awards are calculated from integer cents, not binary floating-point currency values
- Remaining cents are assigned using a deterministic largest-remainder method so awards sum exactly to the distributable pool
- If total credited hours are zero, a nonzero pool cannot be finalized
- Finalization does not create an ongoing balance or spending ledger in this application

**Corrections:**
- A finalized settlement is never edited in place
- Treasurer or Admin may create a corrected revision only after recording a reason and re-confirming the credited-hour snapshot and distributable pool
- Exports and views clearly identify the current revision and retain prior revisions in the audit history

**Outcome:** The scheduler produces finalized Scout Bucks dollar awards that exactly distribute the entered profit pool. Another troop process manages the resulting balances and future spending.

---

### Use Case 56: Admin Archives and Deletes a Completed Season

**Actors:** Admin

**Description:** After end-of-season reporting is complete, an Admin creates and safely stores a restorable archive, then manually removes that season and all of its owned data from the live application.

**Preconditions:**
- The season is completed and inactive
- No shift in the season is in progress
- The Admin has reviewed and finalized end-of-season reports
- No automatic process can initiate archival or deletion

**Archive Process:**
1. Admin opens the completed season's maintenance view and selects "Create Archive"
2. System displays the data that will be included and generates a consistent archive containing:
   - Season configuration and schedule
   - Assignments, attendance events, adjustments, walk-ins, and no-shows
   - The season's configured agreement link and per-person confirmation records
   - Season-specific reports and reporting inputs
   - Messages, delivery records, and audit records owned by the season
   - The minimum person, household, and family-unit snapshots needed to interpret historical records
   - The agreement URL, but not the externally maintained Google Doc contents
3. System creates a ZIP containing a format version, manifest, versioned JSON or CSV data files, record counts, generation time, and SHA-256 checksums
4. Admin enters and confirms an archive passphrase
5. System encrypts the ZIP using passphrase-based `age` encryption and downloads it as `season-{name}.zip.age`
6. The application never logs or stores the passphrase
7. Admin stores the encrypted archive and its passphrase separately
8. System verifies that archive generation completed successfully and displays the checksum
9. Creating an archive does not alter or hide the live season

**Deletion Process:**
1. After saving the archive, Admin separately selects "Delete Archived Season"
2. System shows the exact categories and counts that will be permanently removed
3. System explains that the operation cannot be undone without both the external archive and its passphrase
4. Admin re-authenticates with a passkey step-up, confirms that the archive and passphrase have been saved separately, and types the season name
5. System atomically removes all data owned exclusively by the season
6. Shared identities, person profiles, households, family units, reusable shift templates, and roles remain
7. System retains only a minimal non-personal deletion receipt containing the acting Admin, deletion time, former season ID, archive checksum, and deleted record counts

**Failure and Safety Rules:**
- Archive and deletion are never initiated by age, schedule, scheduled background job, or retention timer; an Admin must initiate each operation
- Deletion remains disabled until a successful archive exists and its checksum has been calculated
- An active, draft, or in-progress season cannot be deleted
- A failed deletion transaction leaves the complete season intact
- A prior archive does not silently update; the Admin creates a new archive if live season data changed after archival
- Restoring an archive is a separately controlled technical operation that prompts for the passphrase and is never automatic
- Losing the passphrase makes the archive unrecoverable

**Outcome:** The completed season no longer appears in schedules, history, reports, search, or administration. Its externally held archive remains available for controlled restoration.

---

## Schedule Management & Staffing Views

### Use Case 41: Committee Views Week Schedule with Staffing Levels

**Actors:** Committee Member

**Description:** Committee views the entire week's schedule with visual indicators of staffing status.

**What Happens:**
1. Committee opens "Week View"
2. Sees week summary:
   - Total shifts
   - How many fully staffed vs. understaffed vs. critical
3. Scrolls through each day seeing:
   - Each shift with time and location
   - Visual progress bars showing scout and parent fill rates
   - Status indicators: FULL, OK, LOW, CRITICAL, CLOSURE REQUIRED, CLOSED
   - Separate adult count, total-person minimum, and scheduled local two-deep compliance
   - Special event shifts marked with star icon
   - List of who's signed up
4. Can tap any shift for full details
5. Can navigate to other weeks using arrows or picker

**Staffing Levels:**
- **FULL** - All target slots are filled and scheduled coverage meets local two-deep and minimum-operating rules
- **OK** - The shift can operate safely but is not completely full
- **LOW** - The shift can operate safely but remains meaningfully below its target staffing
- **CRITICAL** - Scheduled coverage fails either the minimum operating headcount or local two-deep rule and the lot must close unless coverage changes
- **CLOSURE REQUIRED** - The response deadline has passed or actual attendance is unsafe and Committee/Admin must resolve coverage or record closure
- **CLOSED** - Committee or Admin has closed the shift and participation actions are disabled

**Outcome:** Committee has at-a-glance visibility into where coverage is needed.

---

### Use Case 42: Committee Reviews Staffing Alerts Dashboard

**Actors:** Committee Member

**Description:** Committee views prioritized list of shifts that need attention.

**What Happens:**
1. Committee sees alert badge indicating shifts needing attention
2. Opens "Staffing Alerts" dashboard
3. Sees prioritized list:
   - Grouped by severity (Critical, Low)
   - Ordered by date (soonest first)
   - Shows the exact volunteer shortfall
   - Identifies whether the problem is total headcount, local two-deep coverage, or both
4. For each understaffed shift:
   - Can send a targeted staffing reminder
   - Can send a troop-wide critical coverage alert when closure is possible
   - Can view full shift details
   - Can share signup link directly
5. As families sign up, alerts update automatically

**Outcome:** Committee can quickly identify and address staffing issues before they become problems.

---

### Use Case 57: Committee Sends a Critical Coverage Alert

**Actors:** Committee Member, Admin

**Description:** A shift is projected to close because it does not meet its minimum operating headcount or scheduled local two-deep coverage, so Committee asks the entire troop for urgent help.

**What Happens:**
1. The Staffing Alerts dashboard identifies a CRITICAL shift
2. Committee or Admin selects "Send Critical Coverage Alert"
3. The system previews:
   - Shift date, time, and location
   - Current adult, scout, and total-person counts
   - The specific unresolved coverage rule
   - The response deadline entered by the sender, which must be before the shift starts
   - A warning that the lot will close for that shift if coverage remains unresolved
   - A direct link to the shift signup page
4. The sender confirms the alert
5. The system publishes one canonical high-priority troop-wide announcement into every active Family Manager and Young Adult Scout in-app inbox
6. If the optional Groups.io integration is enabled, the same alert is posted there
7. Duplicate sends for the same unresolved condition require explicit confirmation and are recorded in the audit trail
8. If signup changes make the shift safe to operate, its status updates immediately and the system places a concise "coverage secured" update in the same recipients' inboxes and, when enabled, posts that update to Groups.io

**Outcome:** Every active Family Manager and Young Adult Scout is informed in-app that urgent signups are needed to prevent a specific shift closure; Committee and Admin can review the alert and delivery status in the web app.

---

### Use Case 58: System Enforces the Local Two-Deep Coverage Rule During a Shift

**Actors:** Assigned Adults, Scouts, Committee Member, Admin, System

**Description:** Scheduled coverage is a forecast; when people arrive, the system uses checked-in adults and scouts to ensure the lot never operates with a prohibited adult-to-scout combination.

**National Policy Boundary:**
- [Scouting America's national Youth Protection and Adult Leadership policy](https://www.scouting.org/health-and-safety/gss/gss01/) currently requires two registered adult leaders age 21 or older at all Scouting activities
- The scheduler's local adult classification does not prove Scouting America registration, age, training, or eligibility to serve as one of those leaders
- Committee remains responsible for confirming that the adults present satisfy the current national policy and any additional chartered-organization requirements

**What Happens:**
1. The first assigned adult checks in during the normal check-in window
2. The system evaluates the people currently checked in:
   - Two or more adults satisfy the local two-deep requirement regardless of their relationship to any scouts
   - Fewer than two adults means the lot cannot operate and no scout may work
3. Every scout check-in is rejected until at least two adults are checked in
4. Young Adult Scout access does not make a scout count as an adult; adult coverage is based on the person's recorded adult classification
5. If an adult checks out or departs and fewer than two checked-in adults remain:
   - Checkout is still recorded
   - The roster immediately displays a local two-deep coverage violation
   - Tree-lot operations stop and remaining volunteers check out unless another adult checks in immediately
   - Committee/Admin receives an urgent operational alert
6. The system records the compliance transition and actors without changing historical attendance events

**Outcome:** Actual attendance—not merely scheduled signups—ensures that the tree lot operates only while at least two adults are present.

---

### Use Case 59: Committee Closes a Shift for Insufficient Coverage

**Actors:** Committee Member, Admin

**Description:** A critical shift remains below its minimum operating or local two-deep requirement, so the tree lot closes for that shift.

**What Happens:**
1. Before the shift, Committee/Admin sees that the critical-alert deadline passed without safe coverage; or, at shift time, actual checked-in coverage is noncompliant
2. The system marks the shift "CLOSURE REQUIRED" and explains the unresolved rule
3. Committee/Admin selects "Close Shift" and records a reason
4. The shift status becomes CLOSED and the system:
   - Disables new signups, check-ins, and walk-in additions
   - Preserves assignments in the audit history and marks them cancelled by shift closure
   - Allows checkout for any existing open attendance records
   - Places a troop-wide closure notice in every active Family Manager and Young Adult Scout in-app inbox, with assigned volunteers clearly identified as affected
   - Posts the notice to Groups.io when that optional integration is enabled
5. A closure made before the shift starts may be reversed by Committee/Admin only after coverage satisfies every rule; reopening is audited and places an update in the same recipients' inboxes and, when enabled, on Groups.io
6. A shift closed after operations have begun cannot be reopened retroactively

**Outcome:** The closure is explicit, communicated troop-wide through the in-app inbox and optional Groups.io, and preserved in schedule and audit history rather than appearing as an unexplained empty shift.

---

### Use Case 43: Family Manager Uses Week View to Find Available Shifts

**Actors:** Family Manager

**Description:** A Family Manager looking to sign up household members uses the week view to find shifts that need help.

**What Happens:**
1. Family Manager opens the schedule's Week View
2. Sees staffing indicators for each shift:
   - Green checkmarks on fully-staffed shifts
   - Yellow indicators on shifts needing help
   - Red indicators on critical or closure-required shifts
   - Closed indicators for shifts that are no longer operating
3. Quickly identifies shifts with the most need
4. Taps a critical shift to see details and "We need your help!" messaging
5. Signs up family members for high-need shifts
6. Receives confirmation thanking them for helping where most needed

**Outcome:** Family Managers can easily find and fill shifts that most need coverage, rather than just picking convenient times.

---

### Use Case 44: Web App Loads Season and Week Navigation

**Actors:** Any User

**Description:** When a user opens the web app, the system determines the current season and appropriate week to display.

**What Happens:**
1. User opens the web app
2. System determines:
   - Current active season (or most recent if between seasons)
   - Which week of the season is current
   - All valid weeks for navigation
3. The web app loads the current week's schedule by default
4. User sees week navigation bar with:
   - Current week displayed
   - Arrows to navigate to adjacent weeks
   - Dropdown picker to jump to any week in the season
5. Week picker shows shift count for each week
6. Navigation is constrained to valid season dates

**Off-Season Behavior:**
- If no active season: Shows most recent completed season
- If only draft exists: Users see "Schedule not yet published" (Committee can still see draft)

**Outcome:** Users always land on relevant content and can easily navigate the full season.

---

## Profile Management

### Use Case 45: User Updates Profile Photo

**Actors:** Any authenticated Family Manager, Young Adult Scout, Committee Member, or Admin

**Description:** A user adds or updates their profile photo.

**What Happens:**
1. User navigates to their profile settings
2. Taps on profile photo area (shows placeholder if no photo set)
3. Chooses to take a new photo or select from photo library
4. Crops/adjusts the image as needed
5. Confirms the selection
6. Photo is uploaded through the browser and associated with their profile
7. New photo appears throughout the web app wherever the user is displayed (shift rosters, leaderboards, family views)

**Outcome:** User's profile photo is updated and visible to other authenticated users throughout the web app.

---

### Use Case 46: User Edits Display Name

**Actors:** Any authenticated Family Manager, Young Adult Scout, Committee Member, or Admin

**Description:** A user updates their display name.

**What Happens:**
1. User navigates to their profile settings
2. Taps on their display name field
3. Edits the name (first name, last name, or preferred display name)
4. Saves the change
5. Updated name appears throughout the web app

**Note:** Authentication is tied to passkeys and a claimed email account identifier rather than a phone number or social-login account. Credential and email changes follow Use Case 2B, including assisted recovery by an authorized Family Manager, co-manager, Admin, or secured break-glass process as appropriate.

**Outcome:** User's display name is updated throughout the system.

---

### Use Case 47: Authenticated Person Removes Own Login

**Actors:** Family Manager, Young Adult Scout, Committee Member, Admin

**Description:** An authenticated person requests removal of their personal login through the web app while preserving historical person and volunteer records.

**Preconditions:**
- If another family manager exists, the manager may remove their own login
- If this is the only manager for an active family, removal is blocked until another adult accepts a co-manager invitation or an administrator deactivates/transfers the family account
- A Young Adult Scout may remove their own login at any time; their scout profile returns to managed status
- Scouts without Young Adult Scout access do not have login accounts to delete
- A Committee Member may remove their login; committee privileges are revoked
- An Admin may remove their login only when another active Admin remains

**What Happens:**
1. The person navigates to account settings and selects "Remove My Access"
2. The system checks role-continuity requirements for Family Managers and Admins; this check is not required for a Young Adult Scout or Committee Member
3. The person reviews the consequences and re-authenticates with a passkey step-up
4. On confirmation, the system:
   - Revokes all sessions for that manager
   - Removes all registered passkeys and the claimed email from active authentication
   - Removes the person from future in-app inbox delivery and from any future verified-email notification preference
   - Removes the person's authenticated role and permissions
   - Preserves the family-member profile and historical attendance
   - Returns a Young Adult Scout profile to manager-controlled status
   - Records the removal in the audit trail
5. The person is signed out and returned to the welcome page

**Outcome:** The person's login is removed without deleting their family profile or historical volunteer records. A Young Adult Scout's schedule remains visible and manageable by family managers.

---

## Privacy & Data Requests

The following use cases support privacy regulation compliance (such as GDPR and CCPA). Requests use a separately verified process through troop leadership or a designated privacy contact; an authenticated web session alone is not sufficient for irreversible privacy actions.

### Use Case 48: User Requests Data Export

**Actors:** Any user (current or former), Admin/Privacy Contact

**Description:** A user requests a complete export of all personal data the system holds about them.

**Request Process:**
1. User contacts the troop's designated privacy contact through a published channel separate from normal web-app workflows
2. User provides identifying information to verify their identity
3. Privacy contact verifies the request is legitimate

**What Happens:**
1. Admin/Privacy contact initiates data export for the user
2. System compiles all data associated with the user:
   - Profile information (name, account email if still present, photo)
   - Household memberships
   - All shift assignments (past and future)
   - All attendance records (check-in/out times, hours worked)
   - Walk-in records
   - Any messages sent or received
   - Account activity history
   - Seasonal agreement links and confirmation records associated with the requester
3. Data is exported in a portable format (JSON or CSV)
4. Export is securely delivered to the user

**Outcome:** User receives a complete copy of all personal data held by the system, fulfilling data portability requirements.

---

### Use Case 49: User Requests Permanent Data Removal

**Actors:** Any user (current or former), Admin/Privacy Contact

**Description:** A user requests permanent removal or anonymization of their personal data.

**Request Process:**
1. User contacts the troop's designated privacy contact through a published channel separate from normal web-app workflows
2. User provides identifying information to verify their identity
3. Privacy contact verifies the request is legitimate
4. Privacy contact explains the implications:
   - Personal data, including agreement-confirmation records, will be deleted or anonymized
   - Historical hours will be removed from reports
   - This may affect Scout Bucks calculations for scouts in the user's household
   - This action cannot be undone

**What Happens:**
1. Admin/Privacy contact initiates permanent deletion
2. System permanently removes or anonymizes personal data:
   - User profile document is deleted (not just marked deleted)
   - Display name is removed from all historical records or replaced with "Deleted User"
   - Attendance records are either deleted or anonymized
   - Profile photo is permanently deleted from storage
   - User is removed from all household records
   - Any assignments (past and future) are deleted or anonymized
   - Agreement-confirmation records associated with the person are deleted or anonymized
3. System logs that a deletion request was fulfilled without retaining identifying information about the deleted person

**Important Considerations:**
- This affects reporting accuracy (Scout Bucks totals may change)
- If user was an adult, their allocated hour shares are removed from affected scouts' totals
- Admin should consider timing (ideally after season ends and reports are finalized)
- May want to export data first (Use Case 48) before permanent deletion

**Outcome:** Personal data, including agreement-confirmation records, is removed or anonymized. Some aggregate statistics may be affected.

---

## Permission Summary

| Action | Admin | Committee | Family Manager | Young Adult Scout | Managed Scout |
|--------|-------|-----------|----------------|-------------------|---------------|
| Create new-household invitations | ✓ | | | | |
| Invite a household co-manager | ✓ | | ✓ | | |
| Generate household link codes | ✓ | | ✓ | | |
| Create/manage shifts | ✓ | ✓ | | | |
| Send troop announcements | ✓ | ✓ | | | |
| View own in-app inbox and manage read state | ✓ | ✓ | ✓ | ✓ | |
| Manage own nonessential notification preferences | ✓ | ✓ | ✓ | ✓ | |
| Send troop-wide critical coverage alerts | ✓ | ✓ | | | |
| Close/reopen shifts for insufficient coverage | ✓ | ✓ | | | |
| Invite/remove family managers | ✓ | | ✓ | | |
| Grant/revoke Young Adult Scout access | ✓ | | ✓ | | |
| Add/edit family-member profiles | ✓ | | ✓ | | |
| Set the season's public agreement link | ✓ | | | | |
| Confirm own seasonal agreement | ✓ | ✓ | ✓ | ✓ | |
| Facilitate confirmation for a managed person | ✓ | | ✓ | | |
| View agreement confirmation status | ✓ | ✓ | Household | Own | |
| Open the current agreement link from profile | ✓ | ✓ | Household | Own | Via manager |
| Sign up an eligible family member | ✓ | ✓ | ✓ | Self only | |
| Cancel assignments | ✓ | ✓ | ✓* | Own only | |
| View schedules | ✓ | ✓ | Family | Own | |
| Real-time check in/out | Time-bound | Time-bound | Self and others** | Self only | |
| Post-shift attendance adjustment | ✓ | ✓ | | | |
| Add walk-ins | ✓ | ✓ | ✓**** | | |
| Deactivate families | ✓ | | | | |
| Edit own display name or photo | ✓ | ✓ | ✓ | ✓ | |
| Remove own authenticated access | ✓***** | ✓ | ✓*** | ✓ | |
| Process data export or permanent deletion requests | ✓ | | | | |
| Finalize and export Scout Bucks awards | ✓ | Treasurer only | | | |
| Archive and delete a completed season | ✓ | | | | |

*Household-owned assignments require a manager of the originating household; Young Adult Scout-created assignments may be cancelled by a manager in any linked household; Admin and Committee can override  
**Must be an authenticated adult checked in to the current shift, or completing the immediately preceding shift during the handoff window  
***A Family Manager cannot leave an active household account with no manager  
****Must be an authenticated adult checked in to the in-progress shift; preceding-shift handoff authority does not apply  
*****An Admin cannot remove access when they are the last active Admin

---

## Business Rules

This section consolidates the key business rules that govern system behavior across use cases.

### Core Terminology and Delivery Rules

**Core Records:**
- A **person profile** represents one adult or scout and owns that person's assignments, attendance, hours, and history
- An **authenticated identity** represents one person who can sign in; it may be linked to a Family Manager, Young Adult Scout, Committee Member, or Admin profile
- A **household account** is the family-management and scheduling boundary managed by one or more Family Managers; use cases that say "family account" refer to this household account
- A **family unit** is a reporting-only grouping of related households used to deduplicate shared scouts in family leaderboards and Scout Bucks calculations
- A person has one profile even when linked to multiple households or assigned multiple authenticated roles

**Delivery Constraints:**
- All functionality is delivered through the responsive web app; no workflow requires a native mobile application
- Go server handlers enforce every authorization and validation rule regardless of whether navigation uses HTMX or a full-page request
- HTMX enhances server-rendered HTML but is not an authorization or business-rules boundary
- Browser JavaScript is required for passkey (WebAuthn) registration and assertion and may be required for other browser APIs
- Heavy client-side frameworks such as Angular and React are not used

---

### Multi-Household Rules

**Household Structure:**
- Each household is independently managed by one or more parents/guardians
- Scouts can be members of multiple households simultaneously (e.g., Mom's household AND Dad's household)
- Each household manager can sign up scouts in their household for shifts
- Adult profiles belong to the household(s) in which they participate; linking a scout to another household does not automatically link the adults

**Adult-to-Scout Relationships:**
- Parent, step-parent, and guardian relationships are explicit links between person profiles rather than assumptions derived from household membership
- Family Managers record relationships for profiles they manage; Admin may correct disputed or duplicate relationships
- A relationship continues across linked households because it belongs to the people, not to one assignment or household
- Relationship creation, correction, and removal are audited and recalculate provisional Scout Bucks attribution
- Relationship details are not exposed to unrelated volunteers on shift rosters

**Invitations and Link Codes:**
- **New Household Invitation:** Created by Admin as a link or QR code and provided out of band to establish a household account and its first Family Manager
- **Co-Manager Invitation:** Created by an existing Family Manager or Admin as a link or QR code to grant management of an existing household to another parent/guardian
- **Young Adult Scout Invitation:** Created by a Family Manager or Admin as a link or QR code to grant limited access to an existing scout profile
- **Household Link Code:** Used by a family manager to add an existing scout profile to an additional household; may be shown as text or QR code
- Scouts do not receive authentication credentials unless a family manager explicitly grants Young Adult Scout access
- Invitations and link codes are single-use, expire after a configured short period, and are bound to their intended purpose
- New-household, co-manager, and Young Adult Scout invitations authorize enrollment and passkey registration; a Household Link Code is bound to one scout profile and does not create a login
- Co-manager and Young Adult Scout invitations may optionally record an intended claimed email, but possession of that email alone never grants access
- There is no open self-registration: every new household, co-manager, and Young Adult Scout identity begins with an authorized invitation

**Household Link Code Security:**
- Link codes are cryptographically random and cannot be guessed
- Link codes can be regenerated by household managers (invalidates the old code)
- There is no search feature for scouts—you must have the link code to add a scout
- Link codes require direct communication between parents, which is appropriate for custody situations
- Each scout has their own link code, so sharing one child's code doesn't expose siblings
- Sharing a scout link code never grants the recipient management of the issuing household

**Cross-Household Assignment Visibility:**
- Shift assignments are visible to ALL of the scout's households
- This prevents double-booking: If Mom signs up scout for Dec 15, Dad sees it in his view
- Coordination happens naturally through shared visibility
- Young Adult Scout access is attached to the scout's single profile, so the scout sees one personal schedule across linked households

**Cross-Household Cancellation Rules:**
- Family Managers can only cancel household-owned assignments made by their own household
- If Mom signs up the scout, Dad cannot cancel it (and vice versa)
- A Young Adult Scout can cancel any of their own assignments
- An assignment created by a Young Adult Scout is not owned by one household; the scout and family managers in any linked household can manage it
- Committee and Admin retain audited override authority
- Each household otherwise maintains independence in what it can manage

**Household Status:**
- Deactivation applies to one household account, not automatically to every linked person profile or authenticated identity
- Family Manager authority and new scheduling through the inactive household are suspended
- Future assignments owned by the inactive household are cancelled; historical assignments and attendance remain
- Shared scouts, Young Adult Scout access, and assignments remain active when supported by another active linked household
- If a scout has no active linked household after deactivation, new scheduling and Young Adult Scout access are suspended and future self-created assignments are cancelled
- Committee/Admin roles and access through another active household are not removed by household deactivation
- Reactivation restores household management but does not automatically recreate assignments cancelled during deactivation

---

### Authentication & Family Account Rules

**Authentication:**
- Authenticated access belongs to a family manager, Young Adult Scout, committee member, or administrator, not to the family account as a shared credential
- Each authenticated person has one claimed email account identifier and one or more registered passkeys
- A normalized email address is system-wide unique and can be linked to only one active authenticated identity at a time
- Two people cannot share a login email, even when they belong to the same family
- The same person uses one identity if they hold multiple roles, such as Family Manager and Committee Member
- An email may be reassigned to another person only after its previous identity link, passkeys, and sessions are explicitly revoked and the new owner enrolls through an authorized invitation or self-service change on their own identity
- Conflict messages do not reveal the identity or family currently associated with an email address
- Sign-in uses passkeys (WebAuthn); Google, Apple, SMS one-time codes, magic links, and other social identity providers are not used
- Email is an account identifier and future notification/recovery address; it is not mailbox-verified at enrollment or ordinary sign-in
- Unverified email must not be used for notifications or email-based recovery
- Authentication responses are rate-limited and do not disclose whether an email is registered
- Browser sessions use secure, HTTP-only, same-site cookies and can be revoked
- Users may revoke their own sessions; Family Managers may revoke Young Adult Scout sessions for profiles they manage; Admin may revoke any session
- Security-sensitive actions require a recent passkey step-up
- Self-service email change requires a recent passkey assertion and uniqueness of the new address; the new address remains unverified until a later verification workflow
- If no passkey remains, recovery follows Use Case 2B: a Family Manager may recover a Young Adult Scout, a co-manager or Admin may recover a Family Manager, and another Admin or the secured break-glass process may recover an Admin, by reissuing an enrollment invitation
- Changing email or passkeys updates the existing identity; it never creates a replacement person profile or discards roles, memberships, assignments, or history

**Bootstrap and Role Provisioning:**
- Exactly one first Admin may be created through the configured bootstrap enrollment token and passkey registration
- Bootstrap is permanently disabled after the first Admin is established
- All later authenticated access requires an authorized invitation or role assignment
- Only Admin may grant or revoke Committee and Admin roles
- The last active Admin cannot remove or lose the Admin role without first appointing another Admin or using the separately secured break-glass procedure

**Family Accounts:**
- A household account contains one or more Family Managers and profiles for parents, guardians, managed scouts, and Young Adult Scouts
- A household has one or more parents/guardians as Family Managers
- Scouts do not require an authenticated login; Young Adult Scout access is optional
- A Young Adult Scout has a distinct authenticated identity linked to their existing scout profile, not shared family-manager credentials
- "Young Adult Scout" is an application permission level and does not assert legal adulthood
- Family Managers and Admin can grant or revoke Young Adult Scout access; Family Managers continue to see and manage that scout's schedule
- Users never share passkeys or browser sessions; actions are attributed to the authenticated person who performed them

**Young Adult Scout Identity Association:**
- Family membership belongs to the scout profile and exists before authenticated access is granted
- A Young Adult Scout invitation references one existing scout profile and the Family Manager or Admin who authorized it
- Successful invitation acceptance and passkey registration links an authenticated identity to that profile; it does not create a second family member
- On sign-in, the session resolves to the authenticated identity, then to the linked scout profile and its family memberships
- Self-service actions always use the linked scout profile as their target; client-submitted IDs cannot be used to act as another family member

---

### Seasonal Agreement Rules

**External Agreement Link:**
- Troop leadership owns and maintains the rules of conduct in a publicly readable Google Doc outside the scheduling system
- Each season stores one display title and one public HTTPS Google Doc link that is currently in effect
- The scheduling system does not upload, copy, proxy, render, cache, version, or retain the document contents
- The system cannot detect changes made to a Google Doc while its URL remains unchanged
- If the applicable rules change, Admin must replace the configured link
- Replacing the link atomically resets every person's confirmation for that season
- A prior season's confirmation never satisfies the current season

**Required Confirmation:**
- Every adult and scout who may participate has a boolean Confirmed or Not Confirmed status for the season
- Each participant explicitly selects "I have read and agree to the tree-lot rules of conduct"
- An authenticated adult or Young Adult Scout confirms through their own session
- A Managed Scout or adult without login access may confirm on a Family Manager's device, with both the participant and authenticated facilitating manager recorded
- No typed signature, guardian-consent record, paper form, uploaded scan, or agreement document is collected
- Merely opening the public link does not set the confirmation boolean

**Confirmation Record:**
- A confirmation records the person profile, season, current agreement-link identifier, boolean value, server timestamp, and acting authenticated identity
- Family Managers can see confirmation status for profiles in households they manage
- Young Adult Scouts can see their own status
- Committee can see Confirmed or Not Confirmed status
- Admin can also see the confirmation time and facilitating actor

**Participation Gate:**
- Invitation and authentication occur before confirmation so users can access agreement-first onboarding
- Unconfirmed users may manage profiles, confirm agreements, read announcements, and view published schedules within their normal permissions
- A Family Manager who is not personally confirmed may still facilitate confirmation or manage scheduling for another person who is confirmed
- The selected volunteer's confirmation—not merely household or acting-manager status—controls eligibility
- A person must be Confirmed for the shift's season and current agreement link before signup, assignment, check-in, or walk-in creation
- Checkout remains allowed for an existing open attendance record if the agreement link later changes, so hours and audit history can be closed correctly
- Committee and Admin cannot override confirmation or set it without the participant's explicit checkbox submission

**Privacy and Deletion:**
- Removing authenticated access does not by itself delete the person's profile or confirmation record
- A verified permanent-deletion request deletes or anonymizes the person's confirmation records with their other personal data
- Otherwise, confirmation records remain until an Admin manually archives and deletes the completed season; no scheduled process deletes them

---

### Shift Scheduling Rules

**Signup Validation (in order):**
1. Acting user must be authenticated and active
2. Target household and selected person profile must be active; a Young Adult Scout linked to multiple households needs at least one active linked household
3. Acting user must be authorized for the selected person:
   - A Family Manager may select an eligible member of a household they manage
   - A Young Adult Scout may select only their own linked scout profile
   - Committee or Admin may select any eligible active person as an audited override
4. Selected person must be Confirmed for the shift's season and current agreement link
5. Shift must be published, accepting signups, not cancelled or closed, and not in the past
6. Selected person must not already be assigned to the shift
7. Shift must have an available slot matching the selected person's role (scout or parent/adult)
8. Capacity and duplicate checks must be enforced atomically on the server to prevent concurrent overbooking
9. The assignment records the selected volunteer, acting user, originating household when applicable, and whether override authority was used
- A person may hold at most one assignment on a shift and cannot simultaneously occupy both scout and parent/adult slots
- Assignments target person profiles; the selected volunteer does not need authenticated access
- A signup may be accepted while projected coverage is CRITICAL so additional volunteers can continue filling the shift
- Every signup atomically recalculates total headcount and scheduled local two-deep coverage

**Cancellation Validation:**
- Committee and Admin can cancel any assignment (override capability)
- A household-owned assignment can be cancelled by Family Managers of the originating household
- An assignment created directly by a Young Adult Scout can be cancelled by that scout or a Family Manager in any linked household
- In multi-household situations, household ownership applies except to assignments created by the Young Adult Scout
- Every cancellation records the acting user and frees capacity only once; repeated requests are idempotent
- A cancellation that makes coverage CRITICAL remains allowed, but the user sees a prominent warning and Committee/Admin receives a staffing alert

---

### Local Two-Deep & Minimum Coverage Rules

**Policy Authority:**
- This section implements the troop's separately approved local tree-lot operating rule requiring at least two adults throughout every operating shift
- Troop leadership and the chartered organization must document approval of this local rule before it is enabled
- [Scouting America's national Youth Protection and Adult Leadership policy](https://www.scouting.org/health-and-safety/gss/gss01/) currently requires two registered adult leaders age 21 or older at all Scouting activities
- The application does not represent its adult classification alone as proof of national-policy compliance because it does not verify leader registration, age, training, or eligibility
- Committee must verify national-policy compliance independently and review the current national policy before each tree-lot season
- If another governing policy imposes stricter supervision, the stricter policy controls and these use cases must be revised before deployment

**Two Separate Operating Requirements:**
- Each shift defines a minimum operating headcount of at least two people; this is distinct from the larger target staffing used to determine whether a shift is full
- Local two-deep coverage always requires at least two checked-in adults while the lot is operating
- Both minimum headcount and local two-deep coverage must be satisfied for the lot to operate

**Relationship Independence:**
- The two adults do not have to be related to each other or to any scouts on the shift
- Parent, step-parent, guardian, household, and family-unit relationships do not create an exception to the two-adult minimum
- One adult cannot operate the lot alone, whether accompanied by their own children, unrelated scouts, other family members, or no scouts

**Who Counts as an Adult:**
- Adult coverage is determined by the person's recorded adult classification, not by authentication role or household
- Young Adult Scout access does not by itself make a person count toward adult coverage
- Committee/Admin counts toward actual coverage only while physically present and checked in to the shift
- Adult walk-ins count after their real-time check-in is recorded

**Projected and Actual Coverage:**
- Projected coverage uses active assignments and drives FULL, OK, LOW, or CRITICAL schedule status
- Actual coverage uses open check-in records and controls whether the lot may operate
- Signups remain open while projected coverage is CRITICAL because a later adult signup can resolve the deficiency
- At shift time, no scout can check in and the lot cannot operate until two adults are checked in
- If actual adult coverage drops below two after opening, operations stop and remaining volunteers check out unless another adult checks in immediately
- Adult checkout is never blocked, but a checkout that creates noncompliance triggers an urgent alert and closure-required state

**Closure State:**
- A shift becomes CLOSURE REQUIRED when its critical-alert deadline passes without projected safe coverage, its scheduled start arrives without safe coverage, or actual attendance becomes noncompliant during operation
- CLOSURE REQUIRED is an operational warning, not an automatic claim that the physical lot has closed
- Committee/Admin records the closure decision and reason; the resulting CLOSED state disables signup, check-in, and walk-in creation while preserving checkout and audit history
- Reopening is allowed only before the shift begins, after both projected requirements are satisfied, and through an audited Committee/Admin action

---

### Attendance & Check-In Rules

**Actor Definitions:**
- An **authenticated adult** is an adult person profile with an active authenticated identity, such as a Family Manager or an adult serving as Committee/Admin
- A non-manager adult profile without authenticated access cannot operate the roster; another eligible adult or on-site Committee/Admin must perform the action
- "Immediately preceding shift" means the shift at the same location whose scheduled end is nearest to, but not after, the target shift's scheduled start

**Timing Windows:**
- Check-in is allowed only from 15 minutes before through 30 minutes after the target shift's scheduled start
- Check-out is allowed only from 15 minutes before through 30 minutes after the target shift's scheduled end
- All timing decisions and attendance timestamps use server time
- Users cannot submit a custom, past, or future timestamp
- No role, including Committee or Admin, can create a normal check-in or checkout event outside these windows

**Self Check-In and Checkout:**
- The assigned person's current-season agreement confirmation is revalidated before check-in; checkout requires an existing open attendance record
- Authenticated adults assigned to a shift can check themselves in/out
- Young Adult Scouts assigned to a shift can check themselves in/out
- Young Adult Scouts can check in/out only themselves
- Managed scouts are checked in/out by an authorized checked-in adult or Committee/Admin
- Adult check-in is allowed when it improves coverage; scout check-in must also pass the actual local two-deep rule
- Checkout requires an open check-in for the same shift; repeated check-in and checkout requests are idempotent

**Working Adult Checking In Others:**
1. Actor must be an authenticated adult
2. Actor must be checked in to the target shift, or have been checked in to the immediately preceding shift
3. A preceding-shift adult may act only during the handoff period while their shift is in its checkout window and the target shift is in its check-in window
4. Target may be an adult, managed scout, or Young Adult Scout from any family
5. Target must be assigned to the target shift and not already checked in
6. Current server time must be inside the target shift's check-in window
7. If the target is a scout, the resulting checked-in group must satisfy the local two-deep rule

**Working Adult Checking Out Others:**
1. Actor must be an authenticated adult checked in to the same shift as the target
2. Target may be an adult, managed scout, or Young Adult Scout from any family
3. Target must have an open check-in for that shift
4. Current server time must be inside the target shift's checkout window
- Preceding-shift handoff authority applies only to check-in, never to checkout

**Committee/Admin Override:**
- On-site Committee and Admin can check assigned volunteers in/out without being assigned to the shift, but only during the same real-time windows
- Can mark an assigned volunteer as a no-show only after the shift's check-in window has closed
- Can create a post-shift attendance adjustment for verified missing or incorrect attendance
- An adjustment records the acting user, creation time, corrected value, and required reason
- An adjustment never creates or backdates a normal check-in/out event
- Raw real-time events are immutable; an adjustment is stored alongside them and determines corrected hours without erasing the original audit trail

---

### Walk-In Coverage Rules

**When Walk-Ins Can Be Added:**
- Walk-ins can ONLY be added to shifts that are in progress (started but not ended)
- Cannot add walk-ins before shift begins or after shift ends
- A walk-in is a real-time exception to the scheduled check-in window because no assignment existed before arrival

**Who Can Add Walk-Ins:**
- Committee and Admin can add any eligible adult or scout as a walk-in to an in-progress shift
- Any authenticated adult checked in to the in-progress shift can add an eligible scout as a walk-in without separate committee approval

**Walk-In Processing:**
- Target person must be Confirmed for the shift's season and current agreement link
- A scout walk-in is rejected unless the resulting checked-in group satisfies the local two-deep rule
- An adult walk-in may be added to resolve a coverage deficiency
- Walk-in assignments are automatically checked in at time of creation
- The assignment and check-in use the current server time; neither may be backdated or future-dated
- The assignment is recorded as a walk-in rather than a scheduled assignment
- Walk-in hours count toward totals; reporting distinguishes walk-in assignments from scheduled assignments
- Duplicate-person and role-eligibility checks still apply
- Walk-ins may exceed scheduled capacity because they represent real-time coverage or extra help; the roster records the resulting over-capacity state and acting user
- When a walk-in covers a no-show, the original assignment remains in the audit trail and Committee marks it as no-show with a note
- Walk-in checkout follows the same real-time checkout window and actor rules as scheduled attendance

---

### Notification Rules

**Channel Policy:**
- The system does not send operational notifications by SMS or to phone numbers
- Every notification for an authenticated recipient is recorded in that recipient's personal in-app inbox
- Groups.io is an optional deployment integration and is disabled by default
- When Groups.io is enabled, it may receive troop-wide messages only: troop announcements, season-publication summaries when selected, critical coverage alerts and their resolution updates, and shift-closure or reopening notices
- Direct messages to an individual authenticated user, and messages scoped to one household or another targeted eligible set, remain in-app only and are never posted to Groups.io
- A later mailbox-verification workflow may allow a user who opts in to also receive notifications at a verified email address; unverified claimed email must not be used for notification delivery
- Until that verified-email channel exists, in-app inbox delivery is the required recipient channel and Groups.io remains the only optional external channel for troop-wide messages

**In-App Inbox:**
- Each authenticated identity has a personal inbox of announcements, reminders, staffing notices, closures, and other operational messages
- Read/unread state is stored separately for each authenticated identity, including a read timestamp
- New inbox messages are unread for each active recipient except the authenticated human author
- Pre-existing messages do not become unread when a new authenticated identity is granted access
- Displaying the complete message marks it read; users may mark it unread again or mark all inbox messages read
- Optional Groups.io delivery and any future opted-in email delivery do not alter in-app read state
- Individual read state is private and is not treated as delivery confirmation
- Recipient users do not see other people's delivery details

**Troop Announcements:**
- Every announcement is placed in the in-app inbox of all active Family Managers and Young Adult Scouts
- When Groups.io is enabled, the same announcement is posted to the configured troop group
- A canonical copy is retained with the acting committee member, timestamp, content, priority, and per-channel delivery status
- Active Family Managers and Young Adult Scouts can view current and historical inbox messages in reverse chronological order
- Committee and Admin can view the same announcements and their delivery status
- In-app visibility is independent of every external delivery channel
- Every enabled delivery channel is tracked independently
- Retries are idempotent and target only failed external deliveries so recipients do not receive duplicate messages
- When Groups.io is enabled, its failure does not roll back successful in-app publication, and an in-app or Groups.io failure does not roll back the other successful channel
- Troop announcements are not suppressed by optional shift-reminder preferences
- Priority is stored and displayed with the announcement but does not change the required inbox recipient set

**Draft vs Published Shifts:**
- Draft shifts do NOT generate notifications (prevents spam during bulk creation)
- Only published shifts are visible to regular users

**Bulk Schedule Publishing:**
- Publishing an entire season may generate ONE in-app summary notification per active Family Manager and Young Adult Scout, according to the Committee's publish choice
- When Groups.io is enabled, Committee may also choose to post that troop-wide summary there
- The notification highlights special events (Lot Setup, Tree Delivery, etc.)
- Recipients receive one comprehensive notification, not one message per shift

**Individual Shift Additions:**
- Shifts added individually after season publication generate an in-app inbox message for active Family Managers and Young Adult Scouts eligible to fill the new shift
- Because recipients are a targeted eligible set, Groups.io is not used
- This is appropriate for unplanned additions during the season

**Shift Reminders:**
- Automated in-app reminders are recorded approximately 24 hours before shifts
- For a household-owned assignment, recipients are active Family Managers in the originating household
- For an assignment created directly by a Young Adult Scout, recipients are active Family Managers in every linked household
- A Young Adult Scout receives the reminder directly for their own assignment in addition to the relevant Family Managers
- The reminder names the family member assigned to the shift and links to the web app
- Family Managers and Young Adult Scouts can opt out of nonessential shift reminders in notification preferences
- Shift reminders are direct/family-scoped and are never posted to Groups.io
- Invitation and recovery enrollment links are not controlled by notification preferences; when verified-email notifications exist, transactional security notices are likewise independent of optional reminder preferences

**Agreement Reminders:**
- The web app shows an Agreement Center alert while any person in the household is Not Confirmed for the current season
- The system may place configurable, deduplicated in-app reminders for Family Managers about unconfirmed household members and directly for a Young Adult Scout who has not confirmed
- Reminders identify the person, link to the Agreement Center, follow operational-message notification preferences, and remain in-app only

**Targeted Staffing Reminders:**
- Committee may send a staffing reminder to active Family Managers and Young Adult Scouts eligible for a specific understaffed shift
- The reminder identifies the shift and links directly to its signup page
- A targeted staffing reminder is stored as an in-app operational message for the eligible recipients only; Groups.io is not used unless Committee separately sends a troop-wide announcement

**Critical Coverage Alerts and Closures:**
- A critical coverage alert is a troop-wide high-priority announcement, not a targeted staffing reminder
- The in-app announcement is placed for every active Family Manager and Young Adult Scout and is not suppressed by optional shift-reminder preferences
- The alert identifies the shift, unresolved minimum-headcount or local two-deep rule, response deadline, closure warning, and direct signup link
- When the condition is resolved, one deduplicated coverage-secured update is placed in the same recipients' inboxes
- A recorded shift closure generates a troop-wide in-app notice
- The optional Groups.io channel receives the same alert, resolution, and closure messages only when enabled
- Delivery attempts and per-channel results are audited and idempotent

**Future Verified-Email Notifications:**
- Claimed account email is an identifier and is not mailbox-verified at enrollment or ordinary sign-in
- After a future verification workflow proves mailbox ownership, a user may opt in to also receive selected notifications at that verified address
- Opted-in email delivery supplements the in-app inbox; it does not replace inbox records or private read state
- Unverified email must not be used for notifications or email-based recovery

---

### Hours Tracking & Leaderboard Rules

**Individual Hours:**
- Each person's unadjusted hours are calculated from immutable real-time check-in and checkout events
- When Committee/Admin records an approved attendance adjustment, reports use the corrected hours while retaining the original events and adjustment audit trail
- Checkout must be later than check-in, and calculated hours cannot be negative
- Walk-in hours and scheduled hours are both counted
- Hours are tracked per shift for detailed history

**Family Leaderboard Deduplication:**
- For scouts in multiple households, hours are counted ONCE toward the family total (not doubled)
- A configured family unit groups related households for reporting; household linking alone does not silently merge unrelated adults into a reporting unit
- Each distinct person profile contributes hours at most once to a family-unit total, even when linked through multiple households or roles
- Admin/Committee must be able to review the households and profiles included in a family unit

**Statistics and Leaderboard Visibility:**
- Family Managers can view individual and family leaderboards and detailed statistics for profiles in households they manage
- Young Adult Scouts can view only their own detailed statistics; they do not gain access to private family-member details
- Committee and Admin can view season-wide statistics and reporting details

---

### Scout Bucks Calculation Rules

Scout Bucks credits are based on each scout's own hours plus equal allocations of associated adults' hours.

**Attribution Rules:**
- Rules apply to every scout profile regardless of whether the scout is managed or has Young Adult Scout access
- Each adult's hours form one allocation pool and are split equally among that adult's eligible scouts
- A parent or guardian's eligible scouts are their active scout profiles across linked households
- A step-parent's eligible scouts are active scout profiles in the household(s) where the step-parent participates
- The same scout is included only once in an adult's allocation, even when multiple household paths connect them
- Scout's own hours are counted once, even if in multiple households
- Each adult's hours are allocated only once, even when the adult has multiple roles or household links
- Adults with no eligible scouts retain their volunteer hours but contribute no hours to Scout Bucks
- Approved attendance adjustments, rather than superseded raw event totals, are used in final calculations
- Both scheduled and walk-in hours are eligible for Scout Bucks calculations
- Allocations are calculated at full precision; display rounding must not cause allocated shares to exceed or fall short of the adult's source hours
- Committee reviews and finalizes the season's adult-to-scout allocation roster before the Treasurer enters the distributable profit pool

**Dollar Finalization Rules:**
- Credited hours remain Provisional until attendance adjustments and adult-to-scout relationships are finalized
- No dollar estimate is displayed before finalization
- Treasurer enters one distributable pool in integer dollars and cents after expenses and hold-backs
- The system snapshots credited hours and calculates each scout's proportional share of the pool
- Currency calculations use integer cents; credited-hour ratios retain full precision
- Whole cents are first allocated without exceeding each scout's exact proportional share
- Remaining cents are distributed by largest fractional remainder, with a stable person-ID tie-breaker
- Final awards must sum exactly to the entered distributable pool
- A nonzero pool cannot be finalized when total credited hours are zero
- A finalized settlement is immutable; a correction creates an audited new revision

**System Scope:**
- The scheduler displays and exports finalized seasonal awards
- It does not maintain Scout Bucks account balances
- It does not record trip or dues redemptions, transfers, expiration, or manual financial adjustments

**Example:**
- Scout Alex worked 12 hours
- Scout Emma worked 8 hours
- Mom Sarah worked 18.5 hours
- Dad John worked 14 hours
- Step-Mom Lisa worked 5.5 hours
- Alex and Emma are the two eligible scouts for each adult
- Each scout receives 9.25 hours from Sarah, 7 hours from John, and 2.75 hours from Lisa
- Alex's credited total = 12 + 9.25 + 7 + 2.75 = 31 hours
- Emma's credited total = 8 + 9.25 + 7 + 2.75 = 27 hours
- Total credited hours = 58, exactly matching the family's 58 actual volunteer hours
- If the Treasurer enters a $580 distributable pool, the effective rate is $10 per credited hour
- Alex's finalized award is $310 and Emma's is $270; together they exactly equal the $580 pool

---

### Template & Schedule Rules

**Template Independence:**
- Changes to templates do NOT affect previously generated schedules
- Each season's shifts are independent once created
- Historical data remains accurate even if templates are updated
- Generated shifts snapshot their target adult/scout counts and minimum operating headcount

**Operating Coverage Configuration:**
- Templates and individually created shifts define target adult slots, target scout slots, and a minimum operating headcount
- Minimum operating headcount cannot be configured below two
- Committee may adjust a generated shift's minimum headcount with an audit record, but cannot configure away the approved local two-deep rule
- CLOSED and CLOSURE REQUIRED are shift operational states and are retained in schedule history

**Special Event Handling:**
- Special event shifts (Lot Setup, Tree Delivery) are created with higher volunteer requirements
- Special events show distinctive visual indicators in the web app (star icon, "ALL HANDS NEEDED" badge)
- Special events are highlighted in the publish notification

**Season and Draft Visibility:**
- Family Managers and Young Adult Scouts can see only published shifts
- Committee and Admin can see and manage draft schedules
- If no season is active, the web app shows the most recently completed season
- If only a draft season exists, regular users see "Schedule not yet published" without draft shift details

---

### Profile Management Rules

**Account Email and Passkeys:**
- Each authenticated user's login is tied to one or more passkeys and one claimed email account identifier
- Email addresses are normalized before comparison and must be unique across all active authenticated identities
- An email cannot be used by multiple Family Managers, Young Adult Scouts, Committee Members, or Admins
- The email identifies the account and is intended for later notification and recovery use after mailbox verification; passkeys authenticate the person
- Enrollment and ordinary sign-in do not require proving mailbox ownership
- Changing the claimed email updates the existing authenticated identity and preserves its internal user ID, roles, family memberships, and history
- The standard email-change flow requires a recent passkey assertion and uniqueness of the new address; the new address remains unverified until a later verification workflow
- After a successful email change, all sessions are revoked
- If no passkey remains, authorized recovery must revoke old credentials and sessions before reissuing an enrollment invitation for passkey re-registration on the same identity
- A Young Adult Scout may be recovered by a Family Manager; a Family Manager may be recovered by a co-manager or Admin; an Admin requires another Admin or the secured break-glass procedure
- A login email is visible only to its owner, authorized Family Managers for that profile, and Admins performing support or recovery; it is not displayed on rosters, leaderboards, or to unrelated households

**Display Name:**
- Authenticated users can edit their own display name
- Family Managers can edit person-profile names in households they manage, including managed and Young Adult Scout profiles
- Name changes are reflected throughout the web app
- Historical audit entries retain a snapshot of the display name used at the time of the activity

**Profile Photo:**
- Authenticated users can add, change, or remove their own profile photo; Family Managers can manage photos for profiles in their household
- Photos are displayed on shift rosters, leaderboards, and family views
- Removing authenticated access does not remove the photo from the preserved person profile; photo removal is a separate profile or privacy action

---

### Authenticated Access Removal Rules

**Non-Destructive Access Removal:**
When an authenticated person removes their login through the web app, the system removes access while preserving records needed for family continuity and historical reporting:

| Data Element | Action |
|--------------|--------|
| Passkeys and claimed email authentication | Removed |
| Active browser sessions | Revoked |
| Display name | Preserved (for historical reporting) |
| Profile photo | Preserved on the person profile |
| In-app inbox delivery and any future email notification preference | Removed |
| Future assignments | Preserved as family-member assignments unless separately cancelled |
| Historical attendance | Preserved (hours, check-in/out times) |
| Authenticated access status | Set to "removed" |
| Removal timestamp | Recorded |
| User ID | Preserved (referential integrity) |

**Family Manager Removal Eligibility:**

| Scenario | Can Remove Access? | Reason |
|----------|-------------|--------|
| Only manager of an active family | ❌ No | Would leave the family with no authenticated manager |
| Co-manager exists (spouse/other parent) | ✓ Yes | Other manager continues |
| Family is being deactivated/transferred by Admin | ✓ Yes | Admin handles continuity |

**Privileged Role Removal Eligibility:**
- A Committee Member may remove their own access; the Committee role is revoked and historical actions remain attributed to their profile
- An Admin may remove their own access only if another active Admin remains
- Removing one role from a multi-role identity does not remove the person's other roles unless explicitly confirmed
- Removing the entire login from a multi-role identity requires confirmation that lists every role and household access that will be revoked

**Before Removing Access (if blocked):**
The manager must take one of these actions:
1. Invite another adult who registers a passkey and accepts co-manager access
2. Ask an administrator to transfer or deactivate the family account

**Young Adult Scout Access Removal:**
- A Young Adult Scout may remove their own access without affecting the family account
- A Family Manager may revoke access for a scout in a household they manage; Admin may revoke any Young Adult Scout access
- Their profile returns to managed-scout status
- Family managers retain the profile, schedule, future assignments, and historical hours

**Managed Family Members:**
- Managed scouts have profiles without authenticated identities
- A non-manager adult normally has only a profile, but may have an authenticated identity when serving as Committee or Admin
- Removing a managed profile is a separate family-management action
- Historical hours remain preserved unless a verified permanent-deletion request requires removal or anonymization

---

### Privacy & Data Request Rules

**Relationship to Access Removal:**
- Removing authenticated access revokes login capability but preserves the person profile and history
- Access removal is not a data-export request or permanent-deletion request
- Permanent removal requires the separately verified privacy process below

**Data Export Requests:**
- Any current or former authenticated person can request a complete export of their data
- A verified parent/guardian may request data for a managed minor scout
- Requests use a separately verified process through the privacy contact
- Identity verification is required before fulfilling requests
- Export includes all personal data: profile, assignments, attendance, messages, and agreement-confirmation records associated with the requester

**Permanent Deletion Requests ("Right to be Forgotten"):**
- Any current or former authenticated person can request permanent removal or anonymization of their data; a verified parent/guardian may request it for a managed minor scout
- Requests use a separately verified process through the privacy contact
- This is MORE extensive than removing authenticated access in the web app:
  - Access removal: Revokes login and preserves historical records
  - Permanent deletion: Removes or anonymizes data to the fullest permitted extent
- Agreement-confirmation records have no special retention exemption and are deleted or anonymized with the person's other data
- Permanent deletion affects reporting (Scout Bucks totals may change)
- Admin should warn user of implications before proceeding
- Cannot be undone once completed

---

## Document History

This table records editions of the complete document. When an edition changes
use-case behavior or policy, the same change must update the affected
per-use-case and user-story revisions in `traceability/manifest.yaml`.

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | Dec 2024 | Initial use cases document extracted from architecture specification |
| 1.1 | Dec 2024 | Added Use Case 0: System Bootstrap for creating the first administrator |
| 1.2 | Dec 2024 | Added comprehensive Business Rules section consolidating multi-household rules, scheduling rules, attendance/check-in rules, walk-in coverage rules, notification rules, hours tracking/leaderboard rules, Scout Bucks calculation rules, and template/schedule rules |
| 1.3 | Dec 2024 | Added Profile Management use cases (UC 45-47): profile photo, display name editing, account deletion with soft-delete approach. Added Privacy & Data Requests use cases (UC 48-49): data export and permanent removal requests handled outside the mobile app. Added business rules for profile management, account deletion eligibility, and privacy request handling |
| 2.0 | Jul 2026 | Reframed delivery as a responsive Go/HTMX web app; replaced Apple/Google authentication with passwordless SMS phone verification; introduced family accounts with one or more family managers; made scouts managed profiles; and required signups to identify the family member filling each shift |
| 2.1 | Jul 2026 | Added optional Young Adult Scout access using the scout's own verified phone, with self-service scheduling and check-in while preserving family-manager visibility and control |
| 2.2 | Jul 2026 | Clarified that Young Adult Scout authentication links to an existing family-member profile and made troop announcements deliver by SMS to all Family Managers and Young Adult Scouts plus Groups.io |
| 2.3 | Jul 2026 | Made normalized authentication phone numbers system-wide unique so one phone number can identify only one person at a time |
| 2.4 | Jul 2026 | Defined self-service and assisted phone-number changes that preserve identity, roles, family membership, assignments, and history |
| 2.5 | Jul 2026 | Allowed authenticated adults on the current or immediately preceding shift to check in any assigned volunteer during bounded real-time windows; separated post-shift attendance adjustments from check-in/out events |
| 2.6 | Jul 2026 | Reconciled Business Rules with current use cases: defined household and family-unit boundaries, closed authentication and role-provisioning gaps, clarified scheduling and walk-in authority, aligned notifications and reminders, and made access removal, reporting, and deactivation behavior consistent |
| 2.7 | Jul 2026 | Made troop announcements explicitly visible and historically accessible in the web app in addition to SMS and Groups.io delivery |
| 2.8 | Jul 2026 | Added Use Case 4A for the web announcement view with per-user unread counts, read timestamps, and private read/unread state |
| 2.9 | Jul 2026 | Changed Scout Bucks attribution so each adult's hours are split equally among their eligible scouts rather than duplicated in full for every sibling |
| 3.0 | Jul 2026 | Added hybrid seasonal contract management with electronic signing, paper fallback, guardian consent plus scout acknowledgement, contract-first onboarding, person-level participation gating, compliance review, and retention controls |
| 3.1 | Jul 2026 | Added Use Case 55 so authorized users can view the current contract from a person profile and revisit the exact version previously completed when a material revision supersedes it |
| 3.2 | Jul 2026 | Clarified that the seasonal contract is an operational agreement to follow tree-lot conduct rules, not a legal or insurance contract; removed mandated-retention assumptions and made its evidence subject to ordinary privacy deletion |
| 3.3 | Jul 2026 | Added manual completed-season archival and confirmed deletion, including a restorable encrypted archive, checksum verification, complete removal of season-owned data, and preservation of shared cross-season records |
| 3.4 | Jul 2026 | Replaced stored and signed contract versions with one public Google Doc link per season plus a per-person boolean confirmation; removed document storage, signatures, guardian-consent records, paper fallback, and scans |
| 3.5 | Jul 2026 | Simplified season archival to a normal checksummed ZIP stored in troop-owned access-controlled storage; removed application-managed archive encryption and key custody |
| 3.6 | Jul 2026 | Added passphrase-based age encryption around season ZIP archives; the application never stores the passphrase, and restoration requires it |
| 3.7 | Jul 2026 | Made Groups.io announcement delivery an optional deployment integration that is disabled by default; in-app and SMS announcement delivery remain required |
| 3.8 | Jul 2026 | Separated provisional Scout Bucks credited hours from finalized dollar awards; added Treasurer-entered distributable profit, exact proportional cent allocation, immutable settlement revisions, and CSV export without an ongoing balance ledger |
| 3.9 | Jul 2026 | Added minimum operating headcount, the separately approved local two-deep coverage rule with its sole-adult family exception, troop-wide critical coverage alerts, actual-attendance enforcement, and audited shift closure/reopening workflows |
| 3.10 | Jul 2026 | Removed the sole-adult family exception so every operating shift requires at least two adults regardless of family relationships or whether scouts are present |
| 3.11 | Jul 2026 | Explicitly documented Scouting America's national two-deep policy and clarified that local adult counts do not verify registered-leader, age, training, or eligibility requirements |
| 3.12 | Jul 2026 | Replaced SMS phone authentication with passkeys and claimed email account identifiers; deferred email mailbox verification until notifications or recovery need it; switched enrollment and recovery to invitation links/QR codes; required browser JavaScript for WebAuthn while continuing to forbid heavy client-side frameworks; left operational SMS notification use cases in place pending a later move to email and Groups.io |
| 3.13 | Jul 2026 | Removed operational SMS/phone notification delivery; made the per-user in-app inbox the required notification channel with private read state; reserved Groups.io for troop-wide messages when enabled; kept direct and family-scoped messages in-app only; documented future opted-in delivery to verified email addresses |

---

**End of Document**
