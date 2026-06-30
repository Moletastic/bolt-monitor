## Why

Monitor CRUD alone does not provide actual monitoring value because no checks are executed. Execution pipeline is next core step, but it must include a way to stop or disable periodic monitoring before constant healthchecks can increase cost during development.

## What Changes

- Add execution pipeline that turns monitor definitions into runnable checks.
- Define how enabled monitors are selected, routed to probe locations, and executed for supported protocols.
- Define normalized execution result shape handed off to downstream status/result storage.
- Define mandatory stop/disable control so periodic checks can always be halted.

## Capabilities

### New Capabilities
- `check-execution-pipeline`: Pipeline that selects monitors, routes checks to probe locations, executes healthchecks, and emits normalized execution results.

### Modified Capabilities
- `monitor-configuration`: Clarify that disabled monitors must not be scheduled or executed by the pipeline.

## Impact

- Affects scheduler, worker routing, probe execution, and monitor lifecycle semantics.
- Affects result/status persistence and incident logic that depend on execution output.
- Adds hard operational safety requirement: no periodic execution without reliable stop/disable path.
