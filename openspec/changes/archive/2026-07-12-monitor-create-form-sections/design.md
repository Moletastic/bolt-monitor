## Context

The dashboard monitor form currently renders one flat grid with identity, HTTP request, validation, and headers controls mixed together. Headers are entered as a multiline `Header-Name: value` textarea, expected status codes are entered as comma-separated text, and `Expected body contains` is exposed even though the next form direction is protocol-based validation.

The create page also includes a side card titled “Create flow notes”. That card documents implementation assumptions rather than helping the operator complete the task. The monitor detail page currently embeds the same `MonitorForm` as an edit surface, but the detail view is due for a separate refactor.

## Goals / Non-Goals

**Goals:**

- Make monitor create/edit flow easier to scan by grouping fields by operator intent.
- Keep HTTP as the only supported protocol in this phase while showing TCP and gRPC as future options.
- Preserve existing HTTP create/update payload shape.
- Replace freeform header text with structured key/value rows.
- Replace comma-separated expected status codes with tag-style removable selections.
- Remove noisy create-flow notes from monitor creation.
- Remove the current embedded monitor edit form from monitor detail until that view is refactored.

**Non-Goals:**

- Add TCP monitor creation.
- Add gRPC monitor creation.
- Add DNS monitor creation.
- Change monitor CRUD API contracts.
- Change check execution behavior.
- Redesign the whole monitor detail page beyond removing the current edit form.
- Implement expected-body validation UI in this phase.

## Decisions

### Use sections around operator intent

The form should read as setup flow rather than API payload order:

```txt
Create monitor

[ Identity ]
  [Monitor name] [Frequency]

[ HTTP ] [TCP - coming soon] [gRPC - coming soon]

[ Request ]
  [Method] [Target URL] [Timeout ms]
  Headers
  [Key] [Value] [delete]
  [+ Add header]

[ Validation ]
  Expected status codes
  [200 x] [201 x] [selector]
```

This creates a stable structure for future protocol-specific request and validation controls without adding unsupported behavior now.

### Keep HTTP selected and submitted

The protocol tab row is a UI affordance for future protocol support. `HTTP` is preselected. `TCP` and `gRPC` are disabled and should communicate “Coming soon” through visible hint text, tooltip, or equivalent accessible copy. Submissions continue to send `type: 'http'`.

### Treat request and validation as protocol-variable regions

For this phase, only HTTP renders real controls. The `Request` section contains method, target URL, timeout, and headers. The `Validation` section contains expected status codes. Future TCP and gRPC work can swap these regions without redesigning the whole form.

### Store headers through structured rows

Headers should be edited as key/value pairs with one pair per row. Operators can remove rows and add rows with a full-width add-header button. The default header should be `Content-Type: application/json`.

The existing server action can still receive serialized header data in whatever hidden or named fields the implementation chooses, as long as it builds the same `headers?: Record<string, string>` HTTP configuration.

### Represent status codes as selectable badges

Expected status codes should default to `200` and render selected values as removable badge-style tags such as `[200 x]` and `[201 x]`. The selectable options should use a common status code list rather than allowing arbitrary freeform values in this phase.

### Hide expected body contains

`Expected body contains` is not part of this phase. The control should be hidden from the active form. If mentioned, it should be framed as coming soon rather than submitted as a visible operator option.

### Remove current monitor detail edit form

The monitor detail page should no longer render the current embedded `MonitorForm` edit panel. Edit behavior can return later through a separate monitor detail refactor. This avoids polishing an edit surface known to be temporary while still allowing the shared `MonitorForm` to support edit mode where it remains used.

## Risks / Trade-offs

- **Client-side form state increases complexity** -> Keep state local to the form controls and preserve server actions for submission.
- **Disabled protocol tabs may look functional** -> Use disabled state plus coming-soon copy so operators do not expect TCP or gRPC submission.
- **Common status codes may omit a valid use case** -> Start with common options for clearer UX; broader custom input can be a later explicit requirement.
- **Removing detail edit form reduces edit discoverability** -> Accept temporarily because the detail view is planned for a separate refactor.
- **Hidden expected-body validation may block existing edits** -> Preserve existing values where practical, but do not expose the field as an active phase-one control.
