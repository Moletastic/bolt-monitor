import { describe, expect, it } from 'vitest'

import { actionErr, actionErrorMessage } from './action-state'
import { ApiError, ApiErrorCode } from './errors'

describe('action-state', () => {
  it('serializes ApiError without passing class instances to clients', () => {
    const state = actionErr(
      new ApiError(
        ApiErrorCode.ValidationFailed,
        400,
        { field: 'recurringEnabled' },
        'Invalid toggle'
      )
    )

    expect(state).toEqual({
      status: 'error',
      error: {
        code: ApiErrorCode.ValidationFailed,
        details: { field: 'recurringEnabled' },
        message: 'Invalid toggle',
      },
    })
    if (state.status !== 'error') {
      throw new Error('Expected error action state')
    }

    expect(state.error).not.toBeInstanceOf(ApiError)
  })

  it('falls back to humanized copy when error message is absent', () => {
    const state = actionErr(new ApiError(ApiErrorCode.Internal, 500, { requestId: 'req-1' }))

    expect(state.status).toBe('error')
    if (state.status !== 'error') {
      throw new Error('Expected error action state')
    }

    expect(state.error.message).toBeUndefined()
    expect(actionErrorMessage(state.error)).toBe('The server hit an unexpected error.')
  })
})
