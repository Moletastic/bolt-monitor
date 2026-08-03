## 1. Request Correlation

- [x] 1.1 Add generated/fallback request ID handling at the monitor API response boundary.
- [x] 1.2 Return `X-Request-Id` on success and error responses.
- [x] 1.3 Emit one safe structured boundary log with generated and secondary correlation metadata.
- [x] 1.4 Add direct Lambda handler tests for response IDs and safe correlation behavior.

## 2. Health and Readiness

- [x] 2.1 Document `/api/health` as public liveness with no dependency checks.
- [x] 2.2 Add target-scoped `setup-readiness` and authenticated protected-read `readiness-api` commands using SSM SecureString credentials.
- [x] 2.3 Document readiness operation and failure diagnosis without exposing credentials.
- [x] 2.4 Add `INSTALLATION.md` covering first deploy, readiness lifecycle, rotation, and teardown.

## 3. Verification

- [x] 3.1 Confirm log retention and probe schedule stay bounded and no high-cardinality metrics are added.
- [x] 3.2 Run `make test-go-all`, `make check-api-contract`, and relevant infrastructure checks.
