## ADDED Requirements

### Requirement: System supports expanded service category catalog

The monitor configuration model SHALL support a generic service category catalog for service identity and dashboard icon rendering.

#### Scenario: Service uses existing category
- **WHEN** a service is created or updated with one of the existing categories `server`, `database`, `cache`, `http`, `queue`, `container`, or `function`
- **THEN** system accepts the category as valid

#### Scenario: Service uses expanded technical category
- **WHEN** a service is created or updated with category `web`, `api`, `worker`, `scheduler`, `storage`, `search`, `auth`, `payments`, `analytics`, `observability`, `ai`, or `integration`
- **THEN** system accepts the category as valid

#### Scenario: Service uses expanded purpose category
- **WHEN** a service is created or updated with category `media`, `content`, `finance`, `learning`, `gaming`, `commerce`, `messaging`, `support`, `marketing`, `admin`, `security`, `location`, or `social`
- **THEN** system accepts the category as valid

#### Scenario: Service uses unsupported category
- **WHEN** a service is created or updated with a category outside the supported catalog
- **THEN** system rejects the request with a validation error identifying `serviceCategory`

#### Scenario: Service has no category
- **WHEN** a service is created or updated without a service category
- **THEN** system continues to allow the missing category
