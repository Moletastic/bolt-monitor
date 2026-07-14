## ADDED Requirements

### Requirement: System exposes module-oriented sidebar navigation in dashboard
System SHALL render a persistent sidebar in the dashboard application with module-level navigation for `Dashboard`, `Services`, `Integrations`, `Audit Trail`, and `Config`.

#### Scenario: Operator opens dashboard application
- **WHEN** operator loads any dashboard route
- **THEN** system shows sidebar navigation with all five modules in a consistent order

### Requirement: Sidebar navigation indicates active dashboard module
System SHALL reflect the operator's current module in sidebar navigation.

#### Scenario: Operator navigates within one module
- **WHEN** operator opens a route that belongs to one dashboard module
- **THEN** system highlights the corresponding sidebar item as active

#### Scenario: Operator opens monitor operations
- **WHEN** operator opens the moved monitor overview or related monitor workflow routes
- **THEN** system highlights `Services` as the active sidebar module rather than `Dashboard`

### Requirement: Sidebar modules remain navigable before feature detail exists
System SHALL provide routable landing surfaces for sidebar modules even when deeper module functionality is not yet implemented.

#### Scenario: Operator opens module with no detailed feature screen yet
- **WHEN** operator selects sidebar item for a not-yet-expanded module
- **THEN** system navigates to a valid dashboard page that explains module scope instead of failing with missing-route behavior

### Requirement: Dashboard root route remains a valid landing module
System SHALL use the root dashboard route as the `Dashboard` module landing surface even when no deeper dashboard content exists yet.

#### Scenario: Operator opens root dashboard route
- **WHEN** operator navigates to `/`
- **THEN** system shows a valid `Dashboard` page with explicit work-in-progress or empty-state messaging

### Requirement: Sidebar provides source repository access
The dashboard SHALL render an accessible external link to the public `Moletastic/bolt-monitor` GitHub repository in a utility area separate from module navigation.

#### Scenario: Operator opens any dashboard route
- **WHEN** an operator views the shared dashboard shell
- **THEN** the sidebar shows a `View source on GitHub` link
- **AND** the link points to `https://github.com/Moletastic/bolt-monitor`

#### Scenario: Operator activates source link
- **WHEN** the operator activates the source repository link
- **THEN** the repository opens in a new browser tab
- **AND** the current dashboard tab remains available

#### Scenario: Operator reviews sidebar navigation state
- **WHEN** the operator views any route
- **THEN** the source repository link is visually separate from product module navigation
- **AND** the source repository link does not receive module active-state styling

#### Scenario: Operator uses assistive technology
- **WHEN** the operator encounters the source repository link
- **THEN** its accessible name identifies GitHub as the destination
- **AND** its external-link behavior is communicated by visible or accessible context
