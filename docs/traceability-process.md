# Requirements and delivery traceability

## Purpose

Traceability answers four questions without replacing the requirements or Git:

1. Which revision of a use case or user story is currently authoritative?
2. What is its delivery status?
3. Which executable specification verifies that exact revision?
4. Which merged pull request and Git SHA delivered it?

The authoritative workflow and policy remain in
[`docs/use-cases.md`](use-cases.md). User stories remain delivery slices, and
Git remains the immutable change history. Traceability connects those sources;
it does not duplicate their prose.

## Identifiers and revisions

Use-case and user-story identifiers are permanent. Revisions are positive,
artifact-local integers:

- `UC-0@r1` is revision 1 of UC-0.
- `US-001@r1` is revision 1 of US-001.
- `INC-01` identifies an increment. Increments are delivery containers and do
  not use requirement revisions.

Increment a revision when externally observable behavior, permission, privacy,
business policy, precondition, outcome, or acceptance expectation changes.
Do not increment it for spelling, formatting, clearer wording with identical
meaning, link repair, or other editorial work.

The existing version on `docs/use-cases.md` remains the edition of the
use-case document as a whole. It is separate from per-use-case revisions. The
traceability baseline adopted all requirements from document version 3.13 as
artifact revision 1; it does not attempt to reconstruct per-artifact revisions
before adoption.

## Status vocabulary

Requirement revisions use:

- `proposed`: under review and not authoritative.
- `accepted`: authoritative for new implementation work.
- `superseded`: retained as history after a later revision is accepted.

Delivery uses:

- `planned`: accepted but implementation has not begun.
- `in_progress`: an executable example or implementation is being developed.
- `verified`: matching revision-tagged acceptance examples pass against the
  deployed application and the required CI checks pass.

Requirement and delivery status are independent. Accepting a new revision does
not inherit verification from the revision it supersedes.

## Sources of truth

- `docs/use-cases.md` owns workflows, permissions, and business rules.
- Individual `docs/user-stories/**/us-*.md` files own story scope and
  acceptance criteria.
- `docs/user-stories/roadmap.md` describes recommended increment sequencing,
  outcomes, and exit criteria.
- `traceability/manifest.yaml` owns artifact revisions, statuses, source
  relationships, increment membership, and implementation pull-request
  evidence.
- Acceptance tests own executable evidence through `// Trace:` metadata.
- Git owns commits. The generated report resolves a recorded PR number to the
  squash-merge SHA found in Git history.
- `docs/traceability.md` is generated and must not be edited directly.

`UC-52` is intentionally undefined and therefore has no manifest entry.
`INC-01` is a technical enabler with no invented user story.

## Acceptance metadata

Place trace metadata immediately above an acceptance test:

```go
// Trace: UC-0@r1 US-001@r1
func TestDesignatedAdminBootstrapsExactlyOnce(t *testing.T) {
    // Business-facing example.
}
```

One example may reference multiple artifacts, and an artifact normally has
multiple examples. Use exact revisions: an example tagged `US-001@r1` does not
verify `US-001@r2`. Foundation examples without numbered requirements use:

```go
// Trace: INC-01
```

Focused domain, application, persistence, and adapter tests need not carry
trace metadata. The acceptance examples are the requirement-level evidence.

## Changing requirements

Requirements changes follow this sequence:

1. Identify the affected UC and all cross-cutting business rules.
2. Add a proposed UC revision in the manifest and change the authoritative
   prose. Preserve the prior revision as `superseded` when the new revision is
   accepted.
3. Identify every affected story through the manifest. Increment a story only
   when its scope, behavior, or acceptance expectations change.
4. Set affected delivery revisions to `planned` or `in_progress`; never carry
   `verified` forward automatically.
5. Add or change revision-tagged acceptance examples and observe the expected
   red result.
6. Implement with focused red-green-refactor cycles until the deployed
   acceptance examples pass.
7. Record the implementation PR and set the delivered revisions and increment
   to `verified` only when their gates are satisfied.
8. Regenerate the report and run all required checks.

An agent may draft or implement a requirement change, but must not silently
invent product policy. Human acceptance is represented by merging the
requirements PR. Ambiguity or conflict with an authoritative use case must be
identified before behavior changes.

If only a story is refined while its source use-case behavior remains
unchanged, increment the story revision only. If a use-case behavior changes,
review every mapped story even when only some need new revisions.

## Starting implementation without changing requirements

1. Read the accepted UC and US revisions in the generated report.
2. Change the story and increment delivery status to `in_progress`.
3. Add a failing acceptance example tagged with those exact revisions.
4. Implement through the public boundary using the project ATDD/TDD workflow.
5. Open the PR early and record its number before marking work `verified`.
6. Regenerate `docs/traceability.md`.

## Pull requests and merge SHAs

Every behavior or requirements PR lists its affected `UC-*`, `US-*`, or
`INC-*` identifiers and revisions. The manifest stores the implementation PR
number, not a hand-entered SHA. Before merge, the generated report shows
`pending merge`; after the squash merge is present in local Git history, the
same report resolves the PR to its merge SHA.

This avoids a self-reference: the final merge SHA cannot be included in the
commit that creates it. It also avoids recording branch commit SHAs that do not
survive a squash merge.

## Commands and gates

```sh
go run ./cmd/traceability write  # validate and regenerate the report
make traceability                # validate manifest, docs, tags, and report
make ci                          # includes the traceability gate
make acceptance                  # execute deployed acceptance specifications
```

Validation rejects:

- documented UC or US identifiers missing from the manifest;
- manifest artifacts that do not exist in the requirements;
- missing or invalid current revisions;
- invalid status transitions represented in revision history;
- unknown source requirement, increment, or acceptance references;
- a `verified` story without an exact revision-tagged acceptance example;
- a `verified` increment without increment-tagged acceptance evidence;
- a verified story without an implementation PR; and
- a stale generated report.

The generated report is a review aid. Passing validation does not prove that
the prose is correct or that every important business-rule example exists;
that remains a responsibility of requirements review and acceptance-example
design.
