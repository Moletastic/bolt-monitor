## Context

Monitor API errors cannot be directly correlated with Lambda logs. Health is public and dependency-independent but documentation does not name it liveness.

## Goals / Non-Goals

**Goals:** safe per-request correlation and separate low-cost authenticated readiness verification.

**Non-Goals:** custom metrics per request, tracing rollout, dependency checks inside public health, or high-frequency synthetic probes.

## Decisions

- Use API Gateway request ID as `X-Request-Id`; generate a fallback only when absent in direct tests.
- Log metadata once at handler boundary. Client correlation ID is secondary and never authoritative.
- Keep `/api/health` liveness-only. Readiness verifies a protected read through deployed API at low frequency.
- Add a dedicated Cognito app client with `USER_PASSWORD_AUTH` for a target-scoped synthetic operator. `setup-readiness` creates or reuses the operator, writes only its generated password to a deterministic SSM SecureString path, and never prints it. `readiness-api` reads that password, mints a short-lived Cognito access token, and verifies one protected read.

## Risks / Trade-offs

- Log volume grows -> emit one small structured line per request and keep existing retention.
- Readiness requires credentials -> use existing operator/synthetic credential mechanism; never embed secrets in source.
- A synthetic user adds lifecycle work -> use a deterministic username, minimum existing membership, idempotent setup, explicit `ROTATE=yes`, and documented teardown.

## Migration Plan

Deploy response header/logging first, then documentation and readiness verification. Health route behavior stays unchanged.

## Open Questions

None.
