## 1. Validation adapter

- [x] 1.1 Add `github.com/go-playground/validator/v10` to the Go module that owns monitor API request DTO validation.
- [x] 1.2 Add a small adapter that runs validator on request DTOs and maps failures to `shared/errors.CodeValidationFailed`.
- [x] 1.3 Ensure mapped fields use JSON names rather than Go struct field names.
- [x] 1.4 Add tests for required fields, max length, minimum slice length, nested fields, and deterministic first-error selection.

## 2. Monitor API DTO validation

- [x] 2.1 Add validation tags to monitor API request DTOs, not persisted/domain structs.
- [x] 2.2 Refactor notification channel request validation to use DTO validation for simple field requirements and length checks.
- [x] 2.3 Keep type-specific channel config checks explicit unless a custom validator is clearer and tested.
- [x] 2.4 Refactor escalation policy request validation to use DTO validation for name and path shape checks.
- [x] 2.5 Keep repository-backed channel existence checks explicit.

## 3. Domain boundary

- [x] 3.1 Preserve `shared/monitorconfig` domain `Validate()` methods and `shared/rules` usage.
- [x] 3.2 Ensure domain packages do not import `github.com/go-playground/validator/v10`.
- [x] 3.3 Ensure domain validation still returns typed `VALIDATION_FAILED` errors with `details.field` where required.

## 4. Verification

- [x] 4.1 Run `make test-go-all`.
- [x] 4.2 Run `make lint-go`.
- [x] 4.3 Run existing monitor API invalid-request tests and update expectations only when field/reason text intentionally changes under this spec.
- [x] 4.4 Validate OpenSpec with `openspec validate hybrid-go-input-validation --strict`.
