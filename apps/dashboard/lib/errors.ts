/**
 * Structured error model for the dashboard.
 *
 * `ApiErrorCode` mirrors the constants declared in
 * `shared/errors/code.go`. Adding a code here without also adding it to the
 * Go registry (or vice versa) fails the drift test in `errors.test.ts`.
 *
 * `ApiError` is the runtime carrier. `fromEnvelope` lifts a wire-format
 * reason into the typed error, throwing on unknown codes so drift is
 * surfaced at the seam rather than at the UI.
 */

export const ApiErrorCode = {
  NotFound: 'NOT_FOUND',
  InvalidJson: 'INVALID_JSON',
  ValidationFailed: 'VALIDATION_FAILED',
  ImmutableField: 'IMMUTABLE_FIELD',
  InlineChannelConfig: 'INLINE_CHANNEL_CONFIG',
  ServiceNotFound: 'SERVICE_NOT_FOUND',
  ServiceAlreadyExists: 'SERVICE_ALREADY_EXISTS',
  ServiceActive: 'SERVICE_ACTIVE',
  ServiceNotArchived: 'SERVICE_NOT_ARCHIVED',
  ServiceHasNoPolicy: 'SERVICE_HAS_NO_POLICY',
  MonitorNotFound: 'MONITOR_NOT_FOUND',
  MonitorAlreadyExists: 'MONITOR_ALREADY_EXISTS',
  MonitorDisabled: 'MONITOR_DISABLED',
  MonitorStatusNotFound: 'MONITOR_STATUS_NOT_FOUND',
  LastMonitor: 'LAST_MONITOR',
  IncidentNotFound: 'INCIDENT_NOT_FOUND',
  IncidentNotActionable: 'INCIDENT_NOT_ACTIONABLE',
  PolicyNotFound: 'POLICY_NOT_FOUND',
  PolicyReferenced: 'POLICY_REFERENCED',
  ChannelNotFound: 'CHANNEL_NOT_FOUND',
  Internal: 'INTERNAL',
} as const

export type ApiErrorCode = (typeof ApiErrorCode)[keyof typeof ApiErrorCode]

const API_ERROR_CODE_VALUES: ReadonlySet<string> = new Set(Object.values(ApiErrorCode))

export function isApiErrorCode(value: string): value is ApiErrorCode {
  return API_ERROR_CODE_VALUES.has(value)
}

/**
 * Wire reason shape as produced by the Go response envelope.
 *
 * Kept structurally compatible with `ApiReason` in `lib/api-response.ts`; the
 * two types are kept separate because `ApiReason` is part of the success-path
 * response model and this file owns the error path.
 */
export interface ApiReasonPayload {
  code: string
  details: Record<string, unknown>
}

export class ApiError extends Error {
  readonly code: ApiErrorCode
  readonly details: Record<string, unknown>
  readonly status: number
  readonly httpStatus: number

  constructor(
    code: ApiErrorCode,
    status: number,
    details: Record<string, unknown> = {},
    message?: string
  ) {
    super(message ?? code)
    this.name = 'ApiError'
    this.code = code
    this.status = status
    this.httpStatus = status
    this.details = details
  }
}

/**
 * Construct an `ApiError` from a wire-format reason.
 *
 * Throws when the wire code does not map to a known `ApiErrorCode`. This is
 * intentional: an unknown code means the Go and TS registries have drifted,
 * and the dashboard should fail loud at the seam rather than silently
 * downgrade to a generic error.
 *
 * @param reason The wire reason from a `response.Err` envelope.
 * @param fallbackStatus HTTP status to attach when the envelope does not
 *                       carry one; defaults to 500.
 */
export function fromEnvelope(reason: ApiReasonPayload, fallbackStatus = 500): ApiError {
  if (!isApiErrorCode(reason.code)) {
    throw new Error(
      `fromEnvelope: unknown API error code "${reason.code}". ` +
        `Add it to ApiErrorCode (apps/dashboard/lib/errors.ts) and the Go ` +
        `registry (shared/errors/code.go).`
    )
  }
  return new ApiError(reason.code, fallbackStatus, reason.details)
}

/**
 * Humanize a code for UI surfacing when no `message` is present.
 *
 * `ApiError.message` always wins when present. This map is a last-resort
 * fallback for the rare case where the server returns a code without a
 * message (e.g. `NOT_FOUND` with no body).
 */
const HUMANIZED: Record<ApiErrorCode, string> = {
  NOT_FOUND: 'The requested resource was not found.',
  INVALID_JSON: 'The server received a malformed request.',
  VALIDATION_FAILED: 'The request was invalid.',
  IMMUTABLE_FIELD: 'A field in the request cannot be changed.',
  INLINE_CHANNEL_CONFIG: 'Escalation steps must reference saved channels by id.',
  SERVICE_NOT_FOUND: 'That service does not exist.',
  SERVICE_ALREADY_EXISTS: 'A service with that identifier already exists.',
  SERVICE_ACTIVE: 'That service is still active and cannot be modified.',
  SERVICE_NOT_ARCHIVED: 'The service must be archived before this action.',
  SERVICE_HAS_NO_POLICY: 'The service has no escalation policy attached.',
  MONITOR_NOT_FOUND: 'That monitor does not exist.',
  MONITOR_ALREADY_EXISTS: 'A monitor with that identifier already exists.',
  MONITOR_DISABLED: 'That monitor is currently disabled.',
  MONITOR_STATUS_NOT_FOUND: 'No status has been recorded for that monitor yet.',
  LAST_MONITOR: 'A service must keep at least one monitor.',
  INCIDENT_NOT_FOUND: 'That incident does not exist.',
  INCIDENT_NOT_ACTIONABLE: 'That incident is no longer actionable.',
  POLICY_NOT_FOUND: 'That escalation policy does not exist.',
  POLICY_REFERENCED: 'The escalation policy is still referenced by a service.',
  CHANNEL_NOT_FOUND: 'That notification channel does not exist.',
  INTERNAL: 'The server hit an unexpected error.',
}

export function humanize(code: ApiErrorCode): string {
  return HUMANIZED[code]
}

/**
 * Build a user-facing message for an `ApiError`.
 *
 * `ApiError.message` wins when present (it carries the server-supplied
 * detail). Otherwise the code is humanized via the static map. Used by
 * server actions that need to surface an error through a redirect query
 * string.
 */
export function messageFor(err: ApiError): string {
  return err.message && err.message !== err.code ? err.message : humanize(err.code)
}
