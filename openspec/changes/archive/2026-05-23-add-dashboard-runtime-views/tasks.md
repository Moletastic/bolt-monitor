## 1. API client and types

- [x] 1.1 Add incident types to `lib/types.ts` (Incident, IncidentListResponse)
- [x] 1.2 Add scheduler config types to `lib/types.ts` (SchedulerConfig, SchedulerConfigResponse)
- [x] 1.3 Add probe location types to `lib/types.ts` (ProbeLocation, ProbeLocationResponse)
- [x] 1.4 Add `listIncidents`, `getIncident`, `acknowledgeIncident`, `resolveIncident` to `lib/api.ts`
- [x] 1.5 Add `getSchedulerConfig`, `updateSchedulerConfig` to `lib/api.ts`
- [x] 1.6 Add `listProbeLocations` to `lib/api.ts`
- [x] 1.7 Add `getMonitorIncidents` and `getMonitorAudit` to `lib/api.ts`
- [x] 1.8 Add `triggerManualRun` to `lib/api.ts` for manual run trigger button

## 2. Server actions

- [x] 2.1 Add `acknowledgeIncidentAction` server action to `lib/actions.ts`
- [x] 2.2 Add `resolveIncidentAction` server action to `lib/actions.ts`
- [x] 2.3 Add `updateSchedulerConfigAction` server action to `lib/actions.ts`
- [x] 2.4 Add `triggerManualRunAction` server action to `lib/actions.ts`

## 3. Monitor detail extensions

- [x] 3.1 Add manual run trigger button to monitor detail page (visible only when monitor is enabled)
- [x] 3.2 Add "Runs" tab and "Incidents" tab and "Audit" tab navigation to monitor detail
- [x] 3.3 Implement incidents tab showing list of incidents for current monitor
- [x] 3.4 Implement audit tab showing audit event history for current monitor
- [x] 3.5 Ensure runs tab is still reachable and shows recent run history

## 4. Incidents pages

- [x] 4.1 Create `/incidents` page with incident list, status filter tabs (all / open / closed)
- [x] 4.2 Create `/incidents/{id}` page showing incident detail with ack/resolve buttons when actionable
- [x] 4.3 Add empty states with helpful messages for no incidents found

## 5. Admin and probe location pages

- [x] 5.1 Create `/admin/scheduler` page showing current recurring execution state
- [x] 5.2 Add toggle form to enable/disable recurring execution from scheduler admin page
- [x] 5.3 Create `/locations` page showing all enabled probe locations with ID and display name

## 6. Navigation and app shell

- [x] 6.1 Add "Incidents" nav link to app shell main navigation
- [x] 6.2 Add "Scheduler" or "Admin" nav link to app shell for scheduler page
- [x] 6.3 Verify all new pages accessible and render without errors (empty or with mock data)