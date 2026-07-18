## 1. Brand Source

- [ ] 1.1 Confirm brand colors or obtain replacement logo asset; otherwise create simplified Bolt Monitor square SVG mark.
- [ ] 1.2 Ensure source SVG has square `viewBox`, no external dependencies, sufficient contrast, and no text-dependent detail.

## 2. Icon Derivatives

- [ ] 2.1 Generate `favicon.ico` with embedded 16x16, 32x32, and 48x48 images.
- [ ] 2.2 Generate `icon.svg`, `icon-192.png`, and `icon-512.png` with lossless, dimension-correct output.
- [ ] 2.3 Generate `icon-512-maskable.png` with central safe-zone padding.
- [ ] 2.4 Generate opaque `apple-touch-icon.png` at 180x180.

## 3. Next.js Metadata

- [ ] 3.1 Add App Router icon files and update root metadata to expose favicon, SVG, and Apple touch icon references.
- [ ] 3.2 Add typed `app/manifest.ts` with dashboard identity, same-origin `start_url`/`scope`, colors, and `any`/`maskable` icon purposes.
- [ ] 3.3 Keep metadata compatible with existing Next.js build and avoid duplicate generated links.

## 4. Verification

- [ ] 4.1 Add tests or validation script for required files, image dimensions, ICO sizes, SVG viewBox, manifest references, and content types where testable.
- [ ] 4.2 Run `make lint-dashboard`.
- [ ] 4.3 Run `make check-dashboard`.
- [ ] 4.4 Run `make test-dashboard`.
- [ ] 4.5 Run `make build-dashboard` and inspect generated metadata/assets.
- [ ] 4.6 Verify desktop tab/bookmark, iOS home-screen, Android/Chromium install, and maskable rendering in supported browser tooling.
