## Context

The SST bootstrap is the deployed API route registry. The primary Bruno collection currently contains legacy top-level monitor requests, while the deployed route shape is service-first and includes additional domains. Bruno files already support request tags and docs, but conventions are inconsistent and no local check compares them with bootstrap routes.

## Goals / Non-Goals

**Goals:**

- Make the primary Bruno collection cover every method-and-path route wired in `infra/stacks/bootstrap.ts`.
- Organize requests by API domain and use predictable verb/resource names.
- Require `domain:<domain>` and `operation:<operation>` tags, route-parameter variable names matching bootstrap names, and useful request docs.
- Add deterministic local validation through `make check-bruno`.
- Keep OpenSpec-to-bootstrap gaps visible as a separate report.

**Non-Goals:**

- No production API changes.
- No deployed-route discovery or network calls.
- No requirement to model every OpenSpec scenario in Bruno; initial coverage is one canonical success request per route.
- No permanent exception or ignore list in the first version.

## Decisions

1. **Use bootstrap as route source.** A static validator extracts method/path declarations from `infra/stacks/bootstrap.ts`. This matches deployed SST wiring without requiring AWS access. OpenSpec route gaps are reported separately rather than silently converted into Bruno requirements.

2. **Normalize paths before comparison.** Compare HTTP method plus path template. Bruno variable syntax such as `{{serviceId}}` maps to bootstrap parameters such as `{serviceId}`. Query strings do not create separate endpoint requirements.

3. **Use domain folders and strict metadata.** Requests live under folders matching API domains. Each request uses a name shaped as `Verb Resource`, exact route-variable names, `domain:<domain>` and `operation:<operation>` tags, and docs containing purpose, setup, and expected result.

4. **Delete stale primary requests.** Requests for routes no longer wired in bootstrap are removed from the primary collection. Git history remains the archive; deprecated requests are not retained in the active manual-testing surface.

5. **Keep guard local.** Add `make check-bruno` as the developer-facing entry point. It validates route coverage and metadata for all Bruno collections, without making CI enforcement part of this change.

## Risks / Trade-offs

- [Static parser couples validation to bootstrap syntax] -> Keep route declarations in the existing direct `api.route('METHOD /path', handler)` form and fail clearly when extraction finds unsupported syntax.
- [Bruno format changes could break metadata parsing] -> Read the checked-in Bruno YAML format and validate only stable `info`, `http`, `tags`, `docs`, and URL fields.
- [One success request may miss important failures] -> Keep negative-case requests optional and add them in capability-specific follow-up changes.
- [OpenSpec and bootstrap may disagree] -> Print a separate diagnostic for spec routes absent from bootstrap; do not mark those routes as Bruno drift.

## Migration Plan

1. Inventory bootstrap routes and classify current Bruno requests.
2. Replace legacy top-level monitor requests with service-scoped requests and add missing domain requests.
3. Apply naming, variable, tag, and documentation conventions.
4. Add validator and `make check-bruno`.
5. Document governance in `CONSTITUTION.md`, operational conventions in `AGENTS.md`, and workflow details in Bruno docs.
6. Run the local guard and existing repository checks.

Rollback means removing the local guard and documentation changes; no production rollback is needed.

## Open Questions

- None for initial scope. Auth-specific tags can be added later if API authentication becomes part of the manual collection contract.
