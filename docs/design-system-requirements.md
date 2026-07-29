# Troop 900 Tree Lot Shift Scheduler

## Logical UI/UX Design System Requirements

**Status:** POC design brief  
**Date:** July 2026  
**Related documents:** [Use Cases](use-cases.md), [System Architecture](architecture.md)

## 1. Purpose

This document defines the logical elements, behavior, states, accessibility requirements, and representative screens for the Tree Lot Scheduler's UI/UX design system.

It is intentionally independent of a specific visual style. Multiple agents should be able to use this brief to create substantially different proofs of concept while solving the same interaction and information-design problems.

The POCs are exploratory artifacts, not production implementations. Their purpose is to compare visual direction, hierarchy, navigation, density, responsiveness, and component treatment before the production design system is built with Go templates, Tailwind CSS, and progressively enhanced HTML.

## 2. Goals

The design system must:

- Make common family workflows quick and understandable on a phone
- Support on-site attendance actions under time pressure
- Make permission boundaries and unavailable actions understandable
- Present complex household relationships without exposing unrelated private information
- Give Committee and Admin users efficient data-dense operational views
- Use one consistent vocabulary for status, severity, actions, and feedback
- Work across full-page navigation and HTMX-enhanced partial updates
- Meet WCAG 2.2 Level AA for core workflows
- Remain small enough to implement and maintain as reusable Go template components

## 3. Non-goals

The logical design system does not prescribe:

- A final brand palette
- A typeface
- A logo or illustration style
- Exact spacing, radius, shadow, or motion values
- A dark theme
- A desktop-first information architecture
- A JavaScript component framework
- The final visual treatment of any POC

The POCs do not need to implement authentication, persistence, SMS, concurrency, provider integrations, or production authorization. They should simulate those outcomes with representative state.

## 4. Product context

The scheduler serves approximately 40–50 volunteers across 15–20 households during a four-to-six-week fundraiser.

The primary usage contexts are:

- A parent using a phone to coordinate family shifts
- A scout using a phone to manage their own participation
- An adult using a phone at the tree lot to check volunteers in or out
- A Committee member monitoring staffing on phone, tablet, or desktop
- An Admin performing lower-frequency setup, privacy, or season-maintenance work
- A Treasurer reviewing a data-dense end-of-season Scout Bucks calculation

The interface must feel useful and complete at phone width. Desktop layouts may increase information density but must not be the only workable form.

## 5. Canonical actors

### Family Manager

An authenticated parent or guardian who manages a household and its people. Their UI includes family-wide schedules, agreement confirmations, signup, attendance history, statistics, invitations, and profile management.

### Young Adult Scout

An authenticated scout with limited self-service access. Their UI is personal rather than household-wide. They can act only for themselves, while Family Managers retain management visibility.

### Managed person

A scout or adult profile without login access. A Family Manager may facilitate permitted actions for this person. UI copy must identify both the selected participant and the authenticated actor where attribution matters.

### Committee Member

An operational user who manages schedules, staffing, announcements, attendance, walk-ins, statistics, and corrections.

### Treasurer

A Committee Member with permission to finalize and export Scout Bucks awards.

### Admin

A privileged user who manages onboarding, roles, households, seasonal agreement links, privacy actions, and season archival/deletion.

## 6. Design principles

### Mobile is operational, not secondary

Phone layouts must support every core workflow. Check-in, checkout, signup, cancellation, roster management, and agreement confirmation must not require desktop access.

### Show the next valid action

The UI should emphasize what the current actor can do now. It should not make users infer actions from raw status data.

### Explain unavailable actions

When useful and safe, disabled or omitted actions should be accompanied by a reason such as:

- Shift is full
- This person has not confirmed the seasonal agreement
- No slot matches this person's role
- This assignment belongs to another household
- Check-in is not open yet
- This household is inactive

Authentication and phone-identity errors remain generic to avoid account enumeration.

### Preserve actor and target clarity

Every action performed for another person must clearly show:

- Who is acting
- Who the action affects
- Which household owns the action when relevant

### Status must be redundant

Color alone never communicates status. Status includes a text label and, where useful, an icon or shape.

### Destructive actions are proportional

Routine cancellation needs a concise confirmation. Account removal, privacy deletion, agreement-link replacement, and season deletion need progressively stronger warnings and confirmation.

### Progressive enhancement is invisible

HTMX may improve speed, but full-page and partial-page interactions must use the same component language and provide the same result.

### Financial language is precise

During a season the UI displays provisional Scout Bucks credited hours, not estimated dollars. Dollar awards appear only after Treasurer finalization.

## 7. Information architecture

POCs may explore different navigation models, but they must preserve the following logical destinations.

### Shared destinations

- Schedule / Week View
- Announcements, with unread count
- Seasonal Agreement
- Profile
- Account security
- Notification preferences

### Family Manager destinations

- Family dashboard
- Family schedule
- Find shifts
- Family members
- Attendance history
- Statistics and leaderboards
- Household settings

### Young Adult Scout destinations

- My schedule
- Find shifts
- My statistics
- Seasonal Agreement
- Announcements
- Profile and account

The Young Adult Scout UI must not reveal private household-member details merely because the scout is linked to those households.

### Committee destinations

- Week View
- Staffing Alerts
- Shift management
- Shift templates
- Draft schedule
- Attendance review
- Announcements
- Season statistics
- Agreement status
- Reports

### Admin destinations

Admin receives Committee destinations plus:

- Household administration
- Invitations and role management
- Season configuration
- Agreement-link configuration
- Privacy operations
- Season archive and deletion

### Persistent context

The page shell must provide a place for:

- Current season
- Current week or date range
- Unread announcement count
- Agreement-required alert
- Staffing-alert count for privileged users
- Selected household when an actor manages more than one
- User menu and role context

## 8. Foundations

Each POC should define concrete visual values for the following logical tokens.

### Color roles

- Canvas
- Primary surface
- Raised surface
- Muted surface
- Primary text
- Secondary text
- Disabled text
- Border
- Strong border
- Primary action
- Focus
- Informational
- Successful / complete
- Warning / needs attention
- Critical / destructive
- Special event

Token names should describe meaning rather than a literal color. For example, use `status-critical`, not `red`.

### Typography roles

- Page title
- Section heading
- Subsection heading
- Body
- Emphasized body
- Control label
- Supporting text
- Caption / metadata
- Numeric statistic
- Monospaced code or checksum

### Space and size roles

- Inline spacing scale
- Stack spacing scale
- Section spacing
- Page gutters
- Control heights
- Icon sizes
- Touch-target minimums
- Content-width limits
- Data-density modes where needed

Primary mobile actions should provide touch targets of approximately 44 by 44 CSS pixels or larger.

### Shape and depth roles

- Control radius
- Card radius
- Dialog radius
- Standard border
- Selected border
- Raised shadow
- Overlay shadow

### Motion roles

- Immediate feedback
- Standard transition
- Content entrance or swap
- Loading indication

All nonessential motion must respect `prefers-reduced-motion`.

### Responsive ranges

Each POC must define behavior for:

- Narrow phone
- Large phone
- Tablet
- Desktop

Breakpoints may differ between POCs. Component behavior matters more than exact breakpoint values.

## 9. Primitive components

Every primitive supports default, hover where applicable, focus, active, disabled, loading, validation-error, and success states as relevant.

### Button

Required variants:

- Primary
- Secondary
- Quiet
- Destructive
- Icon with accessible name
- Full-width mobile action

Required behaviors:

- Clear loading state without changing width unexpectedly
- Disabled reason available when needed
- Destructive styling not reused for ordinary cancellation

### Link

Required variants:

- Inline
- Navigation
- Standalone action
- External link
- Destructive text action

External links must identify that they leave the scheduler when that context is not obvious.

### Text field

Supports label, hint, required/optional state, entered value, validation message, disabled state, and appropriate autocomplete/input mode.

Specialized uses include:

- Display name
- Season name
- Agreement URL
- Confirmation phrase
- Archive passphrase

### Phone-number field

Uses telephone input behavior, explains expected formatting without implying that formatting controls identity, and supports generic security errors.

### OTP field

Supports mobile numeric entry, paste/autofill, resend timing, expiration, loading, and generic invalid-code feedback.

### Money field

Used for the Scout Bucks distributable pool.

It must:

- Clearly indicate currency
- Accept dollars and cents
- Avoid ambiguous formatting
- Present the normalized value before final confirmation

### Textarea

Used for:

- Announcement body
- Attendance-adjustment reason
- Walk-in note
- No-show note

Supports character guidance, validation, and error association.

### Checkbox

Supports a visible label, supporting description, validation, and a large touch target.

The seasonal agreement confirmation checkbox uses the exact intent:

> I have read and agree to the tree-lot rules of conduct.

### Radio group

Used for mutually exclusive choices such as draft versus immediate publication. Uses `fieldset` and `legend` semantics.

### Toggle

Used for notification preferences and optional settings. It must not replace a checkbox when immediate state change would be surprising.

### Select and picker

Supports:

- Week selection
- Household selection
- Role/status filters
- Season selection

### Search field

Used for finding a walk-in participant or narrowing administrative lists. Search results must expose role and agreement eligibility without revealing unrelated private data.

### Badge

Required semantic families:

- Neutral
- Informational
- Complete
- Warning
- Critical
- Special event
- Provisional

Badges do not become buttons unless they have button semantics and affordance.

### Alert

Required variants:

- Informational
- Success
- Warning
- Critical
- Blocking

Alerts may include one primary action and optional supporting actions.

### Inline validation message

Appears next to the relevant control and is included in a form-level error summary for longer forms.

### Confirmation dialog

Used only where an in-context confirmation is appropriate. It has:

- Clear action and consequence
- Safe default focus
- Cancel action
- No reliance on clicking outside to escape
- Full-page fallback when JavaScript is unavailable

### Progress indicator

Includes:

- Determinate progress bar
- Indeterminate loading indicator
- Step progress for onboarding or destructive workflows

Progress bars require accessible names and values.

### Avatar

Supports:

- Profile photo
- Initials fallback
- Generic fallback
- Optional role marker

### Photo input

Supports camera/library selection, preview, crop, replace, remove, upload progress, and failure. This is one of the few components permitted to require browser JavaScript for its enhanced experience.

### Card

Provides a consistent container for grouped content. Cards must not make nested interactive regions ambiguous.

### List and list row

Supports:

- Primary label
- Secondary description
- Metadata
- Status
- Leading avatar or icon
- Trailing action
- Selected/current state

### Data table

Provides:

- Caption or visible heading
- Column headers
- Numeric alignment
- Sort indication when sortable
- Row actions
- Empty state
- Responsive alternative when columns cannot fit

### Tabs or segmented control

Used only for sibling views, such as individual versus family leaderboard. Tabs require keyboard behavior and clear selected state.

### Pagination

Supports current page, next/previous, result count, and preservation of current filters.

### Empty state

Includes:

- What is empty
- Whether this is expected
- The next valid action, if any

### Definition list

Used for compact label/value summaries such as shift details, archive metadata, and settlement facts.

### Timer

Displays elapsed shift time. It must not be the only source of authoritative attendance time and should not announce every tick to screen readers.

### Copy action

Used for household link codes and shareable signup links. Provides explicit copied feedback and a non-JavaScript selectable-text fallback.

## 10. Domain components

### Page shell

Contains application identity, role-appropriate navigation, current season context, alert indicators, main content, and user menu.

POCs may explore top navigation, side navigation, bottom navigation, or hybrids.

### Week navigator

Contains:

- Current date range
- Previous and next controls
- Week picker
- Shift count per week
- Disabled boundary states

### Shift card

Required content:

- Date and time
- Location
- Shift or event name
- Special-event status
- Adult and scout staffing
- Current signup state
- Most relevant action

Required variants:

- Available
- Needs help
- Critical
- Full
- Signed up
- Draft, privileged view only
- Special event
- In progress
- Completed

### Staffing meter

Shows adult and scout staffing separately and the overall category:

- FULL
- OK
- LOW
- CRITICAL

It includes counts, not only percentages or color.

### Special-event treatment

Communicates high priority with a text label such as `ALL HANDS NEEDED`, an icon, and visual emphasis.

### Shift detail header

Combines shift identity, time, location, staffing, lifecycle, notes, and contextual actions.

### Person selector

Used when a Family Manager signs up another person.

It:

- Groups adults and scouts
- Shows slot compatibility
- Shows agreement eligibility
- Marks already-assigned people
- Disables invalid choices with explanations
- Makes the selected person unmistakable before submission

Young Adult Scout self-service does not display this selector.

### Assignment row

Shows:

- Assigned person
- Adult or scout slot
- Scheduled versus walk-in
- Originating household when relevant
- Attendance state
- Cancellation authority

### Agreement status panel

Shows every relevant household member with:

- Confirmed or Not Confirmed
- Confirmation date when available
- Open Agreement action
- Confirm action when authorized
- Facilitated-confirmation context where relevant

It does not imply that the scheduler stores or verifies the linked Google Doc.

### Agreement-link configuration

Admin component containing:

- Display title
- Approved Google Docs URL
- External preview/open action
- Current confirmation count
- Warning that replacing the URL resets every confirmation
- Explicit replacement confirmation

### Check-in action panel

Prominently shows the current valid action:

- Check in now
- Check out now
- Already checked in
- Already checked out
- Check-in not yet open
- Check-in window closed
- Checkout not yet open

### Shift roster

Rows show:

- Person and role
- Scheduled or walk-in
- Pending, checked-in, checked-out, no-show, or adjusted
- Time and actor attribution where relevant
- Actions permitted to the current actor

The phone layout must remain fast to scan during arrivals.

### Attendance row

Distinguishes:

- Raw real-time check-in/out events
- Who performed each event
- Calculated hours
- Approved adjustment
- Adjustment reason and actor

An adjustment never visually masquerades as an original timestamp.

### Walk-in panel

Contains:

- Participant search
- Person and role confirmation
- Agreement eligibility
- Optional note
- Explanation that creation immediately checks the person in
- Over-capacity result when applicable

### No-show marker

Provides a clear status and optional follow-up note without erasing the original assignment.

### Announcement item

Includes:

- Title
- Priority
- Author
- Publication time
- Read/unread state

Unread treatment must be distinguishable without color alone.

### Announcement composer

Contains:

- Title
- Body
- Priority
- Recipient counts
- Required SMS channel
- Optional Groups.io channel only when enabled
- Publish confirmation

### Delivery status summary

Shows web publication, aggregate SMS outcomes, and optional Groups.io status independently. Authorized users can retry failed delivery without duplicating successful channels.

It never exposes recipients' phone numbers to ordinary users.

### Family member row

Combines avatar, display name, person type, household relationship, authenticated/managed status, agreement status, and contextual actions.

### Household context indicator

Identifies the selected household and assignment ownership in multi-household scenarios without suggesting that shared scouts are duplicated people.

### Profile header

Contains avatar, display name, role/access summary, agreement status, and permitted profile actions.

### Statistics summary

Supports:

- Total hours
- Shift count
- Average hours
- Rank
- Family total
- Provisional Scout Bucks credited hours

### Leaderboard row

Contains rank, person/family name, hours, shift count, and a current-user/current-family treatment.

### Scout Bucks credited-hours row

Contains:

- Scout's own hours
- Allocated adult hours
- Total credited hours
- `Provisional` label before finalization
- Drill-down into allocation sources

It must not display estimated dollars.

### Scout Bucks settlement panel

Contains:

- Distributable profit input
- Finalized credited-hour total
- Informational effective rate
- Per-scout proposed awards
- Exact-pool reconciliation
- Rounding explanation
- Finalize action
- Current revision and prior-revision history
- CSV export

The finalization confirmation must state that the scheduler will not maintain balances or redemptions.

### Season archive summary

Shows archive contents, record counts, format version, generation status, and checksum.

### Archive passphrase form

Requires passphrase entry and confirmation. Copy must explain:

- The passphrase is not stored
- It must be saved separately
- Losing it makes the archive unrecoverable

The passphrase must never be redisplayed.

### Season deletion confirmation

Shows:

- Season identity
- Data categories and record counts
- Verified archive status and checksum
- Irreversibility warning
- SMS re-authentication
- Typed season-name confirmation
- Final destructive action

## 11. Status vocabulary

The following labels are canonical. POCs may alter visual treatment but not meaning.

### Staffing

- FULL
- OK
- LOW
- CRITICAL

### Seasonal agreement

- Confirmed
- Not Confirmed

### Shift lifecycle

- Draft
- Published
- In Progress
- Completed
- Cancelled
- Special Event

### Signup

- Available
- Full
- Signed Up
- Ineligible
- Household Inactive

### Attendance

- Pending
- Checked In
- Checked Out
- No Show
- Walk-In
- Adjusted

### Announcement

- Unread
- Read
- Informational priority
- Important priority
- Urgent priority

### Delivery

- Pending
- Sent
- Delivered
- Failed
- Retrying
- Disabled, for optional channels

### Scout Bucks

- Provisional Credited Hours
- Ready to Finalize
- Finalized
- Corrected Revision

### Season archive

- Not Generated
- Generating
- Ready
- Checksum Verified
- Deleted from Live System

## 12. Page templates and patterns

### Welcome and OTP sign-in

Supports phone entry, generic submission feedback, OTP verification, resend, expiration, and role-appropriate landing.

### Invitation onboarding

Logical sequence:

1. Invitation validation
2. Phone verification
3. Profile completion
4. Household setup or access acceptance
5. Family-member setup
6. Agreement Center

### Family dashboard

Prioritizes:

- Agreement actions blocking participation
- Upcoming family shifts
- Shifts that need help
- Unread announcements
- Family-hour summary

### Young Adult Scout home

Prioritizes:

- Next personal shift
- Current check-in/out action
- Find a shift
- Agreement status
- Own hours and provisional credited hours

### Week View

Supports day grouping, week navigation, staffing visibility, special events, signup state, and role-appropriate actions.

POCs should explore at least two possible phone treatments, such as a stacked day list and compact cards.

### Shift detail

Combines shift facts, staffing, notes, roster visibility, signup/cancel actions, and live attendance behavior.

### Family schedule

Supports all household members, filtering by person, assignment ownership, attendance status, and permitted cancellation.

### Live roster

Optimized for rapid phone use with large, unambiguous attendance actions and clear current state.

### Committee operations view

Provides a higher-density week view with staffing summaries, alerts, shift management, and targeted actions.

### Agreement Center

Displays the public agreement link and per-person confirmation. It handles self-confirmation and facilitated confirmation without signatures or document uploads.

### Announcements

Includes list, detail, read-state management, compose, delivery preview, delivery status, and retry.

### Attendance review

Separates immutable real-time events from corrections and no-show disposition.

### Reports and Scout Bucks

Separates provisional credited-hours reporting from Treasurer-only dollar finalization and export.

### Profile and account

Includes display name, profile photo, phone change, session/security information, role continuity, and access removal.

### Season maintenance

Uses a deliberate two-stage flow:

1. Generate, download, and verify encrypted archive
2. Re-authenticate and delete the live season

## 13. Responsive behavior

### Phone

- Single primary content column
- Primary action remains easy to reach
- Tables transform into structured rows/cards where necessary
- Shift and roster status is visible without horizontal scrolling
- Filters may collapse behind a clearly labeled control
- Destructive confirmations remain full-context, not cramped dialogs

### Tablet

- May use split views for schedule/detail or filter/results
- Preserves touch-sized controls
- Avoids desktop-only hover interactions

### Desktop

- May expose persistent navigation
- May use denser tables and side-by-side summaries
- Must preserve the same action labels and status vocabulary as mobile

### Cross-size requirements

- No core action depends on hover
- Content remains usable at 200% text zoom
- Long names, translated browser-generated dates, and large numeric values do not overlap controls
- Safe-area insets are respected on modern phones
- Sticky actions must not obscure validation or page content

## 14. Accessibility acceptance criteria

Every POC should demonstrate the intended treatment of these requirements, even if not fully wired to an automated checker.

### Structure

- One clear page-level heading
- Logical heading order
- Landmark regions
- Semantic lists, tables, forms, and buttons

### Keyboard

- All controls reachable and operable
- Visible focus at all times
- Logical focus order
- No keyboard traps
- Dialog focus enters, remains within, and returns appropriately

### Forms

- Visible labels
- Programmatic label association
- Instructions before they are needed
- Errors associated with controls
- Error summary for long or destructive forms
- Required state not communicated only by color

### Dynamic updates

Signup results, roster changes, unread counts, delivery retry outcomes, and validation messages provide appropriate screen-reader announcements without excessive interruption.

### Visual

- WCAG 2.2 AA contrast
- Status not color-only
- Text reflows at zoom
- Focus is not obscured
- Touch targets are adequately sized
- Reduced-motion preference honored

### Screen readers

Production acceptance includes VoiceOver with Safari and TalkBack with Chrome. POCs should use semantic structures that make those outcomes plausible.

## 15. Progressive-enhancement requirements

Production components are server-rendered Go templates.

Each core action must have:

- A real link or form
- A complete server response or redirect
- A usable full-page result
- Optional HTMX enhancement using the same logical component

POCs may use a rapid-prototyping framework, but they must not present interactions that require a SPA-only architecture in production.

At least one POC path should illustrate how the same form works as:

- A normal submit and redirect
- An enhanced inline update

## 16. Content design

### Voice

Use language that is:

- Clear
- Calm
- Direct
- Respectful of adults and youth
- Helpful under time pressure
- Specific about consequences

### Canonical terms

Use:

- Person
- Adult
- Scout
- Family Manager
- Young Adult Scout
- Household
- Family unit, only in reporting contexts
- Seasonal Agreement
- Confirmed / Not Confirmed
- Credited hours
- Finalized Scout Bucks award

Avoid:

- Shared family login
- Child account
- Legal contract
- Signature
- Estimated Scout Bucks balance
- Backdated check-in

### Action labels

Prefer explicit labels:

- Sign Up Alex
- Cancel Alex's Assignment
- Check In Jordan
- Check Out Jordan
- Add Walk-In
- Open Agreement
- Confirm Agreement
- Publish Schedule
- Finalize Scout Bucks Awards
- Generate Season Archive
- Delete Archived Season

Avoid ambiguous labels such as `Submit`, `OK`, or `Continue` when the actual action can be named.

### Dates and times

- Display in the configured tree-lot time zone
- Include day of week for shifts
- Avoid ambiguous numeric-only dates
- Show time-zone context in administrative or archive views where needed

### Currency

- Use dollars and cents for finalized awards
- Label effective rate as informational
- Use `Provisional` for credited hours before finalization
- Never imply that the scheduler holds or spends a scout's balance

## 17. Privacy and security presentation

- Do not show phone numbers on rosters, leaderboards, or unrelated household views
- Do not expose per-recipient delivery details to ordinary recipients
- Do not reveal whether an unknown phone number has an account
- Do not display archive passphrases after submission
- Do not imply that the scheduler stores the public Google Doc
- Make acting user and affected person visible for facilitated actions
- Explain session revocation after phone-number changes
- Distinguish access removal from permanent personal-data deletion

## 18. Required POC deliverables

Each design POC should include:

### Design-system gallery

Show:

- Foundation tokens
- Typography hierarchy
- Buttons and links
- Form controls
- Badges and alerts
- Cards and list rows
- Tables and mobile alternatives
- Navigation
- Loading, empty, validation, success, warning, and error states
- Domain components in their principal variants

### Required representative screens

#### 1. Mobile Family Dashboard

State:

- One household member is Not Confirmed
- Two upcoming assignments
- One critical shift seeking help
- Two unread announcements

Demonstrates navigation, hierarchy, alerts, shift cards, and family context.

#### 2. Mobile Agreement Center

State:

- Authenticated parent is Confirmed
- Managed scout is Not Confirmed
- Another adult is Confirmed through facilitated confirmation

Demonstrates external document link, status, self/facilitated confirmation, and blocking state.

#### 3. Responsive Week View

State:

- Mixture of FULL, OK, LOW, and CRITICAL shifts
- One special event
- One shift already assigned to the current family

Demonstrates week navigation, staffing, scanability, and responsive transformation.

#### 4. Shift Detail and Signup

State:

- One adult slot and one scout slot remain
- Person selector contains eligible, wrong-role, unconfirmed, and already-assigned people
- Alternate state shows a concurrent submission resulting in Shift Full

Demonstrates selection clarity, validation, capacity feedback, and special-event treatment.

#### 5. Mobile Live Roster

State:

- Pending, checked-in, checked-out, walk-in, and no-show people
- Current adult may check in another volunteer
- Add Walk-In action available

Demonstrates time-critical actions, attendance states, actor attribution, and touch usability.

#### 6. Committee Staffing View

State:

- Week-level summary
- Critical and low staffing alerts
- Targeted reminder and share-link actions

Demonstrates privileged navigation, data density, filters, progress, and prioritization.

#### 7. Announcement Compose and Delivery

State:

- Required web and SMS delivery
- Groups.io disabled by default
- Alternate optional-channel-enabled and failed states

Demonstrates recipient preview, optional configuration, independent delivery status, and retry.

#### 8. Treasurer Scout Bucks Finalization

State:

- Provisional credited hours
- Entered distributable pool
- Proposed awards with a rounding remainder
- Exact reconciliation to the pool
- Existing finalized revision

Demonstrates dense numeric content, currency input, irreversible finalization, revision history, and export.

### Required viewport presentations

Show at least:

- 390-pixel-wide phone
- Tablet-width layout
- 1440-pixel-wide desktop

Not every screen needs all three, but the Week View, Shift Detail, and one data-dense privileged screen must demonstrate responsive behavior.

### State controls

The POC should make important variants easy to inspect through:

- A component-gallery route
- Query parameters
- A local state switcher
- Separate clearly named examples

Reviewers should not need to edit source code to see disabled, loading, error, empty, critical, and success states.

## 19. Optional POC screens

Agents may additionally explore:

- OTP sign-in
- Invitation onboarding
- Multi-household assignment ownership
- Attendance adjustment
- Leaderboard
- Profile photo and account security
- Draft schedule publication
- Season archive and deletion

## 20. POC evaluation criteria

POCs will be compared on:

- Mobile usability
- Information hierarchy
- Speed of understanding
- Clarity of role and household context
- Clarity of status and next action
- Quality of time-critical attendance interactions
- Handling of dense Committee and Treasurer information
- Accessibility plausibility
- Responsive behavior
- Consistency across components
- Strength and distinctiveness of visual direction
- Fit with server-rendered progressive enhancement
- Ease of translating the concepts into Go templates and Tailwind CSS

A visually attractive POC that obscures permissions, status, actor/target relationships, or mobile actions does not satisfy this brief.

## 21. Visual directions intentionally open

Agents are encouraged to explore different answers for:

- Troop and tree-lot brand expression
- Seasonal versus year-round visual tone
- Warm, civic, outdoors, utilitarian, or editorial aesthetics
- Typeface and type scale
- Color palette
- Density
- Radius and shadow
- Icon family
- Illustration and photography
- Top, side, bottom, or hybrid navigation
- Calendar-like versus list-like Week View
- Table versus card treatment on mobile
- Loading and success feedback
- Motion within reduced-motion constraints
- Avatar fallback style
- Progress-meter design

These choices should differ between POCs while the logical elements, states, terminology, accessibility expectations, and required scenarios remain stable.
