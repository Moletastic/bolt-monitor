## Purpose

Keep contributor-facing repository commands and top-level directory ownership accurate.

## Requirements

### Requirement: README describes repository command and layout ownership
The README SHALL describe the checked-in infrastructure test command and top-level directories according to their current responsibilities.

#### Scenario: Contributor reads validation commands
- **WHEN** a contributor reads the README command reference
- **THEN** it identifies `make test-infra` as infrastructure and root script test validation

#### Scenario: Contributor reads repository layout
- **WHEN** a contributor reads the README layout table
- **THEN** it lists `tools/`
- **AND** it describes `shared/` as Go domain and backend-platform modules
- **AND** it describes root `scripts/` as repository validation automation
