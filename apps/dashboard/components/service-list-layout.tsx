'use client'

import Link from 'next/link'
import { useDeferredValue, useState } from 'react'

import { ServiceOverviewCard } from '@/components/service-overview-card'
import { buttonVariants } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { filterServiceList, type ServiceFilter } from '@/lib/service-list-filter'
import type { Service } from '@/lib/types'
import { cn } from '@/lib/utils'

interface SummaryIndicatorProps {
  label: string
  value: number
  tone?: 'default' | 'danger'
}

function SummaryIndicator({ label, value, tone = 'default' }: SummaryIndicatorProps) {
  return (
    <div className="rounded-lg border border-border/80 bg-card px-3 py-3 shadow-sm md:min-w-32 md:px-4">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-muted-foreground">
        {label}
      </p>
      <p
        className={cn(
          'mt-1 font-mono text-2xl font-semibold text-foreground md:text-3xl',
          tone === 'danger' && 'text-status-down'
        )}
      >
        {value}
      </p>
    </div>
  )
}

export function ServiceListLayout({ services }: { services: Service[] }) {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<ServiceFilter>('all')
  const deferredSearch = useDeferredValue(search)
  const visibleServices = filterServiceList(services, deferredSearch, filter)
  const activeCount = services.filter((service) => service.lifecycleState === 'active').length
  const draftCount = services.filter((service) => service.lifecycleState === 'draft').length
  const downCount = services.filter(
    (service) => service.rollupStatus?.toUpperCase() === 'DOWN'
  ).length

  return (
    <div className="grid gap-6">
      <section className="grid gap-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
        <div className="hidden md:block">
          <h1 className="text-3xl font-semibold tracking-tight text-foreground">Services</h1>
          <p className="mt-2 max-w-2xl text-sm text-muted-foreground">
            Track service health, lifecycle, and monitor coverage from one place.
          </p>
        </div>
        <div aria-label="Service health summary" className="grid grid-cols-3 gap-2 md:gap-3">
          <SummaryIndicator label="Active" value={activeCount} />
          <SummaryIndicator label="Drafts" value={draftCount} />
          <SummaryIndicator label="Down now" tone="danger" value={downCount} />
        </div>
      </section>

      <section
        aria-label="Service list controls"
        className="grid gap-3 md:grid-cols-[1fr_auto_auto]"
      >
        <div>
          <label className="sr-only" htmlFor="service-search">
            Search services
          </label>
          <Input
            id="service-search"
            onChange={(event) => setSearch(event.target.value)}
            placeholder="Search services..."
            type="search"
            value={search}
          />
        </div>
        <div>
          <label className="sr-only" htmlFor="service-filter">
            Filter services
          </label>
          <Select
            id="service-filter"
            onChange={(event) => setFilter(event.target.value as ServiceFilter)}
            value={filter}
          >
            <option value="all">All services</option>
            <option value="active">Active</option>
            <option value="draft">Drafts</option>
            <option value="down">Down now</option>
          </Select>
        </div>
        <Link className={cn(buttonVariants(), 'hidden md:inline-flex')} href="/services/new">
          Create service
        </Link>
      </section>

      <section className="grid gap-4 pb-24 md:grid-cols-2 md:pb-0 xl:grid-cols-4">
        {visibleServices.length > 0 ? (
          visibleServices.map((service) => (
            <ServiceOverviewCard key={service.serviceId} service={service} />
          ))
        ) : (
          <Card className="md:col-span-2 xl:col-span-4">
            <CardContent className="py-8 text-center text-sm text-muted-foreground">
              No services match the current search or filter.
            </CardContent>
          </Card>
        )}
      </section>

      <Link
        aria-label="Create service"
        className={cn(
          buttonVariants({ size: 'lg' }),
          'fixed bottom-[calc(1rem+env(safe-area-inset-bottom))] right-4 z-40 h-14 w-14 rounded-full p-0 text-2xl shadow-lg md:hidden'
        )}
        href="/services/new"
      >
        +
      </Link>
    </div>
  )
}
