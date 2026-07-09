## ADDED Requirements

### Requirement: System stores tenant-scoped search index records

The DynamoDB single-table design SHALL store sparse search index records that support low-I/O global resource search for a tenant.

#### Scenario: Search index record is written
- **WHEN** a searchable service, monitor, escalation policy, or notification channel is created or updated
- **THEN** system stores compact search index records under the tenant partition
- **AND** each search index sort key begins with `SEARCH#` followed by a normalized searchable prefix
- **AND** each search index record includes the resource type, resource stable identifiers, safe display label, safe display description, navigation href, icon discriminator, and match metadata

#### Scenario: Search query reads index records
- **WHEN** system searches for a normalized query
- **THEN** system queries DynamoDB with `PK = TENANT#<tenant>` and a `begins_with(SK, SEARCH#<normalized-query>)` key condition
- **AND** system does not scan the table or list all services, monitors, policies, or channels to satisfy the query
- **AND** system uses a bounded read limit before result de-duplication and ranking

#### Scenario: Search index fields are selected
- **WHEN** system builds search index entries for services
- **THEN** searchable service fields include name, normalized service slug, description, service category, lifecycle state, and rollup status
- **WHEN** system builds search index entries for monitors
- **THEN** searchable monitor fields include name, normalized monitor slug, parent service name, HTTP target safe display, monitor type, and enabled state
- **WHEN** system builds search index entries for escalation policies
- **THEN** searchable policy fields include name, normalized policy slug, description, route structure summary, and referenced channel IDs as safe references
- **WHEN** system builds search index entries for notification channels
- **THEN** searchable channel fields include name, normalized channel slug, channel type, and safe target display

#### Scenario: Sensitive fields are excluded from search storage
- **WHEN** system builds search index entries
- **THEN** system excludes monitor headers, expected body text, notification channel config JSON, inline channel config, tenant identifiers, and secret-bearing URL query strings from search index text and search API display text

#### Scenario: Search index records are bounded
- **WHEN** system derives searchable prefixes from a resource
- **THEN** system limits indexed tokens, prefix lengths, and total records per resource to protect write cost and transaction size

#### Scenario: Search index entries are removed
- **WHEN** a searchable resource is deleted
- **THEN** system deletes the search index records associated with that resource
