## 1. Radius Tokens

- [x] 1.1 Reduce dashboard Tailwind border-radius token values.
- [x] 1.2 Confirm shared card, button, input, select, dialog, toast, table, and menu components use the shared radius scale.

## 2. Shape Audit

- [x] 2.1 Preserve `rounded-full` usage for status chips, dots, progress bars, and floating action buttons.
- [x] 2.2 Review page-level `rounded-lg` and `rounded-xl` usage for obvious outliers after token reduction.
- [x] 2.3 Avoid introducing one-off radius values unless required for a semantic shape.

## 3. Verification

- [x] 3.1 Run `make lint-dashboard`.
- [x] 3.2 Run `make check-dashboard`.
