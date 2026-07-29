## 1. Delivery Claiming

- [x] 1.1 Persist deterministic plan and channel delivery records before provider I/O.
- [x] 1.2 Claim eligible deliveries with fencing token and skip unclaimed or delivered records.
- [x] 1.3 Persist normalized provider outcome before acknowledging queue work.

## 2. Retry Safety

- [x] 2.1 Handle partial multi-channel success without resending delivered channels.
- [x] 2.2 Preserve ambiguous outcomes when provider acceptance cannot be confirmed.

## 3. Verification

- [x] 3.1 Add duplicate SQS, concurrent claim, and post-send failure regression tests.
- [x] 3.2 Run escalation runtime Go tests, race tests, and lint.
