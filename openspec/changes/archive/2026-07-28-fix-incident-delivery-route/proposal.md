## Why

The declared incident-deliveries endpoint is intercepted by generic incident detail routing. Operators cannot inspect per-channel delivery outcomes through its documented API route.

## What Changes

- Order incident route dispatch so the deliveries collection route wins before generic incident detail.
- Add handler-level route regression coverage using a realistic API Gateway request.
- Retain existing API route, response envelope, OpenAPI, and Bruno contract.
- FinOps: no AWS resource or persistence change. Prevent client retry loops against wrong response without adding cache, metrics, or logging volume.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `notification-delivery-operations`: Existing incident-scoped delivery read route must be reachable through API dispatch.

## Impact

- `services/monitor-api/handler.go`
- `services/monitor-api` route tests
- No route shape, infrastructure, dependency, or storage change
