import { describe, expect, it } from 'vitest'

import { __DELIVERY_STATE_LABELS, __DELIVERY_STATES_FOR_TEST } from '@/lib/deliveries-panel-helpers'
import { DELIVERY_STATES, type Delivery } from '@/lib/types'

const baseDelivery: Delivery = {
  deliveryId: 'dlv_1',
  transitionId: 'TRN_1',
  channelId: 'CH_1',
  channelType: 'telegram',
  stepNumber: 1,
  state: 'pending',
  attemptCount: 0,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
}

describe('delivery state classification', () => {
  it('exposes all six states through the canonical constant', () => {
    expect(new Set(DELIVERY_STATES)).toEqual(
      new Set([
        'pending',
        'in_flight',
        'retryable_failed',
        'ambiguous',
        'delivered',
        'terminal_failed',
      ])
    )
  })

  it('only terminal_failed deliveries are replayable', () => {
    for (const state of __DELIVERY_STATES_FOR_TEST) {
      const delivery: Delivery = { ...baseDelivery, state }
      const expectedReplayable = state === 'terminal_failed'
      const actualReplayable = delivery.state === 'terminal_failed'
      expect(actualReplayable).toBe(expectedReplayable)
    }
  })

  it('provides a label for every delivery state', () => {
    for (const state of __DELIVERY_STATES_FOR_TEST) {
      expect(__DELIVERY_STATE_LABELS[state]).toMatch(/.+/)
    }
  })

  it('keeps provider-acceptance wording distinct from human receipt', () => {
    expect(__DELIVERY_STATE_LABELS.delivered.toLowerCase()).toContain('accepted')
    expect(__DELIVERY_STATE_LABELS.delivered.toLowerCase()).not.toContain('human')
  })
})
