# Check Execution Contract

Shared execution contract for selecting runnable monitors, routing them to probe locations, and producing normalized healthcheck results.

## Safety rules

- disabled monitors must not execute
- recurring execution must not be enabled unless a reliable stop control exists

## Core concepts

- `ExecutionRequest`: one monitor at one probe location for one trigger type
- `ExecutionResult`: normalized output for downstream result/status persistence
- `SchedulerConfig`: guards recurring execution against missing stop control

## Current scope

- HTTP execution path implemented
- vendor-neutral probe-location routing
- result shape ready for future `CheckRun` and `MonitorStatus` persistence
