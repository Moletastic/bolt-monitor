import type { Service } from '@/lib/types'

export type ServiceFilter = 'all' | 'active' | 'draft' | 'down'

function matchesSearch(service: Service, query: string) {
  if (!query) {
    return true
  }

  const haystack = [
    service.name,
    service.description,
    service.lifecycleState,
    service.serviceCategory,
    service.rollupStatus,
  ]
    .filter((value): value is string => Boolean(value))
    .join(' ')
    .toLowerCase()

  return haystack.includes(query)
}

function matchesFilter(service: Service, filter: ServiceFilter) {
  if (filter === 'active') {
    return service.lifecycleState === 'active'
  }
  if (filter === 'draft') {
    return service.lifecycleState === 'draft'
  }
  if (filter === 'down') {
    return service.rollupStatus?.toUpperCase() === 'DOWN'
  }
  return true
}

export function filterServiceList(services: Service[], query: string, filter: ServiceFilter) {
  const normalizedQuery = query.trim().toLowerCase()
  return services.filter(
    (service) => matchesSearch(service, normalizedQuery) && matchesFilter(service, filter)
  )
}
