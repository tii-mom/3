# 3API Deployment Guide

This folder contains deployment templates and operational helpers for 3API.

The live platform may contain real user, order, balance, payment, redeem-code, and revenue records. Treat every production rollout as a data-safe release.

## Files

| File | Purpose |
| --- | --- |
| `config.example.yaml` | Public configuration template for new installations |
| `DOCKER.md` | Docker deployment notes and immutable image guidance |
| `production-preflight.sh` | Production backup and isolated restore preflight helper |
| `docker-compose.local.yml` | Local development compose file |
| nginx legacy sample | Legacy nginx sample kept for compatibility with older deployments |

Some compatibility filenames are intentionally left unchanged because renaming them can break existing deployment scripts or operator muscle memory. Public product copy should still say **3API**.

## Safe production release checklist

1. Confirm the commit and image digest to deploy.
2. Confirm CI, frontend checks, backend tests, brand check, and `git diff --check` are passing.
3. Verify no secrets are staged:
   - no `.env`
   - no private `config.yaml`
   - no certificates or private keys
   - no database dumps
   - no production logs containing tokens
4. Create a fresh production database backup.
5. Run isolated restore preflight using the backup and target migration set.
6. Confirm the preflight report:
   - restore completed
   - migrations completed or are already current
   - user count and critical financial gates are sane
   - no production source database was modified
7. Deploy the immutable image digest approved by preflight.
8. Run post-deploy smoke checks.
9. Keep rollback instructions, backup path, backup SHA-256, deployed commit, and deployed digest in the release record.

## Post-deploy smoke checks

At minimum:

```bash
curl -fsSL https://api.3api.shop/health
curl -fsSL https://api.3api.shop/health/ready
curl -fsSL https://api.3api.shop/health/live
curl -fsSL "https://api.3api.shop/api/v1/settings/public?timezone=Asia%2FShanghai"
```

Also verify:

- homepage title, logo, and public settings show 3API
- SEO pages return HTML snapshots with canonical links
- anonymous homepage loading does not call protected auth endpoints
- login, user console, redeem-code page, purchase page, and admin finance page open normally
- payment, balance, redeem, withdrawal, and compute-company features are still gated by backend authority

## Rollback discipline

Rollback should prefer immutable image rollback first. Database rollback is a last resort and must be handled with a fresh backup and explicit operator approval.

Never delete, truncate, or reset production user data during normal rollout.
