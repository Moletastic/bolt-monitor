import { describe, expect, it } from 'vitest'

import { ApiErrorCode } from '@/lib/errors'
import { schedulerConfigFeedback } from '@/lib/scheduler-config-feedback'

describe('schedulerConfigFeedback', () => {
  it('returns success feedback for successful action state', () => {
    expect(
      schedulerConfigFeedback({
        status: 'success',
        data: undefined,
        message: 'Scheduler configuration updated.',
      })
    ).toEqual({ tone: 'success', message: 'Scheduler configuration updated.' })
  })

  it('returns typed error feedback without raw enum copy as primary message', () => {
    expect(
      schedulerConfigFeedback({
        status: 'error',
        error: {
          code: ApiErrorCode.Internal,
          details: { requestId: 'req-1' },
        },
      })
    ).toEqual({
      tone: 'error',
      message: 'The server hit an unexpected error.',
      code: ApiErrorCode.Internal,
    })
  })
})
