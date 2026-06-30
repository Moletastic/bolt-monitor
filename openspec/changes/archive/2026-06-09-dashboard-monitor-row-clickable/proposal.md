## Why

The monitor overview list in the dashboard only has small clickable areas (like the monitor name). Users expect to be able to click anywhere on the row to navigate to the monitor detail view, similar to how most table/list UIs work.

## What Changes

- **Modified**: Monitor overview list items to have the entire row be clickable
- **Modified**: Cursor changes to pointer on hover to indicate clickability
- **Modified**: Visual feedback on hover (background highlight)

## Capabilities

### Modified Capabilities
- `dashboard-web-app`: Monitor list items have full-row clickability for navigation

## Impact

- **Code**: `apps/dashboard` - monitor list component styling
