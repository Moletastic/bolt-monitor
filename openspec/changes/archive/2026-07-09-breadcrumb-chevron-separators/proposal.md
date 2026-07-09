## Why

Breadcrumbs should read like app navigation, not file paths. Replacing slash separators with chevrons improves hierarchy scanning and aligns with common dashboard navigation patterns.

## What Changes

- Replace breadcrumb slash/text separators with muted chevron separators.
- Keep separator decoration hidden from assistive technology.
- Preserve existing breadcrumb labels, links, and current-page semantics.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `dashboard-web-app`: Update breadcrumb visual separator behavior.

## Impact

- Affected dashboard breadcrumb component only.
- No route, API, or data behavior changes.
