## ADDED Requirements

### Requirement: Delivery pages use cursor pagination
Incident delivery collection responses SHALL use the standard cursor pagination envelope and SHALL omit total counts.

#### Scenario: Delivery page has continuation
- **WHEN** the delivery collection has more matching records
- **THEN** the response includes `pagination.size` and opaque `pagination.nextCursor`
- **AND** does not calculate or return a collection total
