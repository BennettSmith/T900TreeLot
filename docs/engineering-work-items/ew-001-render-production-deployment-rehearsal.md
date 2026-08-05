# EW-001: Render production deployment rehearsal

- **Status:** `in_progress`
- **Release gate:** After INC-02 and before INC-03
- **Prerequisite:** INC-02 is verified against the local production-shaped
  acceptance environment
- **Implementation PR:** To be recorded when opened

## Goal

Prove that the acceptance-tested application can be deployed, operated, and
recovered on Render at the canonical production origin before additional
product increments depend on an untested hosting path.

This is technical and operational delivery work. It does not define a new user
workflow or revise a use case or user story.

## Required outcome

The same immutable image that passes CI and the isolated whole-system
acceptance suite runs the production web and worker processes. Render applies
schema migrations through the dedicated migration entry point, serves the
canonical HTTPS origin, and provides a managed PostgreSQL recovery path that
meets the architecture's production objectives.

## Scope

### Application production readiness

- Handle Render termination signals and allow in-flight web requests to shut
  down within the platform's termination window.
- Keep `/health/live` process-only and make `/health/ready` report database
  availability without returning secrets or dependency details.
- Redirect noncanonical public application hosts to
  `https://treelot.troop900livermore.org` without breaking platform health
  checks.
- Confirm production responses use secure, HTTP-only, same-site, host-only
  session cookies and the required browser security headers.
- Ensure production never enables acceptance test-control routes or
  credentials.

### Immutable image delivery

- Build the production image once in CI and publish it to the selected
  container registry using an immutable digest.
- Run the whole-system acceptance suite against that candidate image rather
  than rebuilding application code afterward.
- Deploy the accepted digest to both Render web and worker services.
- Retain enough image and deployment history to redeploy the preceding known
  good digest during rollback.

### Render topology

- Define the production web service, background worker, and managed PostgreSQL
  database as infrastructure as code where Render supports it.
- Run `/app/migrate up` as the single pre-deploy schema migration step. Web and
  worker processes continue to reject incompatible schemas and never apply
  migrations themselves.
- Configure the web health check to use `/health/ready`.
- Place web, worker, and PostgreSQL in the same Render region.
- Select a paid PostgreSQL plan whose backups and point-in-time recovery can
  satisfy the active-season recovery-point and recovery-time objectives in the
  architecture.

### Production configuration

- Set `APP_ENV=production`,
  `PUBLIC_BASE_URL=https://treelot.troop900livermore.org`,
  `TREE_LOT_TIME_ZONE=America/Los_Angeles`, and
  `GROUPS_IO_ENABLED=false`.
- Supply `DATABASE_URL`, `SESSION_KEY` (HMAC key for hashing opaque session
  cookie tokens), the one-time bootstrap secret, and WebAuthn relying-party
  configuration through secret environment settings. No secret is committed,
  embedded in the image, or included in logs or completion evidence.
- Configure the WebAuthn relying-party ID and allowed origin for the canonical
  hostname.
- Point troop-managed DNS at the Render web service and verify Render-managed
  TLS for the canonical hostname.

Operator procedures for first setup, digest promotion, DNS/TLS, rollback, and
isolated PITR live in
[`docs/runbooks/render-production.md`](../runbooks/render-production.md).
Infrastructure as code lives in [`render.yaml`](../../render.yaml).

### Deployment and recovery rehearsal

- Deploy an empty production database and apply every migration from the
  beginning.
- Verify liveness, readiness, static assets, structured startup logs, web
  traffic, and continuous worker operation.
- Have the designated first Admin complete the accepted bootstrap and sign-in
  flow with a real passkey at the canonical HTTPS origin. Do not add a separate
  test-only bootstrap path.
- Redeploy or restart web and worker services and verify the Admin can sign in
  again without identity or session-store corruption.
- Exercise application rollback to the preceding compatible image digest.
- Restore a managed PostgreSQL backup or point-in-time snapshot into an
  isolated recovery target, verify schema compatibility and expected records,
  and document the measured recovery procedure. Do not overwrite production
  during the rehearsal.

## Implementation progress

- Status set to `in_progress` after INC-02 verification on `main`.
- Application hardening: production-only canonical-host redirect with health
  exemptions; graceful web shutdown already present on `main`; `SESSION_KEY`
  remains required and is used to HMAC-hash opaque session tokens.
- Immutable delivery: `ACCEPTANCE_SKIP_BUILD` acceptance path, Release workflow
  publishing GHCR digests with gated Render promotion, and `render.yaml`.
- Operator guide: [`docs/runbooks/render-production.md`](../runbooks/render-production.md)
  and `make render-setup-checklist`.
- Local gates for this candidate: `make ci` and `make acceptance` passed,
  including `ACCEPTANCE_SKIP_BUILD=1` against a local predecessor-tagged image.

Operator-owned remaining steps (Render account, DNS, real passkey, PITR) are
listed in the runbook and must complete before status becomes `completed`.

## Completion evidence

Before changing the status to `completed`, record:

- The merged implementation PR.
- The accepted image digest and corresponding source commit.
- Render deployment identifiers for web and worker.
- The migration version and successful readiness result.
- The canonical-domain and TLS verification date.
- The designated operator's confirmation that bootstrap, sign-in, restart,
  and rollback checks passed; do not record the email address, passkey details,
  bootstrap token, or session material.
- The backup or point-in-time recovery timestamp, isolated restore target,
  measured restore duration, and verification result.
- Successful `make ci` and whole-system acceptance runs for the deployed
  candidate.

### Recorded so far

| Evidence | Value |
|---|---|
| Implementation PR | To be filled after PR open |
| Local `make ci` | Passed on candidate branch |
| Local `make acceptance` | Passed on candidate branch |
| Skip-build acceptance | Passed with `IMAGE=treelot:predecessor` |
| GHCR digest / Render deploy ids | Pending Release workflow + operator approval |
| Canonical TLS / real passkey / PITR | Pending operator rehearsal |

## Non-goals

- Do not add or revise numbered user stories, use cases, requirement revisions,
  or traceability-manifest entries.
- Do not enable Groups.io. Its adapter and operational behavior belong to
  INC-07 and remain optional.
- Do not provision the hourly reminder-enqueue cron job until INC-07 provides
  its executable command and idempotent behavior.
- Do not populate households, schedules, assignments, or other later-increment
  production data.
- Do not treat this rehearsal as the pre-season operational restore drill;
  repeat recovery verification before each season as required by the
  architecture.
