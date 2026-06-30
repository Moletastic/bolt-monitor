# Bolt Monitor API Collection

Collection documentation stored next to Bruno request files.

## Requests

1. `Health Check`
2. `List Probe Locations`
3. `Create Monitor`
4. `List Monitors`
5. `Get Monitor`
6. `Get Monitor Status`
7. `List Monitor Runs`
8. `Run Monitor`
9. `Update Monitor`
10. `Disable Monitor`
11. `Enable Monitor`
12. `List Incidents`
13. `Get Incident`
14. `List Monitor Incidents`
15. `Acknowledge Incident`
16. `Resolve Incident`
17. `Get Scheduler Config`
18. `Update Scheduler Config`
19. `Get Monitor Audit`

## Variables

- `apiUrl`: base API URL
- `monitorId`: monitor resource identifier captured after create
- `runId`: manual run identifier captured after run request
- `incidentId`: incident identifier captured after incident list requests

## Example flow

1. Run `List Probe Locations`
2. Run `Create Monitor`
3. Run `List Monitors`
4. Run `Get Monitor`
5. Run `Get Monitor Status`
6. Run `List Monitor Runs`
7. Run `Run Monitor`
8. Run `Update Monitor`
9. Run `Disable Monitor`
10. Run `Enable Monitor`
11. Run `List Incidents`
12. Run `Get Incident`
13. Run `List Monitor Incidents`
14. Run `Acknowledge Incident`
15. Run `Resolve Incident`
16. Run `Get Scheduler Config`
17. Run `Update Scheduler Config`
18. Run `Get Monitor Audit`

Create/edit monitor flows should read probe-location options from `List Probe Locations` instead of hardcoding picker values.
