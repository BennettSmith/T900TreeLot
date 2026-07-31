# Engineering work items

Engineering work items track technical and operational delivery that is
necessary to ship or operate the application but does not define user-visible
product policy.

## Identifiers and status

- Use a stable `EW-NNN` identifier and a descriptive file name.
- Use `planned`, `in_progress`, or `completed` for status.
- Record prerequisites, scope, non-goals, completion criteria, and durable
  implementation evidence in the work-item document.
- Link the item from the delivery roadmap when it gates an increment.

## Relationship to requirements traceability

Engineering work items are not use cases or user stories. They do not receive
requirement revisions, `// Trace:` acceptance metadata, or entries in
`traceability/manifest.yaml`. Their implementation history is established by
the linked pull request and recorded completion evidence.

If engineering work changes externally observable behavior, permissions,
privacy, business policy, or product acceptance expectations, update the
affected UC and US artifacts through the requirements traceability process in
addition to the engineering work item.

## Work items

- [EW-001: Render production deployment rehearsal](ew-001-render-production-deployment-rehearsal.md)
  — `planned`; release gate after INC-02 and before INC-03
