## ADDED Requirements

### Requirement: System uses chevron breadcrumb separators

Dashboard breadcrumbs SHALL use decorative right-chevron separators between breadcrumb items.

#### Scenario: Breadcrumbs render multiple items
- **WHEN** breadcrumbs contain more than one item
- **THEN** system visually separates items with muted right-chevron separators
- **AND** separators are hidden from assistive technology
- **AND** parent links and current-page semantics remain unchanged
