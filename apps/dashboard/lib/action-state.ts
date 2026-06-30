import { humanize, messageFor, type ApiError, type ApiErrorCode } from '@/lib/errors'

export type SerializedActionError = {
  code: ApiErrorCode
  details: Record<string, unknown>
  message?: string
}

export type ActionState<T = undefined> =
  | { status: 'idle' }
  | { status: 'success'; data: T; message?: string }
  | { status: 'error'; error: SerializedActionError }

export const idleActionState: ActionState = { status: 'idle' }

export function actionOk<T>(data: T, message?: string): ActionState<T> {
  return { status: 'success', data, message }
}

export function actionErr(error: ApiError): ActionState {
  const message = error.message && error.message !== error.code ? messageFor(error) : undefined
  return {
    status: 'error',
    error: {
      code: error.code,
      details: error.details,
      ...(message ? { message } : {}),
    },
  }
}

export function actionErrorMessage(error: SerializedActionError): string {
  return error.message ?? humanize(error.code)
}
