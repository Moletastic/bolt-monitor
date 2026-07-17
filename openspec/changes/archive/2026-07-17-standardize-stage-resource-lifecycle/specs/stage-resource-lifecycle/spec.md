## ADDED Requirements

### Requirement: Every deployable stage has an explicit lifecycle class
The infrastructure system SHALL classify every deployable SST stage as exactly one of `persistent` or `ephemeral` from explicit validated configuration before evaluating or mutating AWS resources. It SHALL NOT infer lifecycle class from an unrecognized stage name or silently fall back when class, owner, service, expected account, expected region, or required class-specific configuration is missing or contradictory.

#### Scenario: Persistent stage is approved
- **WHEN** a caller selects a stage explicitly registered as persistent with complete target configuration
- **THEN** infrastructure evaluation uses the persistent resource policy
- **AND** reports the stage name, class, owner, service, expected account, and expected region

#### Scenario: Ephemeral stage is explicit
- **WHEN** a caller supplies a valid ephemeral stage configuration with `disposable=true`
- **THEN** infrastructure evaluation uses the ephemeral resource policy
- **AND** reports its cleanup or expiration deadline

#### Scenario: Classification is absent or inconsistent
- **WHEN** stage class is missing, unknown, inferred only from a name, or conflicts with its disposal or approval configuration
- **THEN** validation fails before any AWS resource mutation
- **AND** the error identifies the invalid configuration without exposing credentials

### Requirement: Persistent stages protect durable resources
Persistent stages SHALL be restricted to explicitly approved stage names and configuration. Taggable resources SHALL identify `service`, `stage`, and `owner`. `AppTable` and, when provisioned, `AuthTable` SHALL enable point-in-time recovery, DynamoDB deletion protection, and infrastructure retain-on-delete behavior. A Cognito operator user pool SHALL use deletion protection where supported and retain-on-delete behavior, and SSM parameters or SST Secrets holding durable installation material SHALL retain on routine stack removal.

#### Scenario: Persistent application storage is provisioned
- **WHEN** an approved persistent stage provisions `AppTable` or `AuthTable`
- **THEN** PITR, DynamoDB deletion protection, and retain-on-delete are enabled
- **AND** required ownership tags identify the service, stage, and owner

#### Scenario: Persistent identity resources are provisioned
- **WHEN** an approved persistent stage provisions the operator user pool or durable SSM/SST secret material
- **THEN** compatible deletion protection and retain behavior prevent routine stack removal from deleting them
- **AND** no secret value is exposed through tags or outputs

#### Scenario: Persistent name is not approved
- **WHEN** a caller requests the persistent class for a stage name absent from approved configuration
- **THEN** deployment fails before resource mutation

### Requirement: Persistent retained resources remain identifiable and manageable
Every persistent deployment SHALL output a non-secret inventory of intentionally retained resource names, ARNs, or equivalent physical identifiers. The repository SHALL provide a versioned runbook for inventory capture, safe re-adoption or import into infrastructure state, replacement preview, deliberate retirement, removal of protections only with explicit destructive intent, deletion, and residual verification.

#### Scenario: Persistent deployment completes
- **WHEN** infrastructure creates or updates a persistent stage
- **THEN** outputs identify every retained table, user pool, and SSM/SST secret resource without revealing secret values
- **AND** the identifiers can be matched to stage ownership tags and configuration

#### Scenario: Retained resource must be re-adopted
- **WHEN** stack state or a logical resource name changes while the physical resource must survive
- **THEN** the runbook requires identity and tag verification, supported import or adoption, and a no-replacement preview before apply
- **AND** it blocks automated replacement when the pinned tool cannot safely adopt that resource kind

#### Scenario: Persistent stage is deliberately retired
- **WHEN** an authorized operator decides to remove a persistent stage
- **THEN** the runbook requires an inventory and data-preservation decision before protections are changed
- **AND** deletion and residual verification target only the explicitly named resources

### Requirement: Ephemeral stages are disposable and cannot impersonate protected stages
An ephemeral stage SHALL require an explicit disposable acknowledgement, a bounded cleanup or expiration deadline, and a stage name that cannot equal or normalize to a protected production or approved persistent name. Ephemeral resources SHALL use no retain-on-delete behavior and no service deletion protection.

#### Scenario: Protected name is requested as ephemeral
- **WHEN** an ephemeral stage name equals or normalizes to `prod`, `production`, or an approved persistent stage name
- **THEN** validation rejects it before AWS resource mutation

#### Scenario: Ephemeral stage provisions stateful resources
- **WHEN** an explicitly disposable stage provisions a table, user pool, parameter, secret, queue, bucket, schedule, or another removable resource
- **THEN** the resource has no retain-on-delete or deletion protection
- **AND** taggable resources carry stage ownership and lifecycle metadata

#### Scenario: Ephemeral expiry is omitted
- **WHEN** an ephemeral target has no bounded cleanup or expiration deadline
- **THEN** validation fails rather than creating an indefinitely disposable stage

### Requirement: Ephemeral cleanup leaves no stage-owned resources
Ephemeral cleanup SHALL run after successful, failed, or cancelled credentialed workflows, SHALL be idempotent, and SHALL verify removal rather than relying only on the SST remove exit status. Cleanup SHALL leave zero stage-owned Cognito user pools, DynamoDB tables, SSM parameters or SST Secrets, EventBridge schedules, SQS queues, S3 buckets, functions, APIs, dashboard resources, log groups, subscriptions, or SST-managed supporting resources. Native TTL, message retention, log retention, and expiration controls SHALL remain bounded where applicable but SHALL NOT substitute for resource removal.

#### Scenario: Ephemeral workflow succeeds or fails
- **WHEN** a credentialed ephemeral deploy workflow reaches completion, failure, or cancellation handling
- **THEN** cleanup attempts removal of all resources owned by that stage
- **AND** performs a bounded residual inventory after removal

#### Scenario: Cleanup finds an orphan
- **WHEN** residual inventory finds a stage-owned resource
- **THEN** cleanup reports failure and the non-secret physical identifier without hiding the original workflow result
- **AND** an idempotent retry or bounded manual cleanup procedure is available for that stage

#### Scenario: Ephemeral cleanup completes
- **WHEN** cleanup reports success
- **THEN** residual inventory contains none of the covered stage-owned resource kinds

### Requirement: Credentialed mutations confirm the effective AWS target
Before deploy, removal, import, adoption, or protection changes, tooling SHALL resolve the effective AWS caller account and region and compare them with explicit expected configuration. It SHALL present application, stage, lifecycle class, owner, account, region, and profile or credential-source label and require a non-interactive confirmation bound to that target. Persistent removal or protection changes SHALL require separate destructive intent from ordinary deploy confirmation.

#### Scenario: Effective target matches
- **WHEN** the resolved caller account and region match expected configuration and confirmation identifies the same target
- **THEN** the requested infrastructure mutation may proceed
- **AND** no credential or secret value is printed

#### Scenario: Account or region differs
- **WHEN** the resolved AWS account or region differs from expected configuration
- **THEN** tooling fails before resource mutation with the mismatched non-secret identifiers

#### Scenario: Persistent resource destruction is requested
- **WHEN** an operator requests persistent removal or disables a protection
- **THEN** ordinary deployment confirmation is insufficient
- **AND** tooling requires explicit destructive intent for the identified persistent target and resources

### Requirement: Local, staging, and credentialed smoke workflows declare lifecycle intent
The supported workflows SHALL document `staging` as a named persistent stage only when it is explicitly approved for long-lived shared validation. Local SST development SHALL explicitly select either an approved persistent target or a developer-owned ephemeral target and SHALL NOT gain a lifecycle class from an omitted stage. Credentialed smoke SHALL use either an explicitly ephemeral stage with verified cleanup or the named long-lived persistent staging stage; it SHALL NOT create a unique retained stage.

#### Scenario: Developer starts local SST
- **WHEN** a developer starts SST local mode
- **THEN** the command requires an explicitly configured persistent or ephemeral target
- **AND** an ephemeral local target is subject to the same verified cleanup contract

#### Scenario: Smoke uses an isolated stage
- **WHEN** credentialed smoke creates a unique stage for the current revision
- **THEN** that stage is explicitly ephemeral and disposable
- **AND** always-run cleanup verifies zero residual resources

#### Scenario: Smoke uses long-lived staging
- **WHEN** credentialed smoke targets the approved persistent `staging` installation
- **THEN** it does not run ephemeral teardown or create another retained stage

### Requirement: Lifecycle protection precedes retained authentication and detailed recovery
The shared stage classification, basic persistent `AppTable` PITR/deletion/retain protection, ephemeral disposal behavior, ownership tags, retained inventory, and lifecycle runbooks SHALL be active before authentication route cutover introduces retained auth resources. Authentication resources SHALL consume this lifecycle policy. Detailed table restore drills, integrity validation, capacity evidence, and recovery cutover SHALL remain governed by the recovery and capacity capability.

#### Scenario: Authentication infrastructure is prepared
- **WHEN** `AuthTable`, Cognito, or durable SSM/SST secret resources are added before protected API cutover
- **THEN** their protection or disposal settings derive from the validated stage class
- **AND** an ephemeral auth deployment does not retain those resources

#### Scenario: Recovery drill is planned
- **WHEN** operators define restore-to-new-table validation, recovery cutover, or measured drill evidence
- **THEN** they use the persistent lifecycle inventory and protection baseline
- **AND** the recovery and capacity capability remains the source of detailed drill requirements

### Requirement: Lifecycle policy limits orphan cost without a new service
The lifecycle implementation SHALL prevent unique disposable stages from retaining fixed- or usage-based AWS resources and SHALL document protection and residual-resource cost effects. It SHALL use deploy/remove workflows, native resource settings, state, tags, outputs, and bounded verification rather than adding an always-on cleanup service or new AWS service solely for lifecycle classification.

#### Scenario: Lifecycle cost is reviewed
- **WHEN** reviewers evaluate persistent and ephemeral resource behavior
- **THEN** documentation identifies retained storage/identity cost and orphan risk by stage class
- **AND** no new always-on lifecycle service is required

#### Scenario: Ephemeral smoke is selected
- **WHEN** an isolated credentialed smoke stage is used
- **THEN** no retained table, user pool, parameter, queue, bucket, schedule, or generated supporting resource remains to accrue orphan cost after verified cleanup
