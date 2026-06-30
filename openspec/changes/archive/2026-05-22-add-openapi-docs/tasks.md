## 1. OpenAPI workspace setup

- [x] 1.1 Create the dedicated `openapi/` workspace for docs assets and local tooling.
- [x] 1.2 Add the local package configuration and scripts needed to serve API docs views from the `openapi/` workspace.

## 2. OpenAPI contract

- [x] 2.1 Author `openapi/openapi.yaml` to cover the current health, probe-location, monitor CRUD, monitor status, and monitor run-history endpoints.
- [x] 2.2 Add reusable schemas and examples for the current request and response payloads exposed by the API.

## 3. Local docs rendering

- [x] 3.1 Add a local Swagger UI page wired to the checked-in OpenAPI document.
- [x] 3.2 Add a local Redoc page wired to the same checked-in OpenAPI document.
- [x] 3.3 Ensure there is one primary local command surface for viewing the API docs.

## 4. Documentation and verification

- [x] 4.1 Update repository documentation with the local API docs workflow and entry points.
- [x] 4.2 Verify the local docs workflow renders both Swagger UI and Redoc from the same OpenAPI file.
