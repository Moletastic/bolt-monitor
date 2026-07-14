## ADDED Requirements

### Requirement: Monitor evidence tabs lazy-load and retain history pages
Dashboard SHALL load only history needed for monitor-detail evidence content and retain a loaded tab's records while operator changes tabs within mounted monitor-detail view.

#### Scenario: Operator opens monitor detail
- **WHEN** operator opens monitor detail
- **THEN** dashboard loads newest runs needed for run timeline and metrics
- **AND** dashboard does not eagerly request incidents or audit history unless selected by initial tab state

#### Scenario: Operator selects unloaded evidence tab
- **WHEN** operator selects Incidents or Audit tab not previously loaded in current detail view
- **THEN** dashboard shows loading feedback and requests its newest history page once

#### Scenario: Operator returns to loaded evidence tab
- **WHEN** operator returns to an evidence tab already loaded in current detail view
- **THEN** dashboard renders retained records without issuing another history request

### Requirement: Monitor evidence tabs append older history on demand
Dashboard SHALL provide an outlined `Load more` control below a non-final monitor evidence table and append older history records on activation.

#### Scenario: More monitor history exists
- **WHEN** selected evidence tab response contains continuation cursor
- **THEN** dashboard shows outlined `Load more` control below its table

#### Scenario: Operator loads next history page
- **WHEN** operator activates `Load more`
- **THEN** dashboard disables control with pending feedback
- **AND** appends next returned records without removing existing rows

#### Scenario: History page request fails
- **WHEN** a load-more request fails
- **THEN** dashboard retains already rendered rows
- **AND** dashboard shows actionable retry feedback

#### Scenario: No older monitor history exists
- **WHEN** selected evidence tab response contains no continuation cursor
- **THEN** dashboard does not render `Load more` control
