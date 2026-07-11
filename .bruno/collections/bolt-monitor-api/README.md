# Bolt Monitor API Collection

Collection documentation stored next to Bruno request files.

## Requests

1. `Health Check`
2. `Create Monitor`
3. `List Monitors`
4. `Get Monitor`
5. `Get Monitor Status`
6. `List Monitor Runs`
7. `Run Monitor`
8. `Update Monitor`
9. `Disable Monitor`
10. `Enable Monitor`
11. `List Incidents`
12. `Get Incident`
13. `List Monitor Incidents`
14. `Acknowledge Incident`
15. `Resolve Incident`
16. `Get Scheduler Config`
17. `Update Scheduler Config`
18. `Get Monitor Audit`

## Variables

- `apiUrl`: base API URL
- `monitorId`: monitor resource identifier captured after create
- `runId`: manual run identifier captured after run request
- `incidentId`: incident identifier captured after incident list requests

## Example flow

1. Run `Create Monitor`
2. Run `List Monitors`
3. Run `Get Monitor`
4. Run `Get Monitor Status`
5. Run `List Monitor Runs`
6. Run `Run Monitor`
7. Run `Update Monitor`
8. Run `Disable Monitor`
9. Run `Enable Monitor`
10. Run `List Incidents`
11. Run `Get Incident`
12. Run `List Monitor Incidents`
13. Run `Acknowledge Incident`
14. Run `Resolve Incident`
15. Run `Get Scheduler Config`
16. Run `Update Scheduler Config`
17. Run `Get Monitor Audit`
