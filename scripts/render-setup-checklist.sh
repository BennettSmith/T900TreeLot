#!/usr/bin/env bash
# Prints the EW-001 first-time Render setup checklist without requiring secrets.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cat <<'EOF'
EW-001 Render first-time setup (operator checklist)
==================================================

Follow docs/runbooks/render-production.md. Paste secrets only into Render or
the GitHub production environment—never into this shell transcript, chat, or git.

[ ] 1. Confirm INC-02 is verified and make ci / make acceptance pass.
[ ] 2. Confirm Render billing ownership and Oregon region.
[ ] 3. Create GHCR read credential named ghcr-treelot in Render.
[ ] 4. Apply render.yaml Blueprint (web, worker, paid Postgres).
[ ] 5. Enter SESSION_KEY, BOOTSTRAP_ENROLLMENT_TOKEN, BOOTSTRAP_TOKEN_EXPIRES_AT.
[ ] 6. Confirm APP_ENV/PUBLIC_BASE_URL/timezone/WebAuthn/GROUPS_IO_ENABLED.
[ ] 7. Confirm /health/ready and pre-deploy /app/migrate up on web only.
[ ] 8. Create GitHub Environment "production" with RENDER_* secrets.
[ ] 9. Attach treelot.troop900livermore.org and create the Render DNS record.
[ ] 10. Verify TLS, health, and noncanonical-host redirect.
[ ] 11. Run Release workflow; accept digest; deploy with approval.
[ ] 12. Complete real-passkey bootstrap/sign-in, restart, rollback, isolated PITR.
[ ] 13. Record non-secret evidence in EW-001 and mark completed.

Useful local commands:

  make ci
  make acceptance
  ACCEPTANCE_SKIP_BUILD=1 IMAGE=treelot:candidate make acceptance

Runbook:

EOF
echo "  ${ROOT}/docs/runbooks/render-production.md"
