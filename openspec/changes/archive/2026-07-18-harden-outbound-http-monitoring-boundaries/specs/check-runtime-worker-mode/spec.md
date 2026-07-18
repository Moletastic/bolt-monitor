## MODIFIED Requirements

### Requirement: Worker executes HTTP check against target
System SHALL execute the HTTP check as specified in the monitor configuration through the shared public outbound policy.

#### Scenario: HTTP execution
- **WHEN** worker has parsed `ExecutionRequest`
- **THEN** it calls check execution with the monitor's bounded `timeoutMs`
- **AND** the request uses the monitor's HTTP configuration for target, method, headers, expected status codes, and expected body content
- **AND** the runtime injects the shared policy resolver, dialer, redirect controls, phase timeouts, and response-size limit rather than constructing a timeout-only default client

#### Scenario: Worker receives unsafe persisted configuration
- **WHEN** a queued execution contains a target that is now blocked by the public outbound policy
- **THEN** the worker records a typed failed execution result without dialing the blocked destination
