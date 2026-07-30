# US-029 — View weekly staffing levels

- **Epic:** Staffing Intelligence & Shift Operations
- **Source use cases:** [UC-41](../../use-cases.md#use-case-41-committee-views-week-schedule-with-staffing-levels)
- **Primary actor:** Committee Member

As a Committee Member,
I want to view a week of shifts with staffing levels,
so that I can quickly identify where volunteer coverage is needed.

## Scope

Present a navigable weekly summary and per-shift staffing details based on published shift requirements and active assignments.

## Preconditions

- A season schedule has been published.
- Published shifts define adult and scout targets and a minimum operating headcount.
- Active assignments are available for projected staffing calculations.

## Acceptance criteria

1. **Given** a published week containing shifts, **when** the Committee Member opens Week View, **then** the system shows total shifts and counts of fully staffed, understaffed, and critical shifts.
2. **Given** a shift in the selected week, **when** it is displayed, **then** the view shows its time, location, special-event marker when applicable, signed-up roster, adult and scout fill rates, total-person minimum, and scheduled local two-deep result.
3. **Given** current projected staffing, **when** the system labels a shift, **then** it uses FULL, OK, LOW, or CRITICAL according to target staffing, minimum headcount, and scheduled local two-deep coverage.
4. **Given** an operational state requiring action, **when** the shift is displayed, **then** CLOSURE REQUIRED or CLOSED is shown distinctly from projected FULL, OK, LOW, and CRITICAL levels.
5. **Given** another valid season week, **when** the Committee Member uses the arrows or week picker, **then** the requested week's staffing summary and shifts are shown.
6. **Given** a displayed shift, **when** the Committee Member opens it, **then** full shift details are available.

## Business rules

- Projected coverage uses active assignments; it does not represent who is physically present.
- FULL requires all target slots plus projected minimum-headcount and local two-deep compliance.
- OK and LOW are safe-to-operate projections below full target staffing; CRITICAL means at least one projected operating requirement fails.
- Minimum operating headcount and local two-deep coverage are separate requirements.
- Actual attendance controls whether the lot may operate at shift time.

## Dependencies

- US-019
- US-023
- US-024

## Out of scope

- Recording attendance or evaluating actual checked-in coverage
- Sending staffing messages
- Closing or reopening a shift
- Verifying national registered-leader, age, training, or eligibility requirements
