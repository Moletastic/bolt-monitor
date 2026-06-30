## Context

The root `README.md` was written during initial project scaffolding. It uses language like "bootstrap", "scaffold", and "future" that undersells the project as incomplete rather than a functional early-stage open source platform. No `LICENSE` file exists, despite the project positioning itself as open source. This creates a credibility gap for evaluators trying to understand what bolt-monitor is, what works today, and how to run it.

The intended audience for this README is operators and engineers evaluating bolt-monitor for self-hosted uptime/monitoring use, similar to Uptime Kuma but tighter AWS-native infrastructure.

## Goals / Non-Goals

**Goals:**
- Communicate what bolt-monitor is, its current capabilities, and its current limitations honestly
- Provide a clear path from zero to running locally
- Serve evaluator/adopter mindset, not contributor funnel
- Add Apache 2.0 LICENSE to formalize open source posture
- Surface known development gotchas prominently

**Non-Goals:**
- Changing any implementation, API surface, or behavior
- Creating contributor or community infrastructure (no CONTRIBUTING.md, issue templates, etc.)
- Elaborate roadmap beyond near-term milestones
- Production hardening beyond current state
- Multi-language README translations

## Decisions

### 1. Apache 2.0 over AGPL-3.0 or MIT

**Decision:** Use Apache 2.0

**Rationale:** bolt-monitor is a self-hosted support/ops tool, similar in posture to Uptime Kuma. For this category, adoption friction matters more than maximum reciprocity. Apache 2.0 provides:
- Strong legal footing for a platform-grade application
- Explicit patent grant
- Low enterprise/internal approval friction
- Credible open source signal without copyleft pressure

AGPL-3.0 was considered but rejected because the added friction outweighs the reciprocity benefit for a tool that teams deploy internally. MIT was rejected in favor of Apache 2.0 because the project is infrastructure/application-grade, not a small library, and deserves more complete legal language.

### 2. README structure optimized for evaluators

**Decision:** Organize README around evaluator journey: what → why → architecture → quick start → reference

**Rationale:** Primary readers are engineers evaluating the tool for internal use. They need to understand scope fast, run it locally, and assess limitations. Contributor-focused sections (contributing guidelines, community norms) are excluded since this is not a community project yet.

### 3. Explicit current limitations section

**Decision:** List known limitations and partial features prominently in README

**Rationale:** For early-stage OSS, radical honesty about what works and what does not creates trust and prevents wasted evaluation cycles. Partial features like single-region probes and unwired runtime surfaces are documented explicitly.

### 4. ASCII architecture diagram

**Decision:** Include simple ASCII diagram showing Browser → Dashboard → API Gateway → Lambdas → DynamoDB

**Rationale:** A visual architecture sketch helps evaluators understand system shape before reading code. It communicates that this is a real distributed system, not a toy project.

## Risks / Trade-offs

[Risk] README overstates maturity if roadmap items never ship
→ Mitigation: Keep roadmap honest, mark items as near-term focus areas, not promises

[Risk] Apache 2.0 LICENSE incompatible with some enterprise policies
→ Mitigation: Apache 2.0 is widely accepted; this risk is low for self-hosted internal tooling

[Risk] README becomes stale as project evolves
→ Mitigation: Treat README as living document; update alongside changes to keep it accurate

[Risk] Missing screenshots/demos reduces visual credibility
→ Mitigation: Consider adding screenshot in future; for now architecture diagram and feature list provide clarity

## Open Questions

None for this change. LICENSE text will use standard Apache 2.0 full license from the Apache Foundation.