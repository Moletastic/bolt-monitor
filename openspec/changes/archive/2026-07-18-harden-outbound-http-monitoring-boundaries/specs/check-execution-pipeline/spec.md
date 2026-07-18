## MODIFIED Requirements

### Requirement: System emits normalized execution result
System SHALL emit a normalized execution result shape for downstream result and status processing. Failed outbound policy operations SHALL include a stable machine-identifiable failure code and a sanitized operator-safe error message.

#### Scenario: Check finishes
- **WHEN** a healthcheck execution completes
- **THEN** system produces normalized result data describing monitor identity, timing, outcome, and protocol-specific details needed downstream
- **AND** the result does not include probe-location or region identity

#### Scenario: Outbound policy rejects execution
- **WHEN** execution rejects a destination, redirect, timeout, oversized response, or transport operation
- **THEN** the normalized result has a non-success outcome, a stable outbound failure code, and a sanitized error message
- **AND** the result contains no monitor headers, URL credentials, sensitive query values, or response body

## ADDED Requirements

### Requirement: HTTP checks execute through the shared outbound policy
System SHALL execute recurring and manual HTTP checks through the same shared public outbound policy, including redirect validation, pinned resolution and dialing, timeout bounds, and bounded response reads.

#### Scenario: Safe public monitor succeeds
- **WHEN** a permitted public target responds within configured bounds and satisfies the monitor expectations
- **THEN** execution preserves the existing successful outcome and status/body assertion behavior

#### Scenario: Persisted target becomes private
- **WHEN** a previously stored hostname resolves to a blocked address at execution time
- **THEN** execution sends no HTTP request and emits a typed blocked-address failure result

#### Scenario: Response body assertion uses bounded input
- **WHEN** execution evaluates `expectedBodyContains`
- **THEN** it reads no more than 1 MiB and fails with the typed oversized-response code if the response exceeds that bound

#### Scenario: Monitor headers do not escape their origin
- **WHEN** a monitor request with configured headers receives a cross-origin or HTTPS-to-HTTP redirect
- **THEN** the redirect is rejected before any request is sent to the redirect target
- **AND** the redirect is not followed by merely stripping the configured headers
