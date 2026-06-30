## Context

The current `ResourceIcon` component is a text-only badge that shows 2-character abbreviations for service technologies (GO, PG, MY) and monitor protocols (HT, HS). This is a placeholder that will be replaced with proper icons. The decision has been made to use Devicon for service technology icons and keep text pills for monitor protocol badges.

## Goals / Non-Goals

**Goals:**
- Replace service technology text abbreviations with Devicon icons
- Keep monitor protocol badges as styled text (not icons)
- Add Lucide icons to sidebar navigation
- Prepopulate technology allowlist with Tier 1 + Tier 2 keys

**Non-Goals:**
- Replacing monitor protocol badges with icons (text pills only)
- Adding all Tier 3 technologies (kubernetes, aws, azure, etc.)
- Changing the overall design system

## Decisions

### Decision 1: Devicon package choice

**Option A: `@devicons/react-devicon`** — Official React port, named exports like `<GoOriginal />`, MIT license
**Option B: `devicons` original** — Original icon font package, less React-friendly

**Chosen: Option A** — Named exports are cleaner for React components, better type safety.

### Decision 2: Sidebar icon style

**Option A: Minimal line icons** — Clean, unobtrusive
**Option B: Filled icons** — More visual weight
**Option C: Mix (line for nav, filled for status)** — Context-dependent

**Chosen: Option C** — Mix of line and filled Lucide icons for sidebar nav, status indicators use filled variants

### Decision 3: Technology key expansion

**Tier 1 (now):** golang, mariadb, mysql, nginx, postgres, python, typescript
**Tier 2 (add now):** mongodb, redis, kafka, docker, apache, javascript, rabbitmq
**Tier 3 (later):** kubernetes, aws, azure, google-cloud, react, vue

**Chosen: Add Tier 1 + Tier 2 now** — Reasonable set for a backend monitoring tool. Tier 3 is more niche/IaC-focused.

## Component Inventory

### ServiceIcon
- Props: `technologyKey: string`
- Renders: Devicon component for known keys, fallback icon for unknown
- Size: Consistent 24x24 or matching surrounding elements

### MonitorProtocolBadge
- Props: `type: MonitorType` (http, tcp, grpc, dns)
- Renders: Text pill with styled text (HTTP, TCP, gRPC, DNS)
- No icons — just styled text badges

### NavItem (updated)
- Props: `icon: LucideIcon, label: string, href: string, ...`
- Renders: Lucide icon + label for sidebar navigation

## Risks / Trade-offs

[Risk] Devicon icon availability varies by technology
→ **Mitigation**: Provide fallback (generic server icon) for unknown technology keys

[Risk] Icon size consistency across different Devicon icons
→ **Mitigation**: Use consistent sizing via CSS, icons are SVG-based so scale well
