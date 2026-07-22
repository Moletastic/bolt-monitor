## ADDED Requirements

### Requirement: System documents named installation profiles

The repository SHALL document three named installation profiles — default low-cost owner, expected validation, and high-volume stress — with explicit dimensions and explicit support-boundary language. The stress profile SHALL NOT be presented as a default or free-tier claim.

#### Scenario: Operator reads profile documentation

- **WHEN** an operator reads the profile documentation
- **THEN** they find the three named profiles with their dimensions, the default-cost language for the low-cost owner profile, and an explicit disclaimer that the high-volume stress profile is not a default or free-tier usage

### Requirement: System documents a reproducible cost worksheet

The repository SHALL provide a static markdown cost worksheet at `docs/cost-worksheet.md` whose header records the pricing date and region. The document SHALL break down estimated monthly cost per profile by service (`AppTable`, `AuthTable`, Cognito, SSM, Lambda, SQS, API Gateway, logs/alarms, dashboard attribution, PITR), SHALL identify the dominant cost driver per profile, and SHALL include the pricing inputs and formulas used so any reviewer can re-derive the numbers. The low-cost owner profile estimate SHALL NOT be presented as a Free Tier guarantee.

#### Scenario: Operator reads the cost worksheet

- **WHEN** an operator reads `docs/cost-worksheet.md`
- **THEN** they find the pricing date and region in the header
- **AND** per-profile per-service cost breakdowns
- **AND** the dominant cost driver for each profile
- **AND** the pricing inputs and formulas used to derive the numbers
- **AND** an explicit disclaimer that the low-cost owner profile is not a Free Tier guarantee

#### Scenario: Cost worksheet is refreshed

- **WHEN** the worksheet is updated for new pricing
- **THEN** the pricing date in the header is updated
- **AND** the per-profile estimates and dominant driver are recomputed from the documented formulas

### Requirement: Optional stage-attributed AWS Budget is documented and opt-in

The repository SHALL document an optional stage-attributed AWS Budget setup with forecast alert at 80%, actual alert at 100%, alert-only behavior, recipients sourced from deployment configuration (never source-controlled personal addresses), and an explicit no-op default. When budget configuration is absent, the stack SHALL synthesize zero `AWS::Budgets::Budget` resources and SHALL deploy successfully.

#### Scenario: Operator configures a stage budget

- **WHEN** deployment configuration supplies `budgetAmountUsd` and `alertEmails`
- **THEN** one AWS Budget is provisioned scoped to the stage tags with forecast and actual alerts at 80% and 100% respectively
- **AND** recipients come from deployment configuration, not source control

#### Scenario: Operator deploys without budget configuration

- **WHEN** deployment configuration omits budget inputs
- **THEN** the synthesized stack contains zero `AWS::Budgets::Budget` resources
- **AND** the deploy succeeds without budget-related errors

#### Scenario: Budget threshold is reached

- **WHEN** the configured budget threshold is reached or forecasted
- **THEN** alerts are sent to configured recipients
- **AND** the system takes no automatic action; monitoring continues unaffected
