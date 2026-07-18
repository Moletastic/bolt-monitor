## 1. Shared Outbound Policy Foundation

- [x] 1.1 Create the standard-library-only `shared/outboundhttp` Go module, add it to `go.work`, and define production constants for five redirects, 1 MiB responses, 30-second total requests, 10-second notification requests, 5-second dial/TLS phases, and 10-second response-header wait.
- [x] 1.2 Implement typed outbound error kinds and safe formatting/mapping helpers that never render raw URLs, causes, headers, bodies, credentials, tokens, or provider payloads.
- [x] 1.3 Implement pure absolute HTTP(S) URL validation, user-information and port rejection, local/AWS hostname normalization, canonical IPv4/IPv6 parsing, IPv4-mapped IPv6 unmapping before policy lookup, and rejection of ambiguous integer, shortened, octal, and hexadecimal IPv4 notation.
- [x] 1.4 Implement a reviewed, retrieval-date-recorded address table derived from the IANA IPv4 and IPv6 Special-Purpose Address Registries; reject each covered range unless both `Globally Reachable` and `Forwardable` are explicitly true, with false/missing/not-applicable values failing closed.
- [x] 1.5 Add table-driven static policy tests covering safe public HTTP/HTTPS URLs, unsupported schemes, malformed/userinfo URLs, special-purpose range boundaries, `100.64.0.0/10` shared/CGNAT space, benchmarking ranges including `198.18.0.0/15`, IPv4/IPv6 documentation and reserved ranges, mapped blocked addresses, numeric bypass forms, localhost variants, and EC2/ECS/EKS/VPC DNS/Time Sync metadata and control literals/aliases.

## 2. Resolution, Dialing, Redirects, and Bounds

- [x] 2.1 Define injectable resolver and dialer interfaces and implement network-aware validation that rejects empty, failed, blocked, and mixed public/blocked A/AAAA answer sets.
- [x] 2.2 Implement pinned-address dialing with no second DNS lookup, no environment proxy, pre-connect candidate validation, post-connect remote-address validation, preserved HTTP Host/TLS ServerName, deterministic approved-address fallback, disabled keep-alives/connection reuse, and immutable per-request policy state.
- [x] 2.3 Implement every redirect as a new current resolution/dial policy decision with a five-redirect cap; reject cross-origin or HTTPS-downgrade redirects for requests carrying credentials, any configured/caller header, or a body rather than stripping data, while allowing only bounded, revalidated, credential/header/body-free public redirects.
- [x] 2.4 Implement the shared request executor with caller/context/hard total deadlines, dial/TLS/header phase limits, guaranteed body closure, and 1 MiB-plus-one bounded response reads.
- [ ] 2.5 Add fake-resolver/fake-dialer tests for aliases, mixed answers, DNS rebinding, DNS changes between sequential requests, stale pooled-connection attempts, unapproved or blocked selected/remote addresses, proxy bypass prevention, IPv4/IPv6 candidates, and public fallback behavior without live network access.
- [ ] 2.6 Add table-driven and race-safe concurrent redirect tests proving request-state isolation, plus cross-origin configured-header/credential/body rejection, safe credential-free redirect revalidation, timeout-phase, cancellation, body-limit, closure, and typed-redaction tests including secret-bearing URL/header/provider failure cases.

## 3. Monitor Configuration and API Preflight

- [x] 3.1 Replace monitor target parsing with shared static policy validation and enforce `http.timeoutMs <= 30000` while preserving `CodeValidationFailed`, `http.target`, and `http.timeoutMs` field details.
- [x] 3.2 Inject network-aware destination validation into monitor API create and update paths so blocked or unresolved current answers fail before persistence with sanitized typed field errors.
- [ ] 3.3 Add table-driven monitor model and monitor API tests for safe public targets, non-HTTP schemes, blocked literals/aliases/resolutions, mixed DNS answers, excessive timeout, update rejection without persistence, and secret-free error envelopes.

## 4. Check Execution and Runtime Wiring

- [x] 4.1 Refactor `shared/checkexecution` to execute through the outbound executor, preserve expected-status-before-body semantics, and map typed policy failures to timeout/error outcomes, stable machine-readable failure codes, and sanitized messages.
- [x] 4.2 Propagate the optional outbound failure code through check-run persistence, status/run response conversions, and API/dashboard types where execution results are exposed, without changing successful result semantics.
- [x] 4.3 Replace timeout-only clients in monitor API manual runs and both check-runtime worker paths with injected production outbound executors using the monitor timeout and shared limits.
- [ ] 4.4 Add checkexecution tests for safe success, blocked/rebound persisted targets, unsafe redirects, header isolation, oversized bodies with and without content assertions, timeout classification, status/body assertion precedence, and sanitized failure codes.
- [ ] 4.5 Add manual-run and worker fake-based tests proving unsafe queued/stored targets are never dialed, failures are recorded, safe public checks still execute, and no live private endpoint is required.

## 5. Notification Validation and Delivery

- [x] 5.1 Extend notification channel create/update validation with injected static and network-aware policy checks for webhook `target` and non-empty email/SMS/PagerDuty `config.apiBaseUrl`, preserving `target` and `config.apiBaseUrl` typed field paths.
- [x] 5.2 Refactor email, SMS, webhook, PagerDuty, Telegram send, and Telegram chat-ID detection to use injected bounded outbound executors; parse and safely join provider base URLs and endpoint paths instead of unchecked string concatenation.
- [x] 5.3 Construct the same production policy-backed sender registry in monitor API test-send and escalation runtime dispatch, and map policy failures to sanitized `CodeNotificationDelivery` categories rather than raw `err.Error()` or provider bodies.
- [x] 5.4 Ensure notification test-send responses and success/failure audit records contain no raw URL secrets, configured headers, credentials, request bodies, or provider response payloads, including oversized and redirect failures.
- [ ] 5.5 Add table-driven channel validation tests for safe public and blocked webhook/provider URLs, DNS aliases and mixed answers, update non-persistence, exact typed field paths, and redacted details.
- [ ] 5.6 Add fake-executor/resolver/dialer sender tests for every notification type, same-origin behavior, cross-origin credential/body redirect rejection, response limits, timeouts, production escalation dispatch, test-send failure mapping, and audit redaction.
- [x] 5.7 Adapt existing loopback `httptest` sender tests to explicit test policy/fake injection so production public-address rules remain strict and tests remain deterministic.

## 6. Contracts and Verification

- [x] 6.1 Update OpenAPI monitor timeout constraints and any exposed check-run failure-code schema/examples; update generated/manual dashboard contract expectations without adding private-monitoring controls.
- [x] 6.2 Run formatting and workspace synchronization, then run `make test-go-all`, `make lint-go`, `make build-go`, `make check-dashboard`, `make test-dashboard`, and `make check-bruno`, fixing regressions within this change's scope.
- [x] 6.3 Verify no direct permissive `http.Client` construction or unbounded response read remains in monitor execution or notification sender production paths, and verify no new deployable service, recurring resource, proxy, or always-on cost was introduced.
