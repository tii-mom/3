# 3API

3API is an AI API gateway and operations platform for unified access to multiple model providers. It includes API relay, account and channel management, payment and balance operations, redeem-code workflows, subscription packages, image-generation bridging, and the T0-T3 compute-company revenue center.

Production website: https://3api.shop

## What this project contains

- Multi-model API gateway for OpenAI-compatible clients, Codex, Claude, Gemini, Grok, images, embeddings, and streaming scenarios.
- Admin console for users, groups, upstream accounts, API keys, payment plans, finance operations, risk controls, and audit records.
- User console for key usage, purchase, redeem-code exchange/generation, distribution/company revenue, withdrawal requests, and platform balance operations.
- SEO-ready public pages with generated HTML snapshots, sitemap, robots, canonical links, and public-page `noindex` controls for private routes.
- Operational safety gates for migrations, financial reconciliation, backup preflight, and production rollout.

## Brand policy

All public-facing product names, page titles, descriptions, navigation labels, and documentation maintained for this repository should use **3API**.

This repository is derived from an upstream codebase, so a few internal compatibility names may still exist in module paths, database defaults, cache keys, tests, or historical migration documents. Do not perform a global text replacement. Only replace names that are exposed to users, operators, public pages, or public documentation.

Use the brand check before release:

```bash
cd frontend
pnpm run brand:check
```

## Development checks

Frontend:

```bash
cd frontend
pnpm install
pnpm run typecheck
pnpm run lint:check
pnpm run test:run
pnpm run build
```

Backend:

```bash
cd backend
go test ./...
```

Repository formatting check:

```bash
git diff --check
```

## Release safety

The live platform has real user data. Production changes must be handled conservatively:

1. Review the final diff and confirm no secrets, private configs, certificates, dumps, or `.env` files are staged.
2. Run frontend checks, backend tests, brand check, and repository format checks.
3. Build immutable release artifacts.
4. Create a fresh production backup.
5. Run an isolated restore preflight against the backup and migration set.
6. Deploy only the preflight-approved image or artifact.
7. Run health, readiness, public settings, homepage, SEO, admin, payment, redeem, and finance smoke checks.
8. Keep the backup path, SHA-256, deployed version, image digest, and rollback command with the release record.

Never delete or reset the production database as part of a normal release.

## Documentation

- Payment configuration: [docs/PAYMENT.md](docs/PAYMENT.md)
- Chinese payment configuration: [docs/PAYMENT_CN.md](docs/PAYMENT_CN.md)
- Image generation API: [docs/IMAGE_GENERATION_API.md](docs/IMAGE_GENERATION_API.md)
- Admin payment integration: [docs/ADMIN_PAYMENT_INTEGRATION_API.md](docs/ADMIN_PAYMENT_INTEGRATION_API.md)
