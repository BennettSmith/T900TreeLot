# Render production operations

This runbook covers first-time Render provisioning, digest promotion, DNS/TLS,
rollback, and isolated PostgreSQL point-in-time recovery for EW-001.

Do not paste secrets, bootstrap tokens, passkey details, emails, or session
material into tickets, chat, git, or completion evidence.

## Topology

| Resource | Name | Plan | Region | Notes |
|---|---|---|---|---|
| Web | `treelot-web` | `starter` (paid) | `oregon` | Health `/health/ready`, pre-deploy `/app/migrate up` |
| Worker | `treelot-worker` | `starter` (paid) | `oregon` | Same image; `dockerCommand: /app/worker` |
| PostgreSQL | `treelot-db` | `basic-256mb` + 1 GB disk | `oregon` | Paid plan required for PITR |
| Cron | deferred | — | — | Provision in INC-07 only |

Canonical origin: `https://treelot.troop900livermore.org`

Infrastructure as code: [`render.yaml`](../../render.yaml)

## Prerequisites

1. INC-02 is `verified` in `docs/traceability.md`.
2. `make ci` and `make acceptance` pass on the candidate commit locally or in CI.
3. Render workspace billing and ownership continuity are confirmed.
4. Operator can update DNS for `troop900livermore.org`.
5. A private GHCR package will receive images from GitHub Actions.

Confirm live Render pricing before creating billable resources. The plans above
are the current lowest-cost paid combination that supports pre-deploy commands
and PostgreSQL PITR.

## First-time setup checklist

### 1. GitHub Container Registry

1. Ensure the repository can publish packages (`packages: write` is granted by
   [`.github/workflows/release.yml`](../../.github/workflows/release.yml)).
2. After the first successful Release publish, open the package settings in
   GitHub and confirm the package is private.
3. Create a GitHub personal access token (classic) or fine-grained token with
   `read:packages` only for Render image pulls.
4. In Render → Workspace Settings → Registry Credentials, create credential
   name **`ghcr-treelot`** using that token. Never commit the token.

### 2. Apply the Blueprint

1. In the Render Dashboard, create a Blueprint from this repository's
   `render.yaml`.
2. Choose region **Oregon** for all resources if prompted.
3. When Render prompts for `sync: false` secrets, generate them locally and
   paste them only into the Render UI:

```sh
# SESSION_KEY (≥32 characters). Example generator:
openssl rand -base64 48

# BOOTSTRAP_ENROLLMENT_TOKEN (≥24 characters). Example generator:
openssl rand -base64 32

# BOOTSTRAP_TOKEN_EXPIRES_AT — set immediately before the operator window.
# macOS:
date -u -v+4H '+%Y-%m-%dT%H:%M:%SZ'
# GNU/Linux:
date -u -d '+4 hours' '+%Y-%m-%dT%H:%M:%SZ'
```

4. Set the same `SESSION_KEY`, bootstrap token, and expiry on **both** web and
   worker.
5. Confirm non-secret values:

- `APP_ENV=production`
- `PUBLIC_BASE_URL=https://treelot.troop900livermore.org`
- `TREE_LOT_TIME_ZONE=America/Los_Angeles`
- `WEBAUTHN_RP_ID=treelot.troop900livermore.org`
- `GROUPS_IO_ENABLED=false`

6. Confirm web health check path is `/health/ready`.
7. Confirm web pre-deploy command is `/app/migrate up`.
8. Confirm worker start command / Docker command is `/app/worker`.
9. Do **not** set `TEST_CONTROL_KEY` in production.
10. Do **not** provision the hourly cron job.

### 3. GitHub production environment

Create a GitHub Environment named **`production`** with required reviewers.
Add secrets (values only in GitHub; never in git):

| Secret | Purpose |
|---|---|
| `RENDER_API_KEY` | Render API key with deploy permission |
| `RENDER_WEB_SERVICE_ID` | Render service id for `treelot-web` |
| `RENDER_WORKER_SERVICE_ID` | Render service id for `treelot-worker` |

### 4. DNS and TLS

1. In Render, attach custom domain `treelot.troop900livermore.org` to
   `treelot-web` (also declared in `render.yaml`).
2. Create the exact DNS record Render displays for the `treelot` hostname.
3. Do not change the existing `www` marketing site.
4. Wait until Render shows a verified certificate.
5. Verify:

```sh
curl -fsS -o /dev/null -w '%{http_code}\n' https://treelot.troop900livermore.org/health/live
curl -fsS -o /dev/null -w '%{http_code}\n' https://treelot.troop900livermore.org/health/ready
curl -fsSI https://treelot-web.onrender.com/sign-in | rg -i 'HTTP/|location:'
```

Expect health endpoints to return `200`. Expect the Render hostname to
permanently redirect browser paths to the canonical origin while still
answering `/health/*` without redirect.

## Immutable image promotion

1. On `main` (or the release candidate SHA), run the **Release** workflow.
2. Leave `deploy_production` unchecked for the first publish/accept pass, or
   check it only after the production environment reviewers are ready.
3. Optionally pass `predecessor_digest` so the run records the prior accepted
   digest for rollback.
4. The workflow:

- builds once and pushes `ghcr.io/<owner>/<repo>:<git-sha>`
- resolves the immutable `sha256` digest
- pulls that digest
- runs `ACCEPTANCE_SKIP_BUILD=1 IMAGE=treelot:candidate make acceptance`
- if approved, deploys the same digest to web and worker via the Render CLI

5. Record the source SHA, digest, and Render deploy ids in EW-001 evidence.

Local acceptance against an already-built image:

```sh
docker pull ghcr.io/bennettsmith/t900treelot@sha256:<digest>
docker tag ghcr.io/bennettsmith/t900treelot@sha256:<digest> treelot:candidate
ACCEPTANCE_SKIP_BUILD=1 IMAGE=treelot:candidate make acceptance
```

## Deployment rehearsal

1. Deploy the accepted digest to empty production PostgreSQL.
2. Confirm pre-deploy migrate applied every migration and web/worker start.
3. Check structured JSON startup logs in Render for web and worker.
4. Verify static assets: `https://treelot.troop900livermore.org/static/app.css`
5. Confirm worker remains running (no crash loop).
6. Designated first Admin completes bootstrap and sign-in with a **real**
   passkey at the canonical HTTPS origin.
7. Restart or redeploy web and worker; Admin signs in again successfully.
8. Roll back both services to the predecessor accepted digest; confirm health.
9. Redeploy the candidate digest and confirm health again.

## Application rollback

Prefer redeploying a previously accepted digest, not rebuilding:

```sh
# Example using Render CLI after installing it and exporting RENDER_API_KEY.
render deploys create "$RENDER_WEB_SERVICE_ID" \
  --image "ghcr.io/bennettsmith/t900treelot@sha256:<predecessor>" \
  --confirm --wait
render deploys create "$RENDER_WORKER_SERVICE_ID" \
  --image "ghcr.io/bennettsmith/t900treelot@sha256:<predecessor>" \
  --confirm --wait
```

Only roll back across schema-compatible digests. Expand-and-contract migrations
are required before any incompatible schema change.

## Isolated PostgreSQL PITR rehearsal

1. In Render → `treelot-db` → Recovery → Point-in-Time Recovery, restore to a
   **new** database instance.
2. Choose a restore time at least ten minutes in the past.
3. Keep production services pointed at the original database.
4. Measure wall-clock time from Start Recovery until the new instance is ready.
5. Connect to the isolated recovery instance only and verify:

```sql
SELECT MAX(version) AS schema_version FROM schema_migrations;
SELECT COUNT(*) AS identity_count FROM identities;
SELECT COUNT(*) AS session_count FROM sessions;
```

6. Confirm schema version matches the deployed application expectation.
7. Delete or suspend the recovery instance after verification.
8. Record: PITR timestamp, isolated target name/id, duration, and verification
   result. Do not record row contents that include personal data.

## Secret handling rules

- Generate secrets on an operator machine; paste only into Render or GitHub
  Environments.
- Rotating `SESSION_KEY` invalidates all browser sessions.
- Restarting services does not extend `BOOTSTRAP_TOKEN_EXPIRES_AT`.
- After successful first-Admin bootstrap, leave the bootstrap secret expired or
  rotate it out of operational use according to troop policy.
- Never enable acceptance test-control credentials in production.

## Completion evidence pointers

Record non-secret evidence in
[`docs/engineering-work-items/ew-001-render-production-deployment-rehearsal.md`](../engineering-work-items/ew-001-render-production-deployment-rehearsal.md):

- merged implementation PR
- accepted image digest + source commit
- Render web/worker deployment ids
- migration version + readiness result
- canonical domain + TLS verification date
- operator confirmation that bootstrap, sign-in, restart, and rollback passed
- PITR timestamp, isolated target, duration, verification result
- successful `make ci` and whole-system acceptance for the deployed candidate
