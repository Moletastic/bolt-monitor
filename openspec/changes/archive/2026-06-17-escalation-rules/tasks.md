## 1. Domain Models and Schema

- [x] 1.1 Define EscalationPolicy, EscalationStep, EscalationPath, ChannelConfig Go types in shared/
- [x] 1.2 Define EscalationState Go type with status ACTIVE/SUPPRESSED/EXHAUSTED
- [x] 1.3 Define BusinessHoursConfig type with timezone, startHour, endHour, daysOfWeek
- [x] 1.4 Add EscalationPolicyID and BusinessHours fields to Service model in monitorconfig
- [x] 1.5 Add new DynamoDB entity types ESCALATION_POLICY and ESCALATION_STATE to dynamodbschema
- [x] 1.6 Add NEW event types: incident.down, incident.up, escalation.exhausted to notifications package
- [x] 1.7 Remove NotificationChannel and MonitorNotificationLink entity types from dynamodbschema (marked for deletion)
- [x] 1.8 Add escalation.exhausted incident type to incident record model

## 2. Escalation Policy CRUD API

- [x] 2.1 Implement escalation policy repository (DynamoDB CRUD for ESCALATION_POLICY entity)
- [x] 2.2 Add POST /api/v1/escalation-policies endpoint handler
- [x] 2.3 Add GET /api/v1/escalation-policies endpoint handler (list)
- [x] 2.4 Add GET /api/v1/escalation-policies/{id} endpoint handler
- [x] 2.5 Add PUT /api/v1/escalation-policies/{id} endpoint handler
- [x] 2.6 Add DELETE /api/v1/escalation-policies/{id} endpoint handler with binding check
- [x] 2.7 Wire escalation policy router to monitor-api handler
- [x] 2.8 Add validation: policy must have at least one step per path
- [x] 2.9 Add validation: each step must have at least one channel

## 3. Service Binding Updates

- [x] 3.1 Add escalationPolicyId field to Service create/update request/response types
- [x] 3.2 Add businessHours field to Service create/update request/response types
- [x] 3.3 Wire escalation policy ID into service repository
- [x] 3.4 Implement policy deletion blocking (check if any service references policy before delete)
- [x] 3.5 Add GET /api/v1/services/{serviceId}/escalation-policy endpoint

## 4. Notify Runtime Generalization

- [x] 4.1 Add Email sender implementation (implement NotificationSender for email)
- [x] 4.2 Add SMS sender implementation (Twilio or similar)
- [x] 4.3 Add Webhook sender implementation (HTTP POST with incident payload)
- [x] 4.4 Add PagerDuty sender implementation (PagerDuty Events API v2)
- [x] 4.5 Refactor notify-runtime to receive escalation step events (instead of direct incident events)
- [x] 4.6 Update SenderRegistry to include all channel types
- [x] 4.7 Test each channel type sender in isolation

## 5. check-runtime Integration

- [x] 5.1 Update check-runtime to emit incident.down event on DEGRADED→DOWN transition (instead of incident.opened)
- [x] 5.2 Update check-runtime to emit incident.up event on RECOVERING→UP transition (instead of incident.resolved)
- [x] 5.3 Update check-runtime to NOT emit notification events directly
- [x] 5.4 Update incidentRecordsForResult to emit incident.down/up to escalation queue instead of calling notification directly
- [x] 5.5 Update SQS queue configuration for escalation handler invocations

## 6. Escalation Execution Engine

- [x] 6.1 Create escalation handler Lambda function
- [x] 6.2 Implement escalation state repository (create/read/update ESCALATION_STATE)
- [x] 6.3 Implement business-hours evaluation logic (given BusinessHoursConfig and current time, return true/false)
- [x] 6.4 Implement step 1 immediate firing (lookup policy, select path, fire first step channels)
- [x] 6.5 Implement CloudWatch scheduled rule creation for delayed steps
- [x] 6.6 Implement escalation handler invoked by CloudWatch (read state, fire next step, reschedule if more)
- [x] 6.7 Implement escalation exhaustion detection (all steps fired, incident still open → create escalation.exhausted incident)
- [x] 6.8 Implement escalation suppression (incident.up received → set ESCALATION_STATE status to SUPPRESSED)
- [ ] 6.9 Verify CloudWatch rules survive Lambda freeze/restart
- [x] 6.10 Add unit tests for escalation state machine (all transitions)

## 7. Dashboard Updates

- [x] 7.1 Add escalation policy list and detail views
- [x] 7.2 Add escalation policy create/edit form with step builder
- [x] 7.3 Add business-hours configuration UI (timezone picker, hour sliders, day-of-week checkboxes)
- [x] 7.4 Add escalation state display on incident detail (show current escalation step, steps fired)
- [x] 7.5 Display escalation.exhausted incidents with visual distinction and original incident link
- [x] 7.6 Add escalation policy assignment to service detail view

## 8. Cleanup

- [x] 8.1 Remove NotificationChannel repository and CRUD endpoints
- [x] 8.2 Remove MonitorNotificationLink repository
- [x] 8.3 Remove NotificationRouter and per-monitor routing logic from notify-runtime
- [x] 8.4 Remove DEFAULT_CHANNEL entity type from dynamodbschema
- [x] 8.5 Remove telegram-only sender hardcoding, verify multi-channel registry works
- [x] 8.6 Run make lint-go and make test-go-all to verify no regressions
