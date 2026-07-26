# 3API Apple Container Notes

Apple Container support is optional and intended for local or isolated development only.

Use the same safety principles as Docker:

- keep secrets in local environment files that are not committed;
- use local-only database and Redis volumes;
- do not point experiments at production data;
- record image tags or digests when reproducing an issue;
- back up local volumes before destructive testing.

For production, prefer the standard CI-built immutable image plus backup and isolated restore preflight process.
