# Spec Infrastructure Feedback (for JARVIS)

Filed by: Eitri
Date: 2026-03-24
Source spec: saas-architecture.md

---

## Items for future INFRA specs

1. **RSA keypair setup is not documented anywhere in the spec.**
   The JWT_PRIVATE_KEY_PATH / JWT_PUBLIC_KEY_PATH env vars are required at boot,
   but the spec does not describe key generation, rotation policy, or where keys
   should be stored in production (Fly.io secrets, Vault, etc.).
   Recommend: Add a "Secret Management" section to the next INFRA spec covering
   key generation procedure and Fly.io secrets setup.

2. **Database provisioning is assumed, not specified.**
   The spec says "Fly.io Postgres" but does not specify instance size, storage,
   connection pooling (PgBouncer?), or backup retention for production.
   Eitri defaulted to MaxConns=25 in the pgx pool — this should be validated
   against actual Fly.io Postgres plan limits.

3. **Migration run strategy in production is undefined.**
   The cmd/migrate binary exists. The spec does not say whether migrations
   should run automatically on deploy (risky) or as a pre-deploy step (safer).
   Recommend: Specify migration strategy in the CI/CD section so Falcon can wire it correctly.

4. **R2 bucket creation and CORS policy not specified.**
   The R2 client is written but the bucket policy (public read for avatars?
   presigned upload only?) is not documented in the spec.
   Recommend: Add an R2 section with bucket permissions and CORS config.

5. **Zeffy webhook verification is missing from the spec.**
   The dues handler has HandleZeffyWebhook but Zeffy does not sign webhooks
   like Stripe does. The spec should document how to validate Zeffy payloads
   (IP allowlist? shared secret header?).

6. **The spec lists 47 routes but only ~12 feature packages.**
   Many routes (store, fundraising, committees, study-groups, milestones, etc.)
   have database tables and frontend routes but no corresponding Go packages.
   This creates a gap between the DB schema (fully specified) and the API surface.
   Recommend: JARVIS should generate Iron Man tasks for the remaining packages
   before those frontend pages are built.
