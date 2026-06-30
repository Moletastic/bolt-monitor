## Context

The current system intentionally models probe locations even though only one built-in location exists:

```text
Monitor
  └─ probeLocations: ["iad"]
       └─ scheduler fan-out
            └─ ExecutionWork per probe location
                 └─ ExecutionResult.probeLocationId
                      └─ MonitorStatus.lastProbeLocationId
                           └─ Dashboard/OpenAPI displays location
```

That seam was valuable while multi-regional execution was plausible. The product direction has changed: multi-regional health checks would increase infrastructure complexity and are not planned. The location model now creates false affordances and extra code paths without delivering value.

## Goals / Non-Goals

**Goals:**

- Make the product contract explicitly single-execution-environment.
- Remove requirements that force monitors to carry region or probe-location selections.
- Remove requirements that force execution results/statuses to carry location identity.
- Avoid compatibility shims because there is no production data or external API consumer requirement.

**Non-Goals:**

- Implement backend, dashboard, or OpenAPI changes in this spec-only phase.
- Add a replacement region selector, availability-zone selector, worker selector, or deployment target selector.
- Preserve old `iad` examples or old probe-location wire fields for backwards compatibility.

## Decision

### Decision: Treat execution location as an internal deployment detail

**Choice:** Monitors execute from the system's configured runtime environment. Operators do not select, view, or reason about probe locations.

**Rationale:** This matches the infrastructure scope and removes a misleading product promise. It also simplifies scheduler semantics: one enabled monitor produces one check attempt per due interval.

**Alternatives considered:**

- Rename `iad` to `default`: rejected because it hides the hard-coded region while preserving the fan-out model and future multi-region affordance.
- Keep catalog internally but hide it from UI: rejected because the API and runtime would still carry unnecessary location state.
- Preserve deprecated fields temporarily: rejected because there is no production compatibility requirement.

## Risks / Trade-offs

- **[Risk] Large cross-stack cleanup follows from this decision.** Mitigation: split implementation into backend/runtime and dashboard/docs phases.
- **[Risk] Tests and examples currently rely on location fields.** Mitigation: subsequent changes should update tests around the new one-monitor-one-execution contract rather than performing string-only replacements.
- **[Trade-off] Future multi-region checks would require a new proposal.** This is intentional; the product should not carry latent complexity for an out-of-scope capability.
