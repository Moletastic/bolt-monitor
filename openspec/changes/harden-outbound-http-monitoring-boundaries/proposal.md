## Why

HTTP monitor targets and notification destinations currently reach the network through permissive default clients, allowing untrusted URLs, redirects, DNS changes, or oversized responses to cross the runtime's intended public-network boundary. A shared, fail-closed outbound policy is needed before these operator-configurable endpoints can be treated as safe public monitoring inputs.

## What Changes

- Add one shared outbound HTTP request policy for monitor targets, webhook destinations, and configurable notification-provider base URLs.
- Accept only absolute `http` and `https` URLs and derive the default deny boundary from the IANA IPv4 and IPv6 Special-Purpose Address Registries: reject every range that is not both globally reachable and forwardable, including shared/CGNAT, benchmarking, reserved, documentation, loopback, link-local, private, multicast, unspecified, and AWS metadata/control destinations in IPv4, IPv6, mapped, alias, redirect, and DNS-rebinding forms.
- Revalidate every redirect and the addresses used at dial time, disable connection reuse so stale pooled connections cannot bypass a current resolution decision, cap redirect count, response bytes, and connection/request phase timeouts, and reject cross-origin redirects for requests carrying credentials, configured headers, or bodies.
- Return typed validation or execution/delivery failures with sanitized details while preserving successful access to safe public HTTP and HTTPS targets.
- Apply the policy to recurring and manual monitor execution and to both notification test sends and escalation delivery without adding an always-on service or recurring infrastructure cost.
- Keep private/VPC target monitoring as an explicit future opt-in and out of this change.

## Capabilities

### New Capabilities
- `outbound-http-request-policy`: Defines the shared public-network URL, resolution, dial, redirect, timeout, response-size, credential-isolation, and typed-failure boundary for operator-configurable outbound HTTP requests.

### Modified Capabilities
- `monitor-configuration`: Restricts HTTP monitor targets to URLs accepted by the shared public outbound policy and preserves typed field validation.
- `check-execution-pipeline`: Executes HTTP checks through the bounded policy and emits sanitized, machine-identifiable execution failures.
- `check-runtime-worker-mode`: Requires worker HTTP execution to use the shared bounded client rather than a timeout-only default client.
- `notification-channel-crud`: Validates webhook targets and configurable provider base URLs against the shared public outbound policy with typed field errors.
- `notification-channel-test-send`: Applies the same outbound policy and sanitized typed failures to real channel test sends.
- `notification-route-channel-reference`: Applies the shared outbound policy when resolved channels are dispatched by escalation runtime.

## Impact

- Affected Go areas include `shared/monitorconfig`, `shared/checkexecution`, `shared/notifications`, a new shared outbound HTTP package, monitor API manual runs and channel validation/test sends, and check/escalation runtime client wiring.
- Existing safe public monitor targets and provider endpoints remain valid; newly created or updated unsafe URL configurations are rejected, and unsafe persisted configurations fail closed when executed.
- Tests require deterministic resolver, dialer, transport, redirect, timeout, and bounded-body fakes, including concurrent redirects, DNS changes, stale connection-reuse attempts, shared/CGNAT space, special-purpose registries, mapped addresses, and AWS metadata/control endpoints; no live network dependency is required.
- No new deployed service, NAT path, proxy, database, queue, scheduler, or always-on cost is introduced.
