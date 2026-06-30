## 1. Establish Dashboard App Shell

- [x] 1.1 Create the `apps/dashboard` Next.js + TypeScript application scaffold.
- [x] 1.2 Add shadcn/ui foundation and map `DESIGN.md` tokens into the app theme.
- [x] 1.3 Create base application shell primitives for navigation, page chrome, cards, tables, and status chips.

## 2. Implement Monitor Overview

- [x] 2.1 Build dashboard home using `GET /api/v1/monitors` as the primary read model.
- [x] 2.2 Render status-aware monitor cards or rows showing name, enabled state, current status, last check, duration, and probe location.
- [x] 2.3 Add loading, empty, and API-error states appropriate for an operator dashboard.

## 3. Implement Monitor Detail

- [x] 3.1 Build monitor detail page backed by `GET /api/v1/monitors/{id}`.
- [x] 3.2 Show recent run history using `GET /api/v1/monitors/{id}/runs`.
- [x] 3.3 Present configuration, current status, and recent history as distinct sections.

## 4. Implement Monitor Mutations

- [x] 4.1 Add create-monitor flow backed by `POST /api/v1/monitors`.
- [x] 4.2 Add edit-monitor flow backed by `PATCH /api/v1/monitors/{id}`.
- [x] 4.3 Add enable and disable controls backed by the existing action endpoints.

## 5. Verify Frontend Integration

- [x] 5.1 Verify local dashboard flows against the SST development API.
- [x] 5.2 Run dashboard linting and fix integration or typing issues.
- [x] 5.3 Capture any newly discovered API gaps as explicit follow-on OpenSpec changes instead of expanding v1 implicitly.
