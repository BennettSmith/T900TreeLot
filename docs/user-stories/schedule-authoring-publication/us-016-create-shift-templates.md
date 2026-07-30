# US-016 Create Shift Templates

- **Epic:** Schedule Authoring and Publication
- **Source use cases:** [UC-20](../../use-cases.md)
- **Primary actor:** Committee Member

**As a** Committee Member,  
**I want** to create reusable shift templates,  
**so that** standard season schedules can be generated consistently.

## Scope

Create, update, review, and deactivate reusable weekday, weekend, Friday, and special-event shift patterns.

## Preconditions

- The Committee Member is authenticated and authorized to manage shift templates.

## Acceptance criteria

1. **Given** template management, **when** the Committee Member creates a template, **then** they can define its name, type, applicable weekdays or specific-event use, shift times, adult targets, scout targets, and minimum operating headcount.
2. **Given** a minimum operating headcount below two, **when** the template is submitted, **then** the system rejects it.
3. **Given** an existing template, **when** it is updated, **then** only schedules generated after the update use the new values.
4. **Given** a template is deactivated, **when** historical references are reviewed, **then** the template remains available for that history but is not selected for new generation.

## Business rules

- Generated shifts snapshot target adult/scout counts and minimum operating headcount.
- Template changes never alter previously generated schedules.
- A template cannot configure away the separately approved local two-deep rule.

## Dependencies

- US-002 — authenticate and authorize the Committee Member.

## Out of scope

- Generating a season schedule.
- Editing individual generated shifts.
- Publishing shifts.
