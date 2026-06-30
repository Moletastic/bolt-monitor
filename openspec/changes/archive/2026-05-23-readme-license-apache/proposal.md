## Why

The root README.md currently uses underselling language ("bootstrap", "future") that misrepresents the project as a scaffold rather than a functional open source platform. A LICENSE file is entirely absent, making the open source claim hollow. This undermines adoption by operators evaluating the tool and fails to clearly communicate what bolt-monitor does, its current scope, and its architecture.

## What Changes

- Replace root `README.md` with a complete, honest, evaluator-focused document that accurately represents the project as an early-stage but functional serverless monitoring platform
- Add `LICENSE` file with Apache License 2.0 full text
- Remove bootstrap/scaffold/future framing throughout README
- Add current capabilities list with honest limitations
- Add architecture diagram
- Add environment variable documentation
- Add quick start guide
- Add roadmap section
- Add development notes with known gotchas
- Add repository layout with semantic descriptions

## Capabilities

### New Capabilities
(None)

### Modified Capabilities
(None)

## Impact

- `README.md` at repo root — completely rewritten
- `LICENSE` file at repo root — new file, Apache 2.0
- No changes to API surface, Lambda behavior, or dashboard functionality
- No spec changes required — this is documentation and license infrastructure