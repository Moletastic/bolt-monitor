# Result And Status Model

Shared contracts for persisted healthcheck run history and latest monitor status.

## Two storage concerns

- `CheckRun`: append-only historical execution records
- `MonitorStatus`: mutable latest snapshot for fast reads

## Retention

- raw `CheckRun` items carry TTL
- default retention target: 30 days

## Purpose

This split keeps dashboard and API reads fast while preserving recent execution history for debugging and future incident logic.
