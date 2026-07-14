## Context

The shared dashboard shell renders product navigation from `apps/dashboard/components/app-shell.tsx`. Its sidebar currently ends with an empty flexible footer area. The repository has a stable public GitHub remote, and the requested link is informational rather than an application route.

## Goals / Non-Goals

**Goals:**

- Make the public source repository discoverable from every dashboard route.
- Keep external source access visually and semantically separate from product navigation.
- Preserve keyboard and screen-reader access, including clear external-link behavior.
- Avoid runtime configuration, API changes, or new dependencies.

**Non-Goals:**

- Add GitHub authentication, issue creation, or release links.
- Make the repository URL tenant- or deployment-configurable.
- Add the repository link to active-route matching or global search.
- Redesign the sidebar or change existing module order.

## Decisions

- **Use a fixed repository URL.** The dashboard belongs to one public repository, so a constant URL is smaller and safer than introducing an environment variable with another missing-config failure mode. Deployment or fork support would justify revisiting this later.
- **Render the link in the existing sidebar footer area.** The current `mt-auto` spacer provides a natural utility area. A top border and spacing will distinguish it from module navigation.
- **Use text plus icons.** `lucide-react` already exists. A GitHub mark communicates destination; an external-link indicator communicates that the destination leaves the dashboard. Text remains the primary accessible label.
- **Open in a new tab.** Source inspection is an auxiliary task and should not replace the operator's current console. Use `target="_blank"` with `rel="noreferrer"`.
- **Keep link outside `navItems`.** It is not a dashboard module and must never receive active styling.

## Risks / Trade-offs

- [Repository URL becomes stale] -> Keep the URL centralized in the shell and cover it with an exact-link test.
- [New-tab behavior surprises users] -> Show external-link icon and use explicit accessible link text.
- [Mobile sidebar becomes taller] -> Reuse existing responsive flow and keep utility link compact.

## Migration Plan

No migration or deployment sequencing required. Add the static link, run dashboard lint, typecheck, and tests. Rollback removes the utility link and its focused tests/spec delta.

## Open Questions

None for current single-repository deployment model.
