## Why

The Next.js dashboard is throwing `NEXT_REDIRECT` errors on every form submit. This indicates incorrect use of Next.js redirect handling in server actions or form submissions. The errors appear as exceptions rather than proper HTTP redirects, breaking the user experience.

## What Changes

- **Fixed**: Form submission handlers to use proper Next.js redirect patterns
- **Fixed**: Server actions to handle redirects correctly without throwing exceptions
- **Verified**: All form submit paths (create service, create monitor, enable/disable, etc.)

## Capabilities

### Modified Capabilities
- `dashboard-web-app`: Fix redirect handling in server actions and form submissions

## Impact

- **Code**: `apps/dashboard` - server actions and form components
