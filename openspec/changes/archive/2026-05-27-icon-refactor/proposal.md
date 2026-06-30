## Why

The current `ResourceIcon` component serves dual purposes: displaying service technology (showing 2-char abbreviations like "GO", "PG") and monitor protocol badges (showing "HT", "HS"). These are conceptually different UI elements that should be separate components. Additionally, the technology icon library (Devicon) provides proper icons for service technologies, and the technology key allowlist should be prepopulated to include common backend technologies.

## What Changes

1. **Split `ResourceIcon` into `ServiceIcon` + `MonitorProtocolBadge`** — separate components for service technology icons and monitor protocol badges
2. **`ServiceIcon` uses Devicon** — replaces text abbreviations (GO, PG) with actual technology icons from `@devicons/react-devicon`
3. **`MonitorProtocolBadge` remains text pill** — HTTP/TCP/gRPC/DNS badges styled as text badges (no icons, just styled text)
4. **Prepopulate technology key allowlist** — expand `TECHNOLOGY_KEYS` to include Tier 1 + Tier 2 technologies (mongodb, redis, kafka, docker, apache, javascript, rabbitmq)
5. **Add Lucide icons to sidebar navigation** — replace text-only nav items with Lucide icons + labels

## Capabilities

### New Capabilities

- **service-icon-component**: New `ServiceIcon` component using Devicon for technology display. Replaces text abbreviations with proper technology icons.
- **monitor-protocol-badge-component**: New `MonitorProtocolBadge` component for monitor type display. Text-only pill styling (HTTP, TCP, gRPC, DNS as styled text).
- **sidebar-lucide-icons**: Add Lucide icons to sidebar navigation items.

### Modified Capabilities

- **dashboard-web-app**: Replace `ResourceIcon` usages with `ServiceIcon` for services and `MonitorProtocolBadge` for monitors.

## Impact

- **Dashboard components**: New components added, old `ResourceIcon` removed after migration
- **Dependencies**: Add `@devicons/react-devicon` and `lucide-react` packages
- **Technology catalog**: 7 keys → 14 keys (Tier 1 + Tier 2)
