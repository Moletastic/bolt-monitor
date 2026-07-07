## ADDED Requirements

### Requirement: System provides service-list layout controls

The dashboard SHALL present the Services landing page as a scannable service list with health indicators, list controls, and viewport-appropriate service creation affordances.

#### Scenario: Operator opens service list on desktop
- **WHEN** services exist and the operator opens the Services landing page on a wide viewport
- **THEN** system shows a visible `Services` page title and concise service-list description
- **AND** system shows equal-size summary indicators for active services, draft services, and services that are down now to the right of the title area
- **AND** system does not show a duplicated operations card containing lower-fidelity service previews above the service card list

#### Scenario: Operator controls the service list on desktop
- **WHEN** services exist and the operator views the Services landing page on a wide viewport
- **THEN** system shows a controls row above the service cards
- **AND** the controls row includes a search field for narrowing the visible service cards
- **AND** the controls row includes filter affordances for service-list refinement
- **AND** the controls row includes a `Create service` action that navigates to the service creation flow

#### Scenario: Operator opens service list on mobile
- **WHEN** services exist and the operator opens the Services landing page on a narrow viewport
- **THEN** system shows active, draft, and down-now indicators in a single row at the top of the visible content
- **AND** system does not show the desktop page title and description in the visible content flow
- **AND** system preserves an accessible page heading for assistive technology
- **AND** system shows search and filter controls in the next row
- **AND** system renders service cards one per row

#### Scenario: Operator creates a service on mobile
- **WHEN** services exist and the operator views the Services landing page on a narrow viewport
- **THEN** system shows the `Create service` action as a fixed floating action button at the bottom-right of the viewport
- **AND** the floating action button has an accessible name describing that it creates a service
- **AND** the floating action button remains reachable without covering primary service card content

#### Scenario: Service list layout is loading
- **WHEN** the Services landing page is loading
- **THEN** system shows loading placeholders that match the final service-list layout structure for the current viewport
