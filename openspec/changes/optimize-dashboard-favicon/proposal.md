## Why

Dashboard currently relies on browser defaults because it has no favicon or install metadata. A deliberate, multi-size icon set improves recognition in tabs, bookmarks, mobile home screens, and installed/PWA-like contexts without forcing one low-resolution `favicon.ico` to serve every device.

## What Changes

- Add a canonical dashboard mark and generated raster/vector variants for browser, Apple touch, and Android/Chromium contexts.
- Add Next.js metadata links for favicon, SVG icon, Apple touch icon, and web app manifest.
- Add a manifest with dashboard name, theme/background colors, and safe standalone display metadata.
- Document required source dimensions, formats, transparency, and validation checks so future logo updates remain consistent.

## Capabilities

### New Capabilities

- `dashboard-brand-assets`: Provide optimized dashboard icons and browser/install metadata across supported device contexts.

### Modified Capabilities

<!-- No existing spec requirements change. -->

## Impact

- `apps/dashboard/app/layout.tsx` metadata and new `apps/dashboard/app/manifest.ts` route metadata.
- New static assets under `apps/dashboard/app/` or `apps/dashboard/public/`, depending on Next.js metadata conventions.
- Dashboard build, metadata tests, and asset validation checks.
- No API, persistence, or runtime service changes.
