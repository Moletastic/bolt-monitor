## 1. Backend Metrics Contract

- [x] 1.1 Define service-card metric response types for recent latency, P99, uptime, monitor-up coverage, sample counts, trend points, and unavailable states.
- [x] 1.2 Add repository support to load bounded recent run samples for all monitors under a service without changing raw run storage.
- [x] 1.3 Implement deterministic aggregation rules for success-only average latency, success-only aggregate P99, sample-based recent uptime, current monitor-up coverage, and trend ordering.
- [x] 1.4 Expose metrics through the selected dashboard-oriented service response or service-card metrics endpoint using the shared API response envelope.
- [x] 1.5 Add monitor-api tests for healthy, degraded, down, no-run, no-monitor, and partial-status services.

## 2. Dashboard Service Cards

- [x] 2.1 Extend dashboard API/types to consume the service-card recent metrics contract without using `any`.
- [x] 2.2 Replace or update Services overview cards to show status chip, existing service technology icon, monitor-up coverage, recent average latency, recent aggregate P99, recent uptime, traffic-light row, and compact trend visualization.
- [x] 2.2.a Wrap service technology icon in a square tile colored by service state (up, degraded, down, unknown).
- [x] 2.2.b Render each child monitor as a protocol text badge inside the card body.
- [x] 2.2.c Position the monitor-up counter on the right side of the card and color it by service state.
- [x] 2.2.d Render the services list grid with up to four cards per row on wide viewports.
- [x] 2.3 Render honest draft, no-monitor, no-run, and unavailable metric states without showing fabricated zero values.
- [x] 2.4 Keep card navigation accessible and preserve the existing Services module route and detail links.
- [x] 2.5 Add dashboard tests for metric formatting, empty states, and service-card rendering behavior.

## 3. Verification

- [x] 3.1 Run `make test-go-all`.
- [x] 3.2 Run `make lint-go`.
- [x] 3.3 Run `make check-dashboard`.
- [x] 3.4 Run `make lint-dashboard`.
- [x] 3.5 Run `make test-dashboard`.
