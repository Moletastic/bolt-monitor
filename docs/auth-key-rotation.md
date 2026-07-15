# Authentication Key Rotation

The dashboard authentication key lives in SSM Parameter Store at
`/<service>/<stage>/auth/aes-256-gcm` as an AWS-managed `SecureString`.

Create or rotate it before an auth-enabled deployment:

```sh
SST_STAGE=<stage> SST_TARGET_CONFIG=<local-target-config> make rotate-auth-key
```

The helper generates 32 random bytes in process memory and passes them once to the
AWS CLI. It does not print, write, log, output, or persist the key in source or SST
configuration. This CLI invocation is the only transient process boundary. Rotation
replaces the sole active generation; existing dashboard sessions and authentication
transactions must be treated as invalid and require a fresh sign-in.
