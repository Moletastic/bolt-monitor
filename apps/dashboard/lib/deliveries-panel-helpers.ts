import { DELIVERY_STATES, type DeliveryState } from '@/lib/types'

export const __DELIVERY_STATES_FOR_TEST: readonly DeliveryState[] = DELIVERY_STATES

export const __DELIVERY_STATE_LABELS: Record<DeliveryState, string> = {
  pending: 'Pending',
  in_flight: 'In flight',
  retryable_failed: 'Retryable failure',
  ambiguous: 'Ambiguous (provider may have received)',
  delivered: 'Accepted by provider',
  terminal_failed: 'Terminal failure',
}
