# Installation Profiles

Bolt Monitor is single-region, single-tenant (`DEFAULT`), serverless, and self-deployed. To keep cost and operational support language unambiguous, the repository recognizes three named installation profiles. They are not service-level commitments. They exist so cost estimates, support boundaries, and any future stress evidence share one vocabulary.

## The three profiles

### Default low-cost owner profile

The baseline used for low-cost cost estimates. This is what a typical owner-operated installation looks like.

| Dimension | Value |
| --- | --- |
| Tenants | 1 active (`DEFAULT`) |
| Services | up to 10 |
| Monitors | up to 10 |
| Monitor cadence | 5 minutes |
| Operator request concurrency | modest (single operator) |
| Raw run retention | 30 days (`CheckRun` TTL) |
| Transient work retention | 7 days (`ExecutionWork` TTL) |
| Region | single AWS region |

> The low-cost owner profile estimate is not an AWS Free Tier guarantee. Account credits, region, log volume, PITR, Cognito usage, and other services affect charges independently.

### Expected validation profile

The routine non-production validation target. Used when staging or a dedicated validation stage is exercising representative load before any release touches production.

| Dimension | Value |
| --- | --- |
| Tenants | 1 active (`DEFAULT`) |
| Services | up to 100 |
| Monitors | up to 100 |
| Monitor cadence | 1 minute |
| Operator request concurrency | up to 10 concurrent |
| Raw run retention | 30 days |
| Transient work retention | 7 days |
| Region | single AWS region |

### High-volume stress profile

A deliberate stress exercise, not a default. The repository may run evidence against this profile to prove bounded access paths and conservative failure modes. It is explicitly not a default-cost claim, not a free-tier expectation, and not a service-level commitment.

| Dimension | Value |
| --- | --- |
| Tenants | 1 active (`DEFAULT`) |
| Services | up to 100 |
| Monitors | up to 1000 enabled |
| Monitor cadence | 1 minute |
| Operator request concurrency | up to 25 concurrent |
| Raw run retention | 30 days |
| Transient work retention | 7 days |
| Region | single AWS region |

## Support-boundary language

Crossing one named dimension emits an actionable signal and requires fresh evidence before operational support is claimed. It does not automatically reject an otherwise safe configuration.

Hard rejection is reserved for separately documented safety limits (transaction limit, maximum payload or page size, minimum safe cadence, resource invariant). Those limits are stated in the relevant operational or API documentation with the reason.

The repository does not claim unlimited scaling beyond the high-volume stress profile.

## Related documents

- [`cost-worksheet.md`](./cost-worksheet.md) — per-profile cost estimates derived from these dimensions.
- [`optional-budget-setup.md`](./optional-budget-setup.md) — optional stage-attributed AWS Budget to catch drift from these profiles.
