# US-033 — Close or reopen unsafe shift

- **Epic:** Staffing Intelligence & Shift Operations
- **Source use cases:** [UC-59](../../use-cases.md#use-case-59-committee-closes-a-shift-for-insufficient-coverage)
- **Primary actor:** Committee Member or Admin

As a Committee Member or Admin,
I want to close an unsafe shift or safely reopen a pre-start closure,
so that the operating decision is explicit, communicated, and audited.

## Scope

Handle CLOSURE REQUIRED, recorded closure, closure effects, troop-wide notices, and the narrowly allowed pre-start reopening workflow.

## Preconditions

- The actor has Committee or Admin authority.
- A critical-alert deadline has passed without safe projected coverage, the scheduled start has arrived without safe coverage, or actual attendance has become unsafe.

## Acceptance criteria

1. **Given** an unresolved deadline or unsafe actual coverage, **when** the shift is evaluated, **then** it becomes CLOSURE REQUIRED and shows the unresolved minimum-headcount or local two-deep rule.
2. **Given** a CLOSURE REQUIRED shift, **when** the actor records a reason and confirms closure, **then** the shift becomes CLOSED and the closure action is audited.
3. **Given** a shift becomes CLOSED, **when** participation is attempted, **then** new signups, check-ins, and walk-in additions are disabled while checkout remains available for existing open attendance.
4. **Given** closure is recorded, **when** closure effects are committed, **then** assignments remain in audit history, are marked cancelled by closure, and one canonical troop-wide in-app notice is queued for every active Family Manager and Young Adult Scout; Groups.io is also queued when enabled.
5. **Given** a shift was closed before its start and now satisfies projected minimum headcount and local two-deep coverage, **when** Committee/Admin records an audited reopening, **then** the shift reopens and an update is sent to the same recipient set.
6. **Given** operations had begun before closure, **when** reopening is attempted, **then** the system rejects retroactive reopening.
7. **Given** a shift is only CLOSURE REQUIRED, **when** no actor has recorded closure, **then** the system does not represent the physical lot as already closed.

## Business rules

- CLOSURE REQUIRED is an operational warning; CLOSED is an explicit human-recorded decision.
- Reopening is allowed only before shift start and only after both projected operating requirements are satisfied.
- A shift closed after operations begin cannot be reopened retroactively.
- Attendance events and coverage transitions remain immutable; closure does not erase them.
- Notice deliveries are independent and idempotent, and external calls run outside the transaction.

## Dependencies

- US-030
- US-031
- US-032

## Out of scope

- Deciding whether an adult satisfies national-policy leader requirements
- Editing attendance events
- Reopening a shift after operations began
- Physically operating or securing the lot
