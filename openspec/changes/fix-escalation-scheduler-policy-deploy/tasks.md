## 1. Scheduler Policy Serialization

- [x] 1.1 Compose notification queue ARN outputs before serializing the EventBridge Scheduler execution-role policy.
- [x] 1.2 Add regression coverage for concrete, least-privilege SQS resources in the rendered policy.

## 2. Verification

- [x] 2.1 Run infrastructure typecheck and tests.
- [x] 2.2 Deploy staging through the lifecycle wrapper and verify the IAM policy succeeds.
