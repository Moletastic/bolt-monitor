## ADDED Requirements

### Requirement: Public outbound policy accepts only HTTP and HTTPS URLs
The system SHALL apply one shared default public-network policy to operator-configurable HTTP monitor targets, webhook URLs, and notification provider base URL overrides. The policy SHALL accept only absolute `http` or `https` URLs with a hostname and SHALL reject URLs containing user information.

#### Scenario: Safe public HTTPS URL is accepted
- **WHEN** the policy validates `https://status.example.com/health` and every resolved address is permitted
- **THEN** validation succeeds

#### Scenario: Safe public HTTP URL is accepted
- **WHEN** the policy validates `http://status.example.com/health` and every resolved address is permitted
- **THEN** validation succeeds

#### Scenario: Non-HTTP scheme is rejected
- **WHEN** the policy validates an absolute URL using `file`, `ftp`, `gopher`, or any scheme other than `http` or `https`
- **THEN** it returns a typed invalid-URL or disallowed-scheme error before network access

#### Scenario: URL embeds credentials
- **WHEN** the policy validates an HTTP or HTTPS URL containing user information
- **THEN** it returns a typed validation error without including the credentials in the error

### Requirement: Public outbound policy rejects non-public destinations
The default public-network policy SHALL derive special-purpose address classification from the IANA IPv4 and IPv6 Special-Purpose Address Registries. It SHALL reject every address in a registry range unless both `Globally Reachable` and `Forwardable` are explicitly true; false, missing, or not-applicable properties SHALL fail closed. Valid globally unicast addresses outside a special-purpose registry entry MAY be accepted. The policy SHALL normalize IPv4-mapped IPv6 addresses before classification and SHALL explicitly fail closed for AWS metadata and runtime control endpoints, including their literal and hostname-alias forms.

#### Scenario: Blocked literal address forms are rejected
- **WHEN** a URL host is any IPv4, IPv6, integer-equivalent parsed IP form, or IPv4-mapped IPv6 representation of a blocked destination
- **THEN** the policy returns a typed blocked-address error before sending a request

#### Scenario: AWS metadata or control endpoint is rejected
- **WHEN** a destination is the EC2 metadata endpoint, ECS task credential endpoint, EKS Pod Identity endpoint, their IPv6 equivalents, or a hostname alias such as `instance-data.ec2.internal`
- **THEN** the policy returns a typed blocked-address error before sending a request

#### Scenario: Hostname resolves to a blocked address
- **WHEN** any A or AAAA answer for a hostname is not both globally reachable and forwardable under the IANA-derived policy
- **THEN** the entire destination is rejected even if another answer is public

#### Scenario: Localhost alias is rejected
- **WHEN** a hostname is a case or trailing-dot variant of `localhost` or another recognized local-host alias, or resolves to a loopback address
- **THEN** the policy returns a typed blocked-address error

#### Scenario: Shared CGNAT space is rejected
- **WHEN** a literal, mapped address, or DNS answer is in `100.64.0.0/10`
- **THEN** the policy returns a typed blocked-address error before dialing

#### Scenario: Benchmarking and documentation ranges are rejected
- **WHEN** a literal or DNS answer is in an IANA benchmarking, documentation, protocol-assignment, or reserved range whose `Globally Reachable` or `Forwardable` property is not true
- **THEN** the policy returns a typed blocked-address error before dialing

### Requirement: Resolution and dialing enforce the same address decision
For each request hop, the policy SHALL resolve the destination once through an injectable resolver, reject the hop if resolution fails, returns no addresses, or includes a blocked address, and pass only the approved address set to an injectable dialer without a second hostname resolution. The dial path SHALL validate the selected and connected remote address before application data is sent and SHALL disable environment proxy routing that could bypass target-address enforcement. The production transport SHALL disable keep-alives and connection reuse so every request and redirect connects through its current resolution and dial decision rather than a stale pooled connection. Concurrent requests SHALL keep resolution, approved addresses, origin, and redirect state isolated per execution.

#### Scenario: DNS answer changes before dial
- **WHEN** an approved hostname would resolve to a blocked address on a later lookup
- **THEN** the request dials only the approved addresses pinned from policy resolution and does not perform the later lookup

#### Scenario: Dialer selects an unapproved address
- **WHEN** the dial path attempts or reports a remote address outside the approved set or in a blocked range
- **THEN** the connection is closed and a typed blocked-address execution error is returned before request bytes are written

#### Scenario: Host has mixed public and private answers
- **WHEN** resolution returns at least one public address and at least one blocked address
- **THEN** no address is dialed

#### Scenario: DNS changes between requests
- **WHEN** a hostname was public for one request and resolves to a blocked address for a later request
- **THEN** the later request performs a new policy resolution and does not reuse the earlier connection

#### Scenario: Stale pooled connection exists
- **WHEN** a prior request established a connection before the hostname's address or policy decision changed
- **THEN** a later request cannot reuse that connection and must pass a fresh resolution and dial decision

#### Scenario: Concurrent redirect chains remain isolated
- **WHEN** concurrent requests to different origins follow redirects with different approved address sets
- **THEN** each redirect uses only its own request's origin and approved addresses and no policy state crosses between requests

### Requirement: Every redirect remains inside the outbound boundary
The policy SHALL validate, currently resolve, and dial every redirect target as a new policy decision, SHALL follow no more than five redirects, and SHALL return a typed redirect failure when a target or limit is rejected. A cross-origin or HTTPS-to-HTTP redirect for any request carrying credentials, configured or caller-supplied headers, or a request body SHALL be rejected rather than followed with those values stripped, including when default HTTP semantics would rewrite the method or omit the body. Only credential-free requests with no configured headers and no body MAY follow cross-origin redirects, and every such hop SHALL remain bounded and fully revalidated.

#### Scenario: Redirect reaches a private destination
- **WHEN** a public target redirects to a URL whose literal or resolved address is blocked
- **THEN** the redirect is not followed and the request returns a typed blocked-redirect error

#### Scenario: Redirect chain exceeds the cap
- **WHEN** a request receives a sixth redirect response
- **THEN** the sixth redirect is not followed and the request returns a typed redirect-limit error

#### Scenario: Header-free monitor redirects to another public origin
- **WHEN** a monitor request without caller-supplied headers or a body redirects to another permitted public origin
- **THEN** the redirect MAY be followed after validation and no target headers are forwarded

#### Scenario: Credentialed request redirects to another origin
- **WHEN** a webhook, provider, or monitor request carrying credentials, any configured header, or a body redirects to another origin or from HTTPS to HTTP
- **THEN** the redirect is rejected before any request is sent to the new target
- **AND** stripping the data does not make the redirect eligible to follow

### Requirement: Outbound requests have fixed resource bounds
The shared policy SHALL cap total request duration at the smaller of the caller deadline, configured request timeout, and 30 seconds. It SHALL also cap connection establishment at 5 seconds, TLS handshake at 5 seconds, response-header wait at 10 seconds, and bytes consumed from any response body at 1 MiB. Notification senders SHALL use a configured total timeout no greater than 10 seconds.

#### Scenario: Monitor timeout exceeds the hard cap
- **WHEN** a monitor is configured with `timeoutMs` greater than 30000
- **THEN** configuration is rejected with a typed validation error for `http.timeoutMs`

#### Scenario: Response exceeds the body cap
- **WHEN** a response body contains more than 1 MiB, including when an expected-body assertion is configured
- **THEN** reading stops at the cap and the request returns a typed response-too-large error

#### Scenario: Request phase times out
- **WHEN** dialing, TLS negotiation, response-header wait, or total request duration reaches its applicable limit
- **THEN** the request is canceled and returns a typed timeout error

### Requirement: Outbound failures are typed and sanitized
The shared policy SHALL return a typed error whose stable kind distinguishes invalid URL, disallowed scheme, blocked address, resolution failure, redirect rejection, redirect limit, timeout, oversized response, and transport failure. Error details SHALL NOT contain full URLs with query values or user information, request or response headers, request or response bodies, credentials, tokens, or provider secrets.

#### Scenario: Secret-bearing provider URL fails
- **WHEN** a provider request whose path, query, or headers contain a credential fails policy validation or transport
- **THEN** the typed error identifies only the safe failure kind and sanitized host-level context

#### Scenario: Caller maps a policy error
- **WHEN** monitor validation, check execution, or notification delivery receives a typed policy error
- **THEN** it can map the stable kind to its domain validation or execution failure without parsing an error string

### Requirement: Private monitoring is not enabled by default
The system SHALL NOT provide an operator switch, configuration flag, or implicit exception that permits private or VPC destinations under this capability.

#### Scenario: Operator supplies a private target
- **WHEN** an operator configures a monitor, webhook, or provider override for a private or VPC address
- **THEN** the default policy rejects it and does not offer an in-request bypass
