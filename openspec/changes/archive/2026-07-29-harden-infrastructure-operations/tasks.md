## 1. Make-Compatible Operation Inputs

- [x] 1.1 Extend the operation CLI parser to accept documented `KEY=value` inputs and retain equivalent `--KEY=value` support for known arguments.
- [x] 1.2 Route `DESTROY=yes` only to persistent-removal intent and route non-empty `EMAIL` to administrator invitation without exposing credentials.
- [x] 1.3 Add CLI-dispatch tests covering Make-compatible persistent removal, invitation, and rejected missing or invalid values.

## 2. Persistent Deployment Postflight

- [x] 2.1 Refactor persistent DynamoDB verification to check deletion protection and point-in-time recovery for both output-named `AppTable` and `AuthTable` through bounded exact-table reads.
- [x] 2.2 Parse the existing public health response and require the documented successful health envelope and healthy result without issuing a second request.
- [x] 2.3 Add deploy tests for AppTable/AuthTable protection failures, malformed or unsuccessful health responses, and successful bounded postflight ordering.

## 3. Ephemeral Cleanup Evidence

- [x] 3.1 Preserve and return the exact-stage pre-removal SST inventory as non-secret cleanup evidence.
- [x] 3.2 Require successful exact-stage SST state verification and ownership-tag residual verification before reporting cleanup success; retain bounded diagnostic output on failure.
- [x] 3.3 Add cleanup tests for successful evidence reporting, missing state evidence, deployed-stage residue, and tagged residual resources without AWS credentials.

## 4. Verification

- [x] 4.1 Update lifecycle and budget-operation documentation only where command behavior or cleanup evidence changes.
- [x] 4.2 Run `openspec validate harden-infrastructure-operations --strict`.
- [x] 4.3 Run `make format-infra-check`, `make check-infra`, and `make test-infra`.
