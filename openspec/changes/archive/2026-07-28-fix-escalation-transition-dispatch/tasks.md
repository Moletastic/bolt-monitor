## 1. Canonical Dispatch

- [x] 1.1 Keep transition outbox records pending after Stream-to-SQS publish.
- [x] 1.2 Load canonical outbox by envelope transition identity and acknowledge only after handling.

## 2. Verification

- [x] 2.1 Add Stream-to-SQS-to-transition-handler regression coverage for identity and acknowledgement ordering.
- [x] 2.2 Run escalation runtime Go tests and lint.
