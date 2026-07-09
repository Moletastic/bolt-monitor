## ADDED Requirements

### Requirement: System uses sharper dashboard border radius

The dashboard SHALL use a reduced border-radius scale for general operational surfaces.

#### Scenario: General dashboard surfaces render
- **WHEN** system renders cards, panels, dialogs, menus, inputs, selects, buttons, or feedback banners
- **THEN** system uses the reduced dashboard radius scale
- **AND** surfaces appear less curved than the previous dashboard style

#### Scenario: Semantic round shapes render
- **WHEN** system renders status chips, traffic-light dots, circular icons, progress bars, or floating action buttons
- **THEN** system preserves intentional pill or circle shapes

#### Scenario: Radius scale is configured
- **WHEN** dashboard styling is configured
- **THEN** the Tailwind border-radius tokens are reduced from the previous scale
- **AND** components use shared tokens instead of one-off radius values where practical
