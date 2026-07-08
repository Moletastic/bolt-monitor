## ADDED Requirements

### Requirement: System renders expanded service technology icons

The dashboard SHALL render and select a generic technology icon for every supported service category.

#### Scenario: Operator scans service category icons
- **WHEN** dashboard surfaces render a service with category `web`, `api`, `worker`, `scheduler`, `storage`, `search`, `auth`, `payments`, `analytics`, `observability`, `ai`, or `integration`
- **THEN** system shows a distinct generic technology icon for that category
- **AND** system shows an understandable category label where text labels are part of the surface

#### Scenario: Operator scans service purpose icons
- **WHEN** dashboard surfaces render a service with category `media`, `content`, `finance`, `learning`, `gaming`, `commerce`, `messaging`, `support`, `marketing`, `admin`, `security`, `location`, or `social`
- **THEN** system shows a distinct generic purpose icon for that category
- **AND** system shows an understandable category label where text labels are part of the surface

#### Scenario: Operator selects service category
- **WHEN** the operator creates or edits a service in the dashboard
- **THEN** system includes every supported service category in the service icon/category selector
- **AND** each selectable category is represented with an icon and understandable label

#### Scenario: Dashboard renders existing category
- **WHEN** dashboard surfaces render a service with existing category `server`, `database`, `cache`, `http`, `queue`, `container`, or `function`
- **THEN** system continues to show the existing category with a technology icon and label

#### Scenario: Dashboard renders missing or unknown category
- **WHEN** dashboard surfaces render a service with no category or a category not recognized by the dashboard catalog
- **THEN** system uses a safe fallback service icon
- **AND** system does not fail rendering the service surface

#### Scenario: Dashboard category catalog matches backend support
- **WHEN** the expanded service category catalog is implemented
- **THEN** dashboard TypeScript category definitions include every backend-supported category value
- **AND** dashboard icon rendering provides a mapping for every dashboard category value
