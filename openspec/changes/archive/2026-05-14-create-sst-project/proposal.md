## Why

The repository has OpenSpec scaffolding but no runnable application or infrastructure project yet. Creating an SST project establishes the deployment foundation early so future healthcheck services, shared resources, and environments can be built on a consistent serverless workflow.

## What Changes

- Bootstrap a new SST project in the repository using TypeScript.
- Define the initial app structure, environment configuration, and developer commands needed to run and deploy SST locally.
- Add a minimal starter stack so the project can synthesize and serve as the base for future platform resources.
- Document the bootstrap expectations so follow-on changes can add services and UI against the same infrastructure entrypoint.

## Capabilities

### New Capabilities
- `sst-project-bootstrap`: Create and configure the repository's initial SST application, including app entrypoints, baseline stack definition, and local developer workflow.

### Modified Capabilities

## Impact

- Affects repository structure by introducing SST configuration and TypeScript infrastructure source files.
- Adds SST and supporting Node.js dependencies for infrastructure development.
- Establishes the primary infrastructure bootstrap path that later backend and frontend changes will build on.
