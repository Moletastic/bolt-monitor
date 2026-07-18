## Context

HTTP monitor configuration currently checks only that `url.ParseRequestURI` produces an absolute URL. Manual and recurring checks then use `http.Client{Timeout: ...}`, which follows redirects, resolves names in the default transport, and reads the entire response body. Notification senders independently construct default clients for Telegram, SendGrid, Twilio, webhooks, and PagerDuty; webhook targets and optional provider `apiBaseUrl` values are not constrained to public HTTP(S) destinations, and provider error bodies are read and formatted without a size boundary.

These call paths run with AWS credentials and network access, so operator-controlled URLs must not reach runtime-local, VPC, link-local, or metadata/control services. Validation alone is insufficient because redirects and DNS answers can change between persistence and execution. The repository uses small Go modules under `shared/`, typed domain errors, dependency injection through narrow functions/interfaces, and no resident application process.

## Goals / Non-Goals

**Goals:**
- Define one reusable default-public outbound HTTP boundary for monitors and all notification senders.
- Reject unsafe URL schemes, literals, aliases, DNS answers, redirects, and dial destinations under an IANA special-purpose-registry-derived, default-public IPv4 and IPv6 policy.
- Prevent DNS rebinding by coupling validation to the addresses actually dialed.
- Bound redirect count, connection phases, total duration, and response bytes.
- Keep caller headers, credentials, payloads, and provider responses out of redirected origins and diagnostics.
- Preserve safe public HTTP(S) monitors and provider delivery with deterministic table-driven tests and fakes.
- Add no new deployed service, proxy, scheduler, or always-on cost.

**Non-Goals:**
- Private, VPC, on-premises, or allowlisted internal monitoring.
- Operator bypass flags, per-tenant network policies, or custom CIDR allowlists.
- A new egress proxy, NAT architecture, firewall product, or DNS service.
- General-purpose HTTP client replacement outside monitor and notification outbound traffic.
- Changing notification provider APIs, monitor request payload shape, or incident state semantics.

## Decisions

### Add a standard-library-only `shared/outboundhttp` module

Create a leaf module that depends only on the Go standard library. It owns URL parsing, address classification, resolution, safe dialing, redirects, phase and total timeouts, bounded response reads, and typed policy failures. `monitorconfig`, `checkexecution`, `notifications`, and service wiring may depend on it; it must not depend on those domain modules.

The package exposes narrow injectable resolver and dialer interfaces plus a policy/executor configured with fixed production limits. A request API returns status, bounded body, and typed errors so callers cannot accidentally bypass the body limit by calling `io.ReadAll` themselves. Production constructors use `net.Resolver` and `net.Dialer`; tests provide deterministic fakes.

Alternative considered: add checks independently to each `http.Client`. Rejected because redirect, DNS, dial, timeout, and redaction rules would drift across manual checks, workers, and five notification senders.

Alternative considered: deploy an egress proxy. Rejected because it adds an always-on component and cost, and the requested boundary can be enforced in-process.

### Separate static URL checks from network-aware preflight

The package has a pure parse/static-validation step and a context-aware destination validation step. Static validation requires an absolute HTTP(S) URL, a hostname, no user information, a valid port, and a permitted literal address. It rejects ambiguous non-canonical numeric IPv4 forms such as single-integer, shortened, octal, and hexadecimal notation rather than relying on resolver-specific interpretation.

Monitor model validation uses the static step and enforces `timeoutMs <= 30000`, preserving `CodeValidationFailed` and `http.target` / `http.timeoutMs` field paths. Monitor create/update handlers additionally perform network-aware preflight before persistence. Channel create/update validation performs static and network-aware checks for webhook `target` and non-empty email/SMS/PagerDuty `config.apiBaseUrl`, mapping failures to `target` or `config.apiBaseUrl`.

Execution repeats network-aware validation for every request. This is mandatory for legacy records, DNS changes, records persisted before this policy, and time-of-check/time-of-use safety. A transient DNS failure therefore fails closed at mutation and execution boundaries with a typed, sanitized reason.

Alternative considered: resolve only at configuration time. Rejected because persisted approval does not prevent DNS rebinding or redirect-based SSRF.

Alternative considered: perform DNS inside `HTTPConfiguration.Validate`. Rejected because that model validator is used in persistence and scheduling paths, has no context, and should remain deterministic. Network-aware validation belongs at I/O boundaries.

### Derive address denial from the IANA special-purpose registries

Address classification uses `net/netip`, removes IPv6 zones where applicable, and calls `Unmap` before applying the IPv4 policy so IPv4-mapped IPv6 cannot select a different classification. The package contains a reviewed table derived from the IANA IPv4 and IPv6 Special-Purpose Address Registries and rejects every covered range unless both its `Globally Reachable` and `Forwardable` properties are explicitly true. A missing, false, or not-applicable property fails closed. Valid globally unicast addresses outside a special-purpose registry entry remain eligible, preserving the public-target default.

This rule rejects more than named private ranges. Regression coverage explicitly includes shared/CGNAT `100.64.0.0/10`, benchmarking ranges such as `198.18.0.0/15`, IPv4 and IPv6 documentation ranges, protocol-assignment and reserved ranges, loopback, private-use, link-local, multicast, and unspecified space according to their registry properties. The implementation records the IANA registry retrieval date in source and tests representative boundary addresses so policy-table drift is reviewable. Non-canonical numeric hosts are rejected before DNS so alternate IPv4 spellings cannot bypass literal checks.

AWS metadata/control addresses are covered by denied special-purpose ranges and retained as explicit regression cases: EC2 metadata (`169.254.169.254`, `fd00:ec2::254`, and `instance-data.ec2.internal`), VPC DNS (`169.254.169.253`, `fd00:ec2::253`), Amazon Time Sync (`169.254.169.123`, `fd00:ec2::123`), ECS task credentials (`169.254.170.2`), and EKS Pod Identity (`169.254.170.23`, `fd00:ec2::23`). Local and AWS aliases are normalized case-insensitively with a trailing dot removed, but DNS-result classification remains authoritative for aliases not present in the explicit name set.

For a hostname, all A and AAAA answers must be permitted. Mixed public/blocked results reject the whole hop instead of filtering to public answers, avoiding answer-order bypasses.

Alternative considered: use only `IP.IsGlobalUnicast` or a hand-written list of familiar private ranges. Rejected because globally-unicast syntax does not imply IANA global reachability and forwardability; shared/CGNAT, benchmarking, documentation, and reserved ranges would be easy to omit.

### Pin approved resolution into the dial path

For each original or redirected hop, the executor resolves once, validates every answer, and gives the dialer only the approved IP/port candidates. The transport does not receive the original hostname for resolution, and `Proxy` is nil so `HTTP_PROXY`, `HTTPS_PROXY`, and `NO_PROXY` cannot move enforcement to an unvalidated proxy hop. TLS still uses the original hostname as `ServerName` and HTTP preserves the original Host header.

The dial path validates each candidate immediately before connect and validates `RemoteAddr` after connect; a mismatch closes the connection before HTTP bytes are written. Candidates may be attempted in deterministic resolver order until one succeeds, but no unapproved address may be substituted.

The production transport disables HTTP keep-alives and connection reuse. Each original request and each redirect therefore creates a fresh connection from that hop's current approved resolution and dial decision; no idle connection pool may carry a prior host/address approval into a later request. Per-request policy state is immutable and not shared between concurrent executions, so simultaneous redirect chains cannot overwrite each other's approved address sets, TLS server names, or origin classification.

Alternative considered: resolve, validate, then let the default transport resolve again. Rejected because that is the DNS-rebinding gap this change must close.

Alternative considered: allow environment proxies and validate only the proxy. Rejected because it would no longer prove which target the proxy reaches and would make behavior environment-dependent.

Alternative considered: retain a shared connection pool keyed by approved host/address decisions. Rejected for this change because invalidating pooled connections across DNS and policy changes adds complexity; disabling reuse is the smaller invariant and introduces no service or infrastructure cost.

### Treat each redirect as a fresh policy decision

Use `http.Client.CheckRedirect` with a five-redirect maximum. Each redirect URL is a new policy decision: it passes static validation, receives its own current resolution and pinned dial decision, and cannot inherit another hop's address approval. The executor keeps no cookie jar.

Same-origin redirects may retain the request's configured headers, credentials, and replayable body under normal HTTP redirect semantics. A cross-origin or HTTPS-to-HTTP redirect is rejected when the request is credential-bearing, has any configured/caller-supplied header, or has a request body; those requests are not followed after merely stripping data, including when Go would rewrite the method or omit the body. Credential-bearing classification includes provider or caller secrets placed in authorization state, configured headers, or URL path/query construction. A cross-origin redirect may be followed only for a credential-free request with no configured headers and no body, preserving common header-free monitor redirects; the new target is still bounded and fully revalidated.

Alternative considered: rely on Go's default sensitive-header rules. Rejected because custom monitor and webhook headers can contain secrets under arbitrary names, and 307/308 redirects can replay bodies.

Alternative considered: reject every cross-origin redirect. Rejected because harmless public monitor redirects such as apex-to-`www` should remain usable when no caller data can leak.

### Enforce fixed limits at the shared executor

Production limits are:
- Maximum redirects: 5.
- Maximum consumed response body: 1 MiB, detected by reading through a limit of 1 MiB plus one byte.
- Maximum total request duration: the minimum of the incoming context deadline, caller-configured timeout, and 30 seconds.
- Notification total timeout: 10 seconds.
- TCP connection establishment: 5 seconds.
- TLS handshake: 5 seconds.
- Response-header wait: 10 seconds, still bounded by the total request deadline.

The shared executor closes bodies on all paths. It returns a typed oversized-response failure whenever the extra byte is present, including non-2xx provider responses and monitor body assertions. Check execution evaluates expected status before expected body content as today, but neither path can consume more than the shared limit.

Alternative considered: expose an unlimited response body to callers and document a limit. Rejected because existing callers use `io.ReadAll`; the limit must be structural.

### Use typed policy kinds and domain-specific mappings

`outboundhttp.Error` carries a stable `Kind` and safe phase metadata. Kinds cover invalid URL, scheme rejected, address blocked, resolution failed, redirect blocked, redirect limit, timeout, response too large, and transport failure. Its rendered message never includes raw causes, full URLs, URL components supplied by operators, headers, or bodies. The underlying cause may be retained for internal classification, but domain responses and logs use only the typed kind and explicit sanitized text.

Static and preflight configuration failures map to the existing `shared/errors.CodeValidationFailed` with a field path. Notification send failures map to the existing `CodeNotificationDelivery`; test-send response and audit details store a stable sanitized category, not `err.Error()` or provider bodies.

`ExecutionResult` gains an optional machine-readable outbound failure code while retaining its existing outcome and human-readable `Error`. Check execution maps policy kinds without string parsing. Timeout remains `OutcomeTimeout`; blocked destinations, redirect rejection, oversized responses, and transport failures use `OutcomeError`. Status/body expectation mismatches remain `OutcomeFailure` and preserve their existing messages.

Alternative considered: add every outbound kind to the HTTP API error registry. Rejected because monitor execution failures are result data, not handler errors, and notification APIs already have stable validation and delivery codes. A separate typed policy kind avoids conflating API status mapping with execution diagnostics.

### Route every sender through the shared executor

Email, SMS, webhook, PagerDuty, Telegram send, and Telegram chat-ID detection use injected outbound executors. Provider defaults are still validated and bounded at execution; optional provider base URLs are parsed and joined with provider paths rather than concatenated as unchecked strings. Tests inject an executor or fake resolver/dialer instead of depending on loopback `httptest.Server` access through the production public policy.

Both monitor API test sends and escalation-runtime dispatch construct the same sender registry with production policy wiring. This keeps test delivery and real escalation behavior aligned.

Alternative considered: protect only webhooks because their target is directly operator-controlled. Rejected because provider base URL overrides are equally controllable, redirects can affect defaults, and Telegram credentials appear in URL paths that require the same redaction boundary.

## Risks / Trade-offs

- [Risk] DNS preflight can reject a safe hostname during a transient resolver failure. -> Mitigation: return a typed, field-specific failure and repeat validation on a later request; failing open would undermine the boundary.
- [Risk] Existing private/VPC monitors or test fixtures stop working. -> Mitigation: this change intentionally makes private monitoring a future explicit opt-in; production safe-public behavior gets fakes, while tests inject approved resolver/dialer behavior rather than weakening policy.
- [Risk] A public service using mixed public/private DNS answers is rejected. -> Mitigation: document the default-public requirement and require all answers to be public; selecting only public answers creates bypass ambiguity.
- [Risk] Cross-origin redirects for credentialed monitors or providers stop working. -> Mitigation: fail with a stable redirect category; operators can configure the final endpoint directly without risking header/body replay.
- [Risk] A 1 MiB body cap prevents assertions on very large pages. -> Mitigation: health endpoints should be small, status checks remain available, and the fixed cap prevents memory amplification in Lambda.
- [Risk] Error redaction reduces provider-specific detail. -> Mitigation: retain safe status/category information where available, test secret-bearing paths explicitly, and never persist raw provider payloads.
- [Risk] Disabling environment proxies changes behavior in an environment that expected them. -> Mitigation: direct validated dialing is part of the security invariant; proxy support requires a future explicit trusted-proxy design.
- [Risk] Disabling keep-alive increases connection and TLS setup latency. -> Mitigation: monitor and notification volume is bounded and correctness across DNS/policy changes takes priority; approved-address-aware pooling can be proposed later with explicit invalidation semantics.

## Migration Plan

1. Add `shared/outboundhttp`, its Go workspace entry, typed errors, production limits, and exhaustive fake-based tests.
2. Integrate static URL and timeout validation into monitor configuration, then add network-aware preflight to monitor create/update handlers.
3. Change check execution and both manual/worker wiring paths to use the shared executor and persist sanitized machine-readable failure codes.
4. Integrate channel create/update validation and migrate every notification sender, test-send registry, and escalation registry to the shared executor.
5. Update OpenAPI descriptions/limits and dashboard-facing validation messages only where existing generated contracts or tests require them.
6. Deploy as one application release. Existing records are not rewritten; safe public records continue working, while unsafe or newly rebound records fail closed at execution.
7. Verify with unit tests and existing repository checks. No live metadata, private network, or provider calls are part of verification.

Rollback is code-only: redeploy the previous application version. No schema migration or infrastructure resource must be reversed. Records created while the policy is active retain their existing shape.

## Open Questions

None. The limits and default-public-only posture are fixed by this change; private/VPC support requires a separate proposal and explicit trust model.
