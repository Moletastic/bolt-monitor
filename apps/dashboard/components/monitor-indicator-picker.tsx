'use client'

import { useState } from 'react'
import { Activity, AlertTriangle, Gauge, TrendingUp, type LucideIcon } from 'lucide-react'

import { cn } from '@/lib/utils'
import type { MonitorIndicator, MonitorIndicatorKey } from '@/lib/monitor-detail-metrics'

const indicatorIcons: Record<MonitorIndicatorKey, LucideIcon> = {
  state: Activity,
  uptime: TrendingUp,
  p99: Gauge,
  errors: AlertTriangle,
}

const indicatorShortLabels: Record<MonitorIndicatorKey, string> = {
  state: 'State',
  uptime: 'Uptime',
  p99: 'P99',
  errors: 'Errors',
}

const toneClasses: Record<MonitorIndicator['tone'], string> = {
  neutral: 'text-muted-foreground bg-surface-high',
  success: 'text-status-up bg-status-up/10',
  warning: 'text-status-warn bg-status-warn/10',
  danger: 'text-status-down bg-status-down/10',
}

export function MonitorIndicatorCard({ indicator }: { indicator: MonitorIndicator }) {
  const Icon = indicatorIcons[indicator.key]

  return (
    <div className="rounded-lg border border-border bg-surface-low p-4">
      <div
        className={cn(
          'flex h-9 w-9 items-center justify-center rounded-md',
          toneClasses[indicator.tone]
        )}
      >
        <Icon aria-hidden="true" className="h-4 w-4" />
      </div>
      <p className="mt-4 text-[11px] font-bold uppercase tracking-[0.24em] text-muted-foreground">
        {indicator.label}
      </p>
      <p className="mt-2 text-2xl font-semibold capitalize tracking-tight text-foreground">
        {indicator.value}
      </p>
      <p className="mt-1 text-sm text-muted-foreground">{indicator.detail}</p>
    </div>
  )
}

export function MobileMonitorIndicatorPicker({ indicators }: { indicators: MonitorIndicator[] }) {
  const [selectedKey, setSelectedKey] = useState<MonitorIndicatorKey>('state')
  const selected = indicators.find((indicator) => indicator.key === selectedKey) ?? indicators[0]

  return (
    <div className="space-y-3 md:hidden">
      <div
        className="grid grid-cols-4 gap-1 rounded-lg border border-border bg-surface-low p-1"
        role="tablist"
      >
        {indicators.map((indicator) => {
          const isSelected = indicator.key === selected.key
          const Icon = indicatorIcons[indicator.key]

          return (
            <button
              aria-selected={isSelected}
              className={cn(
                'inline-flex min-h-10 items-center justify-center gap-1 rounded-md px-2 text-xs font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
                isSelected
                  ? 'bg-surface-high text-foreground'
                  : 'text-muted-foreground hover:text-foreground'
              )}
              key={indicator.key}
              onClick={() => setSelectedKey(indicator.key)}
              role="tab"
              type="button"
            >
              <Icon aria-hidden="true" className="h-3.5 w-3.5" />
              <span>{indicatorShortLabels[indicator.key]}</span>
            </button>
          )
        })}
      </div>
      <MonitorIndicatorCard indicator={selected} />
    </div>
  )
}
