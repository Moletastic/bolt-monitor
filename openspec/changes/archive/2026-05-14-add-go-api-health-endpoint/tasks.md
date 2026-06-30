## 1. Create Go Health Service

- [x] 1.1 Add Go module and Lambda handler under `services/` for health endpoint.
- [x] 1.2 Implement static `200 OK` JSON response for `GET /api/health`.

## 2. Wire Infrastructure

- [x] 2.1 Extend SST stack with API Gateway resource for public HTTP routing.
- [x] 2.2 Wire `GET /api/health` route to Go Lambda service.

## 3. Document And Verify

- [x] 3.1 Document local and deployed validation flow for health endpoint.
- [x] 3.2 Run repo checks needed for infra and Go handler, then verify endpoint path and response shape.
