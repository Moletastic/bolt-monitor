## ADDED Requirements

### Requirement: System exposes recent service-card metrics

System SHALL expose dashboard-oriented recent metrics for service cards derived from persisted check-run samples and current monitor status.

#### Scenario: Service has monitors with recent runs

- **WHEN** a dashboard client requests service-card data for a service with configured monitors and recent `CheckRun` records
- **THEN** the system returns recent average latency, aggregate P99 latency, recent uptime percentage, monitor-up coverage, and trend points for that service
- **AND** the metrics are derived from bounded recent samples rather than long-window SLO history

#### Scenario: Service has monitors but no runs

- **WHEN** a dashboard client requests service-card data for a service whose monitors have no persisted run samples
- **THEN** the system returns an explicit no-data state for latency, P99, uptime, and trend values
- **AND** the response does not report zero uptime as if failures were observed

#### Scenario: Service has no configured monitors

- **WHEN** a dashboard client requests service-card data for a service with no monitors
- **THEN** the system returns an explicit no-monitor state
- **AND** metric values that depend on monitor execution are omitted or marked unavailable

### Requirement: System calculates recent metrics consistently

System SHALL calculate service-card metrics using deterministic recent-sample rules.

#### Scenario: Latency metrics are calculated

- **WHEN** recent run samples include successful and unsuccessful outcomes
- **THEN** average latency and aggregate P99 are calculated from successful run `durationMs` values
- **AND** unsuccessful runs are excluded from latency calculations

#### Scenario: Recent uptime is calculated

- **WHEN** recent run samples include successful and unsuccessful outcomes
- **THEN** recent uptime is calculated as successful run count divided by total sampled run count
- **AND** the response includes enough sample-count context for clients to avoid presenting the value as long-window availability

#### Scenario: Monitor-up coverage is calculated

- **WHEN** service-card metrics are calculated for a service with configured monitors
- **THEN** monitor-up coverage is calculated from current persisted monitor statuses
- **AND** monitors without current status are not counted as up
