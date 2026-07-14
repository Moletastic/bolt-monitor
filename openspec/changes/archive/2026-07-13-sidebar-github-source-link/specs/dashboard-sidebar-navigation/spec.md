## ADDED Requirements

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
