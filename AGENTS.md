# Agent Guide

## Project purpose

This repository defines the Troop 900 Tree Lot Shift Scheduler, a small but
security-sensitive web application for coordinating volunteers during the
troop's annual tree-lot season.

Before changing behavior, read the relevant sections of:

- `docs/use-cases.md` — authoritative user workflows, permissions, and business
  rules. Stable use-case IDs (for example, UC-10 and UC-53) provide traceability.
- `docs/architecture.md` — proposed technical architecture, module boundaries,
  persistence, security, operations, and testing strategy.
- `docs/design-system-requirements.md` — UI, accessibility, responsive behavior,
  and reusable visual patterns.
- `docs/traceability-process.md` — requirement revisions, delivery statuses,
  executable evidence, and the human/agent change workflow.

Do not infer policy from a mockup or existing implementation when it conflicts
with the use cases. If requirements conflict or are unclear, identify the
conflict instead of silently choosing new product behavior.

## Architecture

- Build a Go modular monolith with separate web, worker, migration, and offline
  season-restore entry points backed by one PostgreSQL database.
- Use hexagonal architecture. Domain code contains business behavior and has no
  dependency on HTTP, SQL, templates, Groups.io, or hosting.
- Application services orchestrate use cases, authorization, transactions, and
  outbound ports. The consuming module declares each port.
- Inbound adapters decode requests, resolve the actor, invoke one application
  command or query, and translate the result. HTTP handlers contain no business
  rules.
- Outbound adapters implement PostgreSQL, blob storage, messaging, clock, ID,
  and audit ports. Do not leak adapter types into domain APIs.
- Preserve bounded contexts: Identity and Access, Families, Seasonal
  Agreements, Scheduling, Attendance, Communications, Reporting, and Privacy
  and Audit. Access another context through an explicit policy, query, or
  application interface rather than duplicating its rules.
- Render semantic HTML on the server with `html/template`. HTMX progressively
  enhances the interface. Browser JavaScript is required for passkeys
  (WebAuthn) and may be used for other browser APIs; do not introduce a
  client-side application framework such as React or Angular.
- PostgreSQL is the source of truth and job queue. Profile images initially use
  PostgreSQL `BYTEA` behind a `BlobStore` port.
- External provider calls never run inside database transactions. Persist
  resulting messages in the transactional outbox and make workers idempotent.
- Store instants in UTC. Resolve shift dates and times with the configured
  `TREE_LOT_TIME_ZONE` and an injected clock; never rely on host-local time.

## Domain model and authorization

- An authenticated identity is distinct from a person profile. One person has
  one profile and may hold multiple roles.
- A household is the management and assignment-ownership boundary. A family
  unit is a reporting boundary. A scout may belong to multiple households
  while retaining one profile and one schedule.
- Authorization is relationship-aware. Evaluate identity, roles, linked person,
  household authority, assignment ownership, active status, season agreement,
  shift state, capacity, timing, and audited overrides as applicable.
- Never trust hidden controls, HTMX headers, or submitted person/household IDs
  as authority. Every state-changing command performs server-side authorization.
- Seasonal agreement confirmation belongs to one person, one season, and the
  current configured Google Doc link. Replacing the link resets confirmations.
  The selected participant's confirmation gates signup, check-in, and walk-ins;
  Committee and Admin cannot override it.
- Shift capacity, role-slot matching, and one-assignment-per-person rules are
  transactional invariants. Protect them with database constraints plus row
  locking or conditional updates under concurrency.
- Assignment cancellation follows origin ownership: managers control
  household-owned assignments from their household; self-created Young Adult
  Scout assignments can also be managed by linked households. Committee/Admin
  overrides are audited.
- Youth-protection coverage and minimum headcount are separate rules. Use
  explicit adult-to-scout relationships for the sole-adult family exception;
  household membership alone is insufficient.
- Attendance events are immutable and use server time within defined windows.
  Corrections are separate, reasoned, audited records. Do not rewrite or
  backdate real-time events.
- Removing login access is not personal-data deletion. Preserve profile and
  history unless the separately verified privacy workflow requires deletion or
  anonymization.
- Scout Bucks calculations deduplicate people across households, retain
  full-precision credited hours, allocate integer cents deterministically, and
  create immutable revisioned settlements whose awards exactly equal the pool.

## Persistence, security, and delivery

- Use one PostgreSQL transaction per state-changing command to protect
  invariants, persist domain changes, append audit facts, and enqueue outbox
  records.
- Authenticate with passkeys (WebAuthn). Use a claimed email as the unique
  account identifier; defer mailbox verification until opted-in email
  notifications or email-based recovery require it. Enrollment and recovery use
  authorized invitation links or QR codes, not SMS authentication codes.
- Enforce normalized active login-email uniqueness system-wide. Store only
  public passkey credential material; private keys never leave the
  authenticator. Do not store phone numbers for authentication or operational
  notification delivery.
- Store only hashes of session and single-use tokens. Use secure, HTTP-only,
  same-site cookies and CSRF protection for every state-changing browser request.
- Avoid identity enumeration. Redact emails, tokens, message bodies, and
  provider secrets from logs and audit records.
- Do not use SMS OTPs for authentication or SMS for operational notifications.
  Record every recipient-facing notice in the per-user in-app inbox. Groups.io
  is optional, disabled by default, and limited to troop-wide posts; direct and
  family-scoped messages remain in-app only. One channel's failure must not roll
  back another channel or in-app publication. A later verified-email channel may
  supplement the inbox after mailbox verification and user opt-in.
- `GET` is side-effect free. Use Post/Redirect/Get for ordinary forms where
  appropriate, and ensure full-page and HTMX paths enforce identical behavior.

## Commit messages

- Use Conventional Commit subjects in the form
  `<type>[optional scope][!]: <description>`.
- Allowed types are `build`, `chore`, `ci`, `docs`, `feat`, `fix`, `perf`,
  `refactor`, `revert`, `style`, and `test`.
- Install the tracked local hooks with `make install-hooks`. Do not bypass the
  pre-push commit-message check.

## Test-driven development

Always develop production behavior in short red-green-refactor cycles:

1. **Red:** Write the smallest test that describes the next behavior and run it
   to confirm that it fails for the expected reason.
2. **Green:** Write only enough production code to make the failing test pass.
3. **Refactor:** While all tests remain green, improve names, structure, and
   duplication without changing behavior.

Follow the Three Laws of TDD:

1. Do not write production code unless it is needed to make a failing test pass.
2. Do not write more of a test than is sufficient to fail; compilation failures
   count as failures.
3. Do not write more production code than is sufficient to pass the currently
   failing test.

Run the relevant test after each step. Never skip observing the red state, and
never refactor while tests are failing. Use an acceptance test as the outer
cycle for a use case and focused domain, application, persistence, or adapter
tests as inner cycles.

## Definition of done

No body of work is done until `make ci` has been run from the repository root
and completes successfully. Run it after the final changes, even when narrower
tests have already passed. If `make ci` cannot run or fails, report the blocker
or failure and do not describe the work as complete.

Product behavior is not verified until its revision-tagged acceptance examples
also pass against the deployed system. Run the relevant acceptance suite after
`make ci`; run `make acceptance` when no narrower deployed suite exists.

## Requirement and delivery traceability

- Stable use-case and user-story IDs never change. Refer to exact revisions as
  `UC-0@r1` and `US-001@r1`.
- `traceability/manifest.yaml` is the source of truth for revisions, statuses,
  source relationships, increment membership, and implementation PR evidence.
  `docs/traceability.md` is generated; never edit it directly.
- Increment a requirement revision only for semantic behavior or policy
  changes, not editorial changes.
- A new accepted revision invalidates prior delivery verification. Set affected
  delivery to `planned` or `in_progress` and update executable examples before
  implementation.
- Agents may propose requirement changes but must not silently invent or accept
  product policy. Merging the requirements PR represents human acceptance.
- Open the implementation PR early, record its number before marking a revision
  `verified`, and regenerate the report with
  `go run ./cmd/traceability write`.
- Run `make traceability` before delivery. It validates the manifest, documented
  IDs, exact acceptance revisions, implementation evidence, and generated
  report.
- Follow `docs/traceability-process.md` whenever requirements or delivery
  status changes.

## Implementing a use case

1. Locate the use case and all referenced business rules in
   `docs/use-cases.md`; check the permission summary and cross-cutting rules.
2. Identify the accepted UC and US revisions in `docs/traceability.md`, the
   owning bounded context, and any policies needed from other contexts.
3. Set affected story and increment delivery statuses to `in_progress`.
4. Add a business-facing executable acceptance example tagged with the exact
   use-case and user-story revisions and confirm that it fails for the expected
   reason.
5. Implement domain behavior and application orchestration with focused tests,
   then add persistence and adapter behavior.
6. Exercise the deployed system through public browser/HTTP or provider
   boundaries. Acceptance tests must not call application services directly,
   mutate private tables, or replace internal repositories.
7. Verify successful behavior, denials, privacy boundaries, concurrency,
   idempotency, time-window edges, and audit/outbox effects that apply.
8. Record the implementation PR, mark only fully evidenced revisions
   `verified`, regenerate `docs/traceability.md`, and run the required gates.

Use injected clocks in time-dependent tests. Poll observable asynchronous
outcomes with deadlines; do not use fixed sleeps or retries that hide flakes.
Keep acceptance-test vocabulary in tree-lot domain language rather than routes,
selectors, tables, or implementation terminology.

## Cursor Cloud specific instructions

The Cloud VM has no `systemd`; services are started by the environment update
script on boot, not by an init system. Standard commands and DB resolution are
documented in `README.md` and the `Makefile`; only the non-obvious caveats
below are repeated here.

Services required to develop and validate this repository:

- PostgreSQL 16 on `127.0.0.1:5432`, role/password `treelot`/`treelot`, with
  databases `treelot` (dev) and `treelot_test` (unit/component tests). The
  update script starts the cluster (`pg_ctlcluster 16 main start`) and creates
  the role/databases idempotently.
- Docker with the `vfs` storage driver, used only by `make acceptance`.

Non-obvious caveats:

- `make ci` / `make test` need only local PostgreSQL. `scripts/ensure-test-db.sh`
  finds `treelot_test` on `:5432` and never touches Docker in this environment.
  Do not set `TEST_DATABASE_URL`; leaving it unset uses the reachable `:5432` DB.
- `make acceptance` needs Docker. It builds the production image and runs its
  own throwaway Compose PostgreSQL on host port `:5433` plus host-networked
  acceptance web, production web, and stub on `:18080`, `:18081`, and `:18090`
  by default. Those application ports are configurable with
  `ACCEPTANCE_WEB_PORT`, `ACCEPTANCE_PRODUCTION_PORT`, and
  `ACCEPTANCE_STUB_PORT`. It does not use the dev PostgreSQL on `:5432`, so the
  two never conflict.
- Docker 29 defaults to the containerd snapshotter; the classic `vfs` graph
  driver requires `/etc/docker/daemon.json` to set
  `{"storage-driver":"vfs","features":{"containerd-snapshotter":false}}`.
  `iptables` is set to `iptables-legacy` for nested-Docker networking.
- `scripts/preflight.sh` (run by `make acceptance`) calls `docker info` without
  `sudo`. The `ubuntu` user is in the `docker` group, so a fresh login shell can
  reach Docker without `sudo`; a shell opened before that group was applied
  cannot, and the `Makefile` then falls back to `sudo docker`.
- Run the web app in dev without Docker: export the vars from `.env.example`
  with `DATABASE_URL=postgres://treelot:treelot@127.0.0.1:5432/treelot?sslmode=disable`,
  apply schema with `go run ./cmd/migrate up`, then `go run ./cmd/web` (serves
  `:8080`). `make up`/`make acceptance` remain Docker-only.
