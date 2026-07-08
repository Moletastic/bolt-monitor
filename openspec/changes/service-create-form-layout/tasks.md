## 1. Create Page Layout

- [x] 1.1 Remove the “Create flow notes” side card from `/services/new`.
- [x] 1.2 Replace the split grid layout with full-width create page content.
- [x] 1.3 Add a visible page title and concise page description above the form.

## 2. Service Identity Section

- [x] 2.1 Add an icon-labeled `Service identity` form section.
- [x] 2.2 Replace the text-only service category selector with an icon-based service icon/category selector that still submits `serviceCategory`.
- [x] 2.3 Group service icon/category selection and service name in the same row on wider viewports.
- [x] 2.4 Place the description field below the icon/category and service name controls.

## 3. Notifications Section

- [x] 3.1 Add an icon-labeled `Notifications` form section.
- [x] 3.2 Keep notification route selection in the Notifications section.
- [x] 3.3 Move the business-hours switch below notification route selection with explanatory copy.
- [x] 3.4 Preserve timezone, time-window, and day-of-week controls when business hours are enabled.

## 4. Behavior And Verification

- [x] 4.1 Preserve existing create-service submission, error display, and redirect behavior.
- [x] 4.2 Preserve existing service update behavior without create-only page framing.
- [x] 4.3 Confirm no monitor creation behavior is added to the service creation form.
- [x] 4.4 Run `make lint-dashboard`.
- [x] 4.5 Run `make check-dashboard`.
