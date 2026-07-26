# 3API Development Guide

This guide records the safe local-development workflow for the 3API repository.

## Safety rules

- Do not commit `.env`, private `config.yaml`, certificates, keys, database dumps, Redis dumps, or production logs.
- Do not delete or reset production data.
- Do not run destructive database commands against production.
- Keep public-facing product copy as **3API**.
- Do not globally replace internal compatibility identifiers. Some module paths, tests, cache keys, and legacy filenames may intentionally remain unchanged.

## Common local checks

Frontend:

```bash
cd frontend
pnpm install
pnpm run typecheck
pnpm run lint:check
pnpm run brand:check
pnpm run test:run
pnpm run build
```

Backend:

```bash
cd backend
go test ./...
```

Repository:

```bash
git status -sb
git diff --check
```

## Local database guidance

Use local-only credentials and disposable local data for development. Never point local experiments at production unless the task is read-only and explicitly authorized.

For migration work:

1. Create or restore a local/isolated database.
2. Run the migration.
3. Verify user, order, balance, payment, redeem-code, subscription, and finance invariants.
4. Record the migration version and rollback plan.

## Release handoff

Before pushing or deploying:

1. Review changed files.
2. Confirm tests and brand check pass.
3. Confirm no secrets are staged.
4. Confirm production backup and isolated restore preflight are ready.
5. Document the commit, image digest, backup path, backup SHA-256, and rollback path.
