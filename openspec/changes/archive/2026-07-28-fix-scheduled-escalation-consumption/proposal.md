## Why

EventBridge Scheduler sends delayed escalation work to the notification SQS queue, but the queue handler does not recognize the canonical `scheduled_step` payload. Delayed policy steps retry into the DLQ and never notify operators.

## What Changes

- Parse and validate canonical scheduled-step envelopes in the SQS handler.
- Map canonical scheduled identity and step number to the existing scheduled-step execution path.
- Preserve partial-batch failure behavior for malformed or unsupported records.
- Add an end-to-end Scheduler-payload-to-SQS-consumer regression test.
- FinOps: reuse existing Scheduler, SQS, and DLQ. Prevent futile retry/DLQ traffic; add no schedules, queues, metrics, or persistent indexes.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `escalation-one-time-scheduling`: Scheduled queue messages must execute delayed steps through the same canonical queue path.

## Impact

- `services/escalation-runtime/one_time_scheduler.go`
- `services/escalation-runtime/handler.go`
- Escalation SQS handler tests
- No public API, deployment resource, or dependency change
