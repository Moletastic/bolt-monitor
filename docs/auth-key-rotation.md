# Authentication Key Rotation

The dashboard authentication key lives in SSM Parameter Store at
`/<service>/<stage>/auth/aes-256-gcm` as an AWS-managed `SecureString`. See
[Authentication Operations](./auth-operations.md#aes-key-rotation) for the
maintenance, audit, lifecycle, and recovery procedure.

Create or rotate it before an auth-enabled deployment:

```sh
SST_STAGE=<stage> SST_TARGET_CONFIG=<local-target-config> make rotate-auth-key
```

The helper generates 32 random bytes in process memory and passes them once to the
AWS CLI. It does not print, write, log, output, or persist the key in source or SST
configuration. This CLI invocation is the only transient process boundary. Rotation
replaces the sole active generation; it has no previous-key fallback or online
re-encryption path. Existing dashboard sessions and authentication transactions are
intentionally invalid and require a fresh sign-in.
