## Requirements

### Requirement: Monitor history endpoints provide bounded cursor pages
System SHALL return monitor runs, monitor incidents, and monitor audit history newest first in pages of 20 records.

#### Scenario: Operator requests first monitor-history page
- **WHEN** an operator requests a monitor runs, incidents, or audit endpoint without a cursor
- **THEN** system returns at most 20 records ordered newest first
- **AND** system includes a continuation cursor only when older matching records exist

#### Scenario: Operator requests following monitor-history page
- **WHEN** an operator supplies the continuation cursor from a prior response to the same monitor-history endpoint
- **THEN** system returns the next at-most-20 older records without duplicating records from the prior page

### Requirement: Monitor history cursors are opaque and validated
System SHALL treat continuation cursors as opaque values and reject malformed, incompatible, or resource-mismatched cursors with a typed validation failure.

#### Scenario: Client supplies invalid cursor
- **WHEN** a monitor-history request contains an invalid or incompatible cursor
- **THEN** system returns a validation error
- **AND** system does not issue a DynamoDB query using that cursor

### Requirement: Cursor pages do not calculate collection totals
System SHALL not calculate or return a total record count for cursor-paginated monitor history.

#### Scenario: Cursor page has more records
- **WHEN** a cursor-paginated monitor-history response has additional records
- **THEN** response pagination exposes only page size and opaque continuation metadata
- **AND** response does not expose a synthetic page number or total count
