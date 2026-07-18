## ADDED Requirements

### Requirement: Dashboard publishes cross-device icon variants

The dashboard SHALL publish a recognizable square brand mark in formats and dimensions appropriate for browser tabs, bookmarks, Apple home screens, and Android/Chromium install surfaces.

#### Scenario: Browser requests favicon
- **WHEN** a browser requests the dashboard root document
- **THEN** document metadata exposes a valid `favicon.ico` containing 16x16, 32x32, and 48x48 images
- **AND THEN** the dashboard exposes an SVG icon variant for browsers that support scalable icons

#### Scenario: Mobile home-screen metadata is consumed
- **WHEN** an iOS or iPadOS browser saves the dashboard to the home screen
- **THEN** metadata references a 180x180 `apple-touch-icon.png` with an opaque intentional background

#### Scenario: Chromium install metadata is consumed
- **WHEN** a Chromium-based browser reads the dashboard manifest
- **THEN** manifest declares 192x192 and 512x512 PNG icons
- **AND THEN** manifest declares a 512x512 maskable icon with essential artwork inside the safe zone

### Requirement: Dashboard exposes coherent web app identity metadata

The dashboard SHALL expose a web app manifest containing stable `name`, `short_name`, `start_url`, `scope`, `display`, `theme_color`, `background_color`, and icon purpose metadata consistent with dashboard branding.

#### Scenario: Manifest is requested
- **WHEN** a client requests the dashboard manifest
- **THEN** response has `application/manifest+json` content type
- **AND THEN** manifest `start_url` and `scope` remain within dashboard origin
- **AND THEN** manifest does not imply offline functionality through a service worker requirement

#### Scenario: Metadata links are generated
- **WHEN** Next.js renders dashboard document metadata
- **THEN** document links include favicon, SVG icon, Apple touch icon, and manifest references
- **AND THEN** links do not point to missing or duplicate icon assets

### Requirement: Brand assets remain legible and technically valid

Icon assets SHALL use a simplified high-contrast mark, correct file formats and MIME-compatible extensions, lossless PNG compression, and dimensions declared by their consuming metadata.

#### Scenario: Asset validation runs
- **WHEN** dashboard asset validation executes
- **THEN** it verifies required files exist, PNG dimensions match their contract, ICO contains required sizes, SVG has a square viewBox, and manifest references resolve

#### Scenario: Mark renders at favicon scale
- **WHEN** the mark is rendered at 16x16 pixels
- **THEN** its primary shape remains distinguishable without relying on text or thin decorative detail
