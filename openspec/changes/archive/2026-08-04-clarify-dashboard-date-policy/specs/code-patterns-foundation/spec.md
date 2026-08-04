## MODIFIED Requirements

### Requirement: No native `Date` in the dashboard
The dashboard SHALL use `date-fns` for parsing, formatting, comparison, arithmetic, and duration calculations. Native `Date` construction from a known epoch value is permitted in `apps/dashboard/**`. Direct `Date()` calls and native `Date` method usage SHALL remain prohibited outside the explicit clock wrapper (`apps/dashboard/lib/clock.ts`) and test setup. ESLint SHALL enforce the prohibited syntax.

#### Scenario: Dashboard constructs a known epoch timestamp
- **WHEN** dashboard code constructs `new Date(epochMilliseconds)`
- **THEN** ESLint permits the expression

#### Scenario: Dashboard reads current wall-clock time
- **WHEN** dashboard behavior requires current time
- **THEN** it uses the `now()` helper in `apps/dashboard/lib/clock.ts`

#### Scenario: Dashboard parses external ISO time
- **WHEN** dashboard code parses an API or user-supplied ISO timestamp
- **THEN** it uses `parseISO` rather than `new Date(string)`

### Requirement: `date-fns` covers every time operation
`parseISO` SHALL parse external ISO timestamps. `formatISO` SHALL format timestamps. Date arithmetic, comparison, formatting, and duration calculations SHALL use `date-fns` functions. Manual `getTime()` / `setTime()` math SHALL be replaced with `date-fns` equivalents.

#### Scenario: Dashboard manipulates a timestamp
- **WHEN** dashboard code adds, compares, formats, or calculates a duration from a timestamp
- **THEN** it uses the corresponding `date-fns` function
