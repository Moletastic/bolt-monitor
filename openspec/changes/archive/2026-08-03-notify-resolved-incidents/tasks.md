## 1. Recovery Delivery

- [x] 1.1 Carry canonical transition identity from transition outbox envelopes into escalation event handling, with a safe legacy fallback.
- [x] 1.2 Resolve unique channels from fired escalation steps and create, claim, and send transition-scoped recovery deliveries before suppressing remaining escalation work.

## 2. Verification

- [x] 2.1 Add regression tests for recovery sends, multiple fired steps, duplicate recovery messages, collision-free identities, and suppression of unfired work.
- [x] 2.2 Run targeted escalation runtime tests and the repository Go test suite.
