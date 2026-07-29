## 1. Remove the Unused Dashboard Dependency

- [x] 1.1 Verify `devicons-react` has no dashboard source, test, generated-code, or build-script use.
- [x] 1.2 Remove `devicons-react` from the dashboard manifest and update `pnpm-lock.yaml` deterministically.
- [x] 1.3 Run a frozen dashboard install, `make lint-dashboard`, `make check-dashboard`, `make test-dashboard`, and `make build-dashboard`.
