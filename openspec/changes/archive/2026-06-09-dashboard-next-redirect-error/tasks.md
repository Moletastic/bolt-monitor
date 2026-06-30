## 1. Investigate and Identify Affected Files

- [x] 1.1 Search for all server action files in `apps/dashboard`
- [x] 1.2 Search for any try-catch blocks that may be catching redirect errors
- [x] 1.3 Identify all form components that submit data

## 2. Fix Server Actions

- [x] 2.1 Review `apps/dashboard/actions/` server action files
- [x] 2.2 Fix any incorrect redirect() usage (not used as return value)
- [x] 2.3 Remove redirect() calls from try-catch blocks
- [x] 2.4 Add `revalidatePath()` calls where data changes

## 3. Fix Form Components

- [x] 3.1 Review form components in `apps/dashboard/app/`
- [x] 3.2 Ensure forms use correct server action patterns
- [x] 3.3 Verify error handling doesn't catch NEXT_REDIRECT

## 4. Verify All Redirect Paths

- [x] 4.1 Test service creation redirects correctly
- [x] 4.2 Test service update redirects correctly
- [x] 4.3 Test monitor creation redirects correctly
- [x] 4.4 Test monitor update redirects correctly
- [x] 4.5 Test enable/disable monitor redirects correctly
- [x] 4.6 Test any other form submissions

## 5. Build and Deploy

- [x] 5.1 Run `make lint-dashboard` to check for linting issues
- [x] 5.2 Run `make check-dashboard` for TypeScript checks
- [x] 5.3 Build the dashboard: `make build-dashboard`
- [x] 5.4 Deploy to staging: `make deploy-infra`