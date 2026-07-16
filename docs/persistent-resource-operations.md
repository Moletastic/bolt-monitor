# Persistent Resource Operations (v1)

## Re-adopt or import

1. Run target preflight with explicit persistent configuration and target
   confirmation. Capture its non-secret account, region, stage, owner, and
   retained inventory. SST `4.14.1` does not expose a safe stack preview
   command; `sst diff` must not be used because it enters the deploy path.
2. Verify each physical identifier and `service`, `stage`, and `owner` tags.
3. Confirm SST `4.14.1` and Pulumi support importing/adopting that resource
   kind. If unsupported, stop. Do not permit an automated replacement.
4. Follow pinned SST/Pulumi import instructions for the specific resource,
   then preview. Preview must show no replacement of retained resources.
5. Apply only after the no-replacement preview and verify the inventory output
   still identifies the same physical resources.

`AppTable` recovery, restore-to-new-table integrity validation, cutover,
rollback evidence, and measured restore drills belong to
`establish-data-recovery-and-capacity-guardrails`.

For retained Cognito, `AuthTable`, and AES key material, capture the auth
inventory and follow the auth-specific recovery and break-glass controls in
[auth-operations.md](./auth-operations.md). `AuthTable` PITR restores into a
recovery table; it does not authorize an unreviewed traffic cutover.

## Retire a persistent installation

1. Record target preflight and fresh retained-resource inventory.
2. Record evidence/backup decision and stop dependent services.
3. Obtain standard target confirmation and separate destructive confirmation.
4. Remove protection only from exact inventoried resources; do not use
ephemeral cleanup or broad name matching.
5. Delete approved resources, then run bounded residual verification by exact
stage ownership. Record non-secret residual identifiers and retry safely if
provider cleanup was partial.

No runbook step clears protection to make an ordinary deploy succeed.
