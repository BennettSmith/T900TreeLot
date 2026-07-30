# US-031 — Enforce actual local two-deep coverage

- **Epic:** Staffing Intelligence & Shift Operations
- **Source use cases:** [UC-58](../../use-cases.md#use-case-58-system-enforces-the-local-two-deep-coverage-rule-during-a-shift)
- **Primary actor:** System

As the System,
I want to evaluate actual local two-deep coverage from checked-in people,
so that the tree lot operates only while at least two adults are physically present.

## Scope

Specify the actual-coverage policy evaluated by real-time check-in, checkout, and walk-in workflows. Attendance stories exercise this evaluator; this story does not create attendance events.

## Preconditions

- The shift belongs to a published schedule and defines a minimum operating headcount.
- Assigned people have adult or scout classifications.
- Attendance commands can provide the current immutable set of checked-in people.

## Acceptance criteria

1. **Given** fewer than two adults are checked in, **when** an adult check-in would improve coverage, **then** the system allows the otherwise-valid adult check-in and reevaluates actual coverage.
2. **Given** fewer than two adults are checked in, **when** a scout check-in or scout walk-in is attempted, **then** the system rejects it and creates no attendance or walk-in record.
3. **Given** at least two adults are checked in and minimum headcount is satisfied, **when** actual coverage is evaluated, **then** the lot is locally safe to operate.
4. **Given** an adult checkout leaves fewer than two checked-in adults, **when** checkout is recorded, **then** checkout remains successful, actual coverage becomes noncompliant, operations must stop, and Committee/Admin receives an urgent operational alert.
5. **Given** actual coverage changes between compliant and noncompliant, **when** the transition is recorded, **then** the system preserves the source attendance events and records the transition and actors without rewriting history.
6. **Given** a Young Adult Scout or an authenticated Committee/Admin user, **when** adult coverage is counted, **then** only the person's recorded adult classification and physical check-in determine whether they count.

## Business rules

- Actual coverage uses open check-in records; projected coverage uses assignments.
- Local two-deep coverage always requires at least two checked-in adults, regardless of relationships or whether scouts are present.
- Minimum operating headcount is a separate requirement and must also be satisfied.
- Adult checkout is never blocked by the coverage rule.
- Attendance events are immutable and use server time.
- The local evaluator does not verify Scouting America registration, age, training, or leader eligibility; Committee remains responsible for national-policy compliance.

## Dependencies

- US-007
- US-019

## Out of scope

- Creating check-in, checkout, or walk-in records
- Projected staffing classifications
- Human confirmation of national-policy compliance
- Recording the final decision to close or reopen a shift
