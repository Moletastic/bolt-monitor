import { describe, expect, it } from 'vitest'

import { filterServiceList } from '@/lib/service-list-filter'
import type { Service } from '@/lib/types'

function service(overrides: Partial<Service> & Pick<Service, 'serviceId' | 'name'>): Service {
  return {
    lifecycleState: 'active',
    tenantId: 'DEFAULT',
    ...overrides,
  }
}

describe('filterServiceList', () => {
  const services = [
    service({
      description: 'Customer API edge service',
      lifecycleState: 'active',
      name: 'Public API',
      rollupStatus: 'UP',
      serviceCategory: 'http',
      serviceId: 'svc-1',
    }),
    service({
      description: 'Queue worker draft',
      lifecycleState: 'draft',
      name: 'Billing Worker',
      serviceCategory: 'queue',
      serviceId: 'svc-2',
    }),
    service({
      lifecycleState: 'active',
      name: 'Database',
      rollupStatus: 'DOWN',
      serviceCategory: 'database',
      serviceId: 'svc-3',
    }),
  ]

  it('searches visible service attributes case-insensitively', () => {
    expect(filterServiceList(services, 'api', 'all').map((item) => item.serviceId)).toEqual([
      'svc-1',
    ])
    expect(filterServiceList(services, 'QUEUE', 'all').map((item) => item.serviceId)).toEqual([
      'svc-2',
    ])
  })

  it('filters services by lifecycle and down status', () => {
    expect(filterServiceList(services, '', 'draft').map((item) => item.serviceId)).toEqual([
      'svc-2',
    ])
    expect(filterServiceList(services, '', 'down').map((item) => item.serviceId)).toEqual([
      'svc-3',
    ])
  })
})
