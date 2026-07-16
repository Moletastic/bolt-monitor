# Bolt Monitor API Collection

Collection documentation stored next to Bruno request files. Requests mirror every route wired in `infra/stacks/bootstrap.ts`.

## Domains

- `health`: API availability
- `search`: global resource search
- `channels`: notification channel CRUD and delivery test
- `policies`: escalation policy CRUD
- `services`: service CRUD, service incidents, and service audit
- `monitors`: service-scoped monitor CRUD, status, runs, manual run, incidents, audit, enable, and disable
- `incidents`: incident reads and state transitions
- `admin`: scheduler configuration

## Conventions

- Request names use `Verb Resource`.
- Route variables match API names: `serviceId`, `monitorId`, `runId`, `incidentId`, `channelId`, and `policyId`.
- Every request has exactly one `domain:<domain>` tag and one `operation:<operation>` tag.
- Every request docs block contains `Purpose`, `Setup`, and `Expected result`.
- Run `make check-bruno` after route or request changes.

## Variables

- `apiUrl`: base API URL, configured in ignored `environments/development.local.yml`
- `accessToken`: Cognito access token, configured only in ignored `environments/development.local.yml`
- `serviceId`: service identifier
- `monitorId`: monitor identifier under `serviceId`
- `runId`: manual run identifier
- `incidentId`: incident identifier
- `channelId`: notification channel identifier
- `policyId`: escalation policy identifier

## Direct Cognito Authentication

All versioned request folders inherit Bearer authentication from `{{accessToken}}`.
`health/Health Check` is deliberately public and has no Authorization header. Use the
no-secret direct-operator Cognito app client, not the dashboard client or a dashboard
cookie.

Keep the following values only in the ignored
`.bruno/environments/development.local.yml` file:

```yaml
vars:
  apiUrl: https://replace-with-api-url
  cognitoRegion: us-east-1
  cognitoClientId: replace-with-direct-operator-client-id
  cognitoUsername: operator@example.com
  cognitoPassword: replace-with-password-or-temporary-password
  cognitoNewPassword: replace-with-new-password
  accessToken: replace-with-cognito-access-token
```

The direct client has no client secret. Do not place a `ClientSecret`, password,
challenge session, access token, ID token, refresh token, or dashboard cookie in a
versioned Bruno file.

1. Start password authentication with the local `cognitoUsername` and
   `cognitoPassword` values:

   ```sh
   aws cognito-idp initiate-auth \
     --region <cognitoRegion> \
     --client-id <cognitoClientId> \
     --auth-flow USER_PASSWORD_AUTH \
     --auth-parameters USERNAME=<cognitoUsername>,PASSWORD=<cognitoPassword>
   ```

2. If Cognito returns `NEW_PASSWORD_REQUIRED`, respond with the returned `Session`
   value and the local `cognitoNewPassword` value. Keep that Session local and
   transient:

   ```sh
   aws cognito-idp respond-to-auth-challenge \
     --region <cognitoRegion> \
     --client-id <cognitoClientId> \
     --challenge-name NEW_PASSWORD_REQUIRED \
     --session <returned-session> \
     --challenge-responses USERNAME=<cognitoUsername>,NEW_PASSWORD=<cognitoNewPassword>
   ```

3. Copy only `AuthenticationResult.AccessToken` from a successful response into the
   local `accessToken` variable, then run versioned requests. Do not use the ID token
   or a dashboard session cookie. Reauthenticate when the access token expires; remove
   or replace `accessToken` in the ignored local file when finished.

## Example flow

1. Create or list a service.
2. Create a monitor under `serviceId`.
3. Read, run, update, disable, or enable the monitor.
4. Inspect monitor and service incidents/audit.
5. Test channel and escalation policy operations separately.

Create local environment file from `environments/development.example.yml`; never commit deployed URLs or credentials.
