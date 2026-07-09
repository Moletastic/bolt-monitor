## ADDED Requirements

### Requirement: System provides global top-bar search

The dashboard SHALL provide a top-bar search control for finding and navigating to services, monitors, notification routes, and notification channels.

#### Scenario: Operator views dashboard shell
- **WHEN** an operator opens a dashboard page inside the shared shell
- **THEN** system shows a top-bar search input with a search icon
- **AND** the input communicates that it can search services, monitors, routes, and channels

#### Scenario: Operator types search query
- **WHEN** the operator enters at least the minimum number of search characters
- **THEN** system waits for a debounce interval before requesting search results
- **AND** system shows user feedback while results are loading
- **AND** system avoids issuing a request for every keystroke

#### Scenario: Operator sees search results
- **WHEN** matching search results are returned
- **THEN** system displays a clickable result list below or near the search input
- **AND** each result shows a resource-specific icon
- **AND** each result shows primary text and secondary text appropriate to its resource type
- **AND** each result navigates to the returned resource link

#### Scenario: Operator searches with no matches
- **WHEN** search completes with no matching results
- **THEN** system shows an empty search feedback message without navigating away

#### Scenario: Search request fails
- **WHEN** the search API request fails
- **THEN** system shows an actionable error state near the search input
- **AND** system preserves the current page and input value

#### Scenario: Operator uses keyboard search interaction
- **WHEN** the search input or results are focused
- **THEN** system preserves keyboard accessibility for entering a query, reaching result links, dismissing the result list, and navigating via result activation

#### Scenario: Operator uses search on narrow viewport
- **WHEN** the dashboard is viewed on a narrow viewport
- **THEN** system keeps the search input usable without hiding primary navigation or page content
- **AND** result links remain tappable and readable
