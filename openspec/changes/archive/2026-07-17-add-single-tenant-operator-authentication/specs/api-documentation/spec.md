## ADDED Requirements

### Requirement: API documentation describes authenticated v1 access
The source-controlled API documentation SHALL define the Cognito Bearer access-token security scheme and required `aws.cognito.signin.user.admin` access scope for every `/api/v1/**` operation and SHALL leave `GET /api/health` explicitly unauthenticated. It SHALL state that ID tokens and dashboard session cookies are not API credentials and that API Gateway pre-integration authentication or scope failures can produce non-envelope `401` or `403` responses.

#### Scenario: Developer reviews a v1 operation
- **WHEN** a developer reads any `/api/v1/**` operation in the OpenAPI contract
- **THEN** the operation requires the Cognito Bearer access-token security scheme and documented access scope

#### Scenario: Developer reviews health operation
- **WHEN** a developer reads `GET /api/health` in the OpenAPI contract
- **THEN** the operation declares no authentication requirement

#### Scenario: Developer reviews auth failures
- **WHEN** a developer reads authentication and authorization response documentation
- **THEN** it distinguishes API Gateway pre-integration non-envelope `401` or `403` responses from application envelope `401` and `403` responses

### Requirement: Bruno documents and exercises direct Cognito authentication
The Bruno collection SHALL document a direct, human-operator Cognito authentication setup that obtains an access token without Cognito managed login, Amplify, dashboard cookies, or committed credentials. Protected requests SHALL source the Bearer token from a secret-capable local environment variable, and collection documentation SHALL explain invitation activation, token refresh or reacquisition, and cleanup without printing or persisting tokens in source control.

#### Scenario: Operator prepares Bruno after invitation
- **WHEN** an invited operator follows the Bruno setup documentation
- **THEN** they can complete the supported Cognito challenge flow and provide an access token through local secret configuration

#### Scenario: Bruno sends protected request
- **WHEN** any Bruno request targets `/api/v1/**`
- **THEN** it sends the configured access token as a Bearer credential
- **AND** it retains the route's existing domain and operation tags and documentation conventions

#### Scenario: Bruno sends health request
- **WHEN** the Bruno collection sends `GET /api/health`
- **THEN** the request succeeds without an Authorization header

#### Scenario: Repository is inspected for direct-client secrets
- **WHEN** committed Bruno files and examples are reviewed
- **THEN** no password, temporary password, challenge session, recovery code, access token, ID token, refresh token, or dashboard session value is present
