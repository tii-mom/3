# 3API Docker Deployment Notes

3API production releases are built as immutable Docker images by the repository CI workflow. Use the image digest produced by the successful build workflow instead of deploying a floating tag.

## Recommended production flow

1. Merge the reviewed commit to the release branch.
2. Wait for CI and Docker image build to complete.
3. Record the image digest from the build output.
4. Create a fresh production database backup.
5. Run isolated restore preflight against that backup and the target image.
6. Deploy only the preflight-approved digest.
7. Run health, readiness, public settings, homepage, admin, payment, redeem, and finance smoke checks.

## Minimal compose shape

The exact image name depends on the repository package settings. Keep the digest immutable in production:

```yaml
services:
  3api:
    image: ghcr.io/<owner>/<package>@sha256:<approved-digest>
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: "postgres://3api@db:5432/3api?sslmode=disable"
      REDIS_URL: "redis://redis:6379"
    depends_on:
      - db
      - redis

  db:
    image: postgres:18
    restart: unless-stopped
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: change-me
      POSTGRES_DB: 3api
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7
    restart: unless-stopped
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

## Safety notes

- Never run a production migration without a backup and isolated restore preflight.
- Never publish `.env`, private configs, database dumps, certificates, or service keys.
- Do not delete or reset production volumes during normal rollout.
- Keep the deployed commit, image digest, backup path, backup SHA-256, and rollback command in the release record.
