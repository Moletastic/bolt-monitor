## 1. Make Mobile Card Clickable

- [ ] 1.1 Wrap mobile `article` card in Link
- [ ] 1.2 Ensure toggle form has `pointer-events: auto` to remain clickable
- [ ] 1.3 Add hover styling to indicate clickable card

## 2. Make Desktop Table Row Clickable

- [ ] 2.1 Wrap table row cells (except action column) in Link
- [ ] 2.2 Ensure action column toggle button remains functional
- [ ] 2.3 Add hover styling to indicate clickable row

## 3. Test Click Behavior

- [ ] 3.1 Click on row (non-action area) navigates to monitor detail
- [ ] 3.2 Click on toggle button enables/disables monitor
- [ ] 3.3 Ensure no conflict between row click and button click

## 4. Build and Deploy

- [ ] 4.1 Run `make lint-dashboard`
- [ ] 4.2 Run `make check-dashboard`
- [ ] 4.3 Run `make build-dashboard`
- [ ] 4.4 Deploy to staging