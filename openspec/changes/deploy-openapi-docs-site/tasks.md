## 1. Docs site deployment wiring

- [ ] 1.1 Add an SST-managed static site resource that publishes the existing `openapi/` workspace.
- [ ] 1.2 Expose the hosted docs URL in stack outputs alongside the API URL.

## 2. Deployed OpenAPI configuration

- [ ] 2.1 Add a build or staging step that produces a deployable docs directory from `openapi/`.
- [ ] 2.2 Ensure the deployed `openapi.yaml` uses the real stack API URL for hosted interactive docs.

## 3. Documentation and verification

- [ ] 3.1 Document how to deploy and access the hosted API docs site.
- [ ] 3.2 Verify the deployed docs site serves the OpenAPI contract and both docs entry pages.
