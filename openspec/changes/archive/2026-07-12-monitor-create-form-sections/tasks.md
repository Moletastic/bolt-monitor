## 1. Monitor Create Page Layout

- [x] 1.1 Remove the “Create flow notes” side card from `/services/[serviceId]/monitors/new`.
- [x] 1.2 Replace the split layout with full-width monitor create form content.
- [x] 1.3 Preserve service lookup, breadcrumb, error, and not-found behavior.

## 2. Monitor Form Structure

- [x] 2.1 Split `MonitorForm` into `Identity`, `Request`, and `Validation` sections.
- [x] 2.2 Move monitor name and check frequency into `Identity`.
- [x] 2.3 Add protocol tabs between `Identity` and `Request` with `HTTP` selected.
- [x] 2.4 Render `TCP` and `gRPC` as disabled coming-soon options.
- [x] 2.5 Preserve create/update server action submission behavior and error display.

## 3. HTTP Request Controls

- [x] 3.1 Render method, target URL, and timeout controls in the HTTP `Request` section.
- [x] 3.2 Replace the multiline headers textarea with key/value header rows.
- [x] 3.3 Default new monitors to `Content-Type: application/json`.
- [x] 3.4 Add delete controls for header rows.
- [x] 3.5 Add a full-width add-header button.
- [x] 3.6 Ensure submitted headers still build the existing HTTP configuration payload.

## 4. HTTP Validation Controls

- [x] 4.1 Render expected status codes as removable badge-style tags.
- [x] 4.2 Default expected status codes to `200` for new monitors.
- [x] 4.3 Provide common status code options only for this phase.
- [x] 4.4 Hide `Expected body contains` from the active form and keep it as coming soon.
- [x] 4.5 Ensure submitted expected status codes still build the existing HTTP configuration payload.

## 5. Monitor Detail Edit Surface

- [x] 5.1 Remove the current embedded monitor edit form from the monitor detail page.
- [x] 5.2 Preserve the rest of monitor detail behavior and status display.

## 6. Deprecated Probe Locations

- [x] 6.1 Ensure monitor creation accepts dashboard payloads without `probeLocations`.
- [x] 6.2 Keep the dashboard form free of probe-location fields.

## 7. Verification

- [x] 7.1 Run `make lint-dashboard`.
- [x] 7.2 Run `make check-dashboard`.
- [x] 7.3 Run `make test-dashboard` if touched form behavior has existing tests or new tests are added.
