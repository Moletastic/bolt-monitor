## 1. Dependencies

- [x] 1.1 Add `@devicons/react-devicon` package to `apps/dashboard/package.json`
- [x] 1.2 Add `lucide-react` package to `apps/dashboard/package.json` (if not already present)
- [x] 1.3 Run `cd apps/dashboard && npm install` to install dependencies

## 2. Technology Key Expansion

- [x] 2.1 Expand `TECHNOLOGY_KEYS` array in `apps/dashboard/lib/types.ts` to include Tier 2 keys: mongodb, redis, kafka, docker, apache, javascript, rabbitmq
- [x] 2.2 Update `TechnologyKey` type to reflect expanded keys

## 3. ServiceIcon Component

- [x] 3.1 Create `apps/dashboard/components/service-icon.tsx` with Devicon integration
- [x] 3.2 Map technology keys to Devicon components (e.g., PostgresOriginal for postgres)
- [x] 3.3 Handle fallback for unknown technology keys (generic server icon)
- [x] 3.4 Ensure consistent icon sizing via CSS

## 4. MonitorProtocolBadge Component

- [x] 4.1 Create `apps/dashboard/components/monitor-protocol-badge.tsx`
- [x] 4.2 Render styled text pills for HTTP, HTTPS, TCP, gRPC, DNS
- [x] 4.3 No icon elements — text-only badge styling
- [x] 4.4 Consistent styling with dashboard design system

## 5. Sidebar Navigation Icons

- [x] 5.1 Update `apps/dashboard/components/app-shell.tsx` nav items to include Lucide icons
- [x] 5.2 Add icons for Services (Server/Grid), Monitors (Activity), Incidents (AlertTriangle), Settings (Settings)

## 6. ResourceIcon Migration

- [x] 6.1 Find all usages of `ResourceIcon` in dashboard components
- [x] 6.2 Replace service-related `ResourceIcon` with `ServiceIcon`
- [x] 6.3 Replace monitor-related `ResourceIcon` with `MonitorProtocolBadge`
- [x] 6.4 Remove `ResourceIcon` component after migration is complete

## 7. Verification

- [x] 7.1 Run `cd apps/dashboard && npm run lint` to verify linting passes
- [x] 7.2 Run `cd apps/dashboard && npm run check` or `tsc --noEmit` to verify types
- [ ] 7.3 Verify service list shows Devicon icons for known technologies
- [ ] 7.4 Verify monitor list shows styled text badges for protocols
- [ ] 7.5 Verify sidebar navigation shows Lucide icons
