## 1. Spec Decision

- [x] 1.1 Update monitor configuration requirements to remove selected probe locations from required monitor fields.
- [x] 1.2 Remove probe-location catalog validation from monitor configuration requirements.
- [x] 1.3 Update scheduler requirements so one enabled monitor creates one execution request.
- [x] 1.4 Update check result/status requirements so persisted result data does not require probe location identity.
- [x] 1.5 Retire the probe-location catalog capability from the active product model.

## 2. Validation

- [x] 2.1 Run `openspec validate spec-remove-probe-location-concept --strict`.
- [x] 2.2 Review follow-on implementation changes for alignment with this spec decision.
