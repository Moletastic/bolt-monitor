import { actionErrorMessage, type ActionState } from '@/lib/action-state'

export function schedulerConfigFeedback(
  state: ActionState
):
  | { tone: 'success'; message: string; code?: undefined }
  | { tone: 'error'; message: string; code: string }
  | null {
  if (state.status === 'success') {
    return { tone: 'success', message: state.message ?? 'Scheduler configuration updated.' }
  }
  if (state.status === 'error') {
    return {
      tone: 'error',
      message: actionErrorMessage(state.error),
      code: state.error.code,
    }
  }
  return null
}
