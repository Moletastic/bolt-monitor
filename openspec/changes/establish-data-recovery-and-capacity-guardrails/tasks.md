## 1. Profiles

- [x] 1.1 Document three named profiles (default low-cost owner, expected validation, high-volume stress) with explicit dimensions, support-boundary language, and explicit disclaimer that the stress profile is not a default or free-tier claim.

## 2. Cost Worksheet

- [x] 2.1 Write `docs/cost-worksheet.md` as a static markdown document with the pricing date and region in the header, the pricing inputs used per service, the resulting monthly cost estimate per profile broken down by service, and the dominant cost driver identified for each profile. Include an explicit disclaimer that the low-cost owner profile is not a Free Tier guarantee.

## 3. Optional Budget Setup

- [x] 3.1 Document optional stage-attributed AWS Budget setup in `docs/optional-budget-setup.md`: forecast at 80%, actual at 100%, no auto-shutdown, recipients from deployment configuration not source-controlled, explicit no-op default.
- [x] 3.2 Add infrastructure wiring under `infra/` for one stage-attributed AWS Budget gated by target configuration (`budgetAmountUsd`, `alertEmails`). Absent inputs yield zero resources and a successful deploy.
- [x] 3.3 Add an infrastructure test that asserts zero `AWS::Budgets::Budget` resources exist in the synthesized stack when budget configuration is absent.

## 4. Final Gates

- [x] 4.1 Run `make test-go-all`, `make check-infra`, `make format-infra`, `make lint-dashboard`, `make check-dashboard`, and `make check-bruno`. Resolve any drift.
