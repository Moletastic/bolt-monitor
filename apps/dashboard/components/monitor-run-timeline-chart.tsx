'use client'

import {
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'

import type { MonitorChartPoint, MonitorIndicator } from '@/lib/monitor-detail-metrics'
import { formatDateTime, formatDuration, formatOutcome } from '@/lib/utils'

interface ChartDatum extends MonitorChartPoint {
  index: number
}

interface DotProps {
  cx?: number
  cy?: number
  payload?: ChartDatum
}

function isDotProps(value: unknown): value is DotProps {
  if (!value || typeof value !== 'object') {
    return false
  }
  const candidate = value as DotProps
  return typeof candidate.cx === 'number' && typeof candidate.cy === 'number'
}

function RunDot(props: unknown) {
  if (!isDotProps(props)) {
    return <g />
  }

  const success = props.payload?.success ?? false
  return (
    <circle
      cx={props.cx}
      cy={props.cy}
      fill={success ? 'hsl(var(--primary))' : 'hsl(var(--status-down))'}
      r={success ? 2.75 : 3.5}
      stroke="hsl(var(--background))"
      strokeWidth={1.25}
    />
  )
}

function ActiveRunDot(props: unknown) {
  if (!isDotProps(props)) {
    return <g />
  }

  const success = props.payload?.success ?? false
  return (
    <circle
      cx={props.cx}
      cy={props.cy}
      fill={success ? 'hsl(var(--primary))' : 'hsl(var(--status-down))'}
      r={5}
      stroke="hsl(var(--background))"
      strokeWidth={2}
    />
  )
}

function TimelineTooltip({
  active,
  payload,
}: {
  active?: boolean
  payload?: Array<{ payload?: ChartDatum }>
}) {
  const point = payload?.[0]?.payload
  if (!active || !point) {
    return null
  }

  return (
    <div className="grid gap-1 rounded-md border border-border bg-surface-high px-3 py-2 text-xs text-foreground shadow-panel">
      <p className="font-mono text-[11px] text-muted-foreground">{formatTime(point.startedAt)}</p>
      <p className="font-semibold text-foreground">{formatDuration(point.durationMs)}</p>
      <p className="text-muted-foreground">
        {formatOutcome(point.outcome)}
        {point.statusCode !== undefined ? ` · HTTP ${point.statusCode}` : ''}
      </p>
      {point.error ? (
        <p className="break-words font-mono text-[11px] text-status-down">{point.error}</p>
      ) : null}
    </div>
  )
}

function formatTime(value: string) {
  const formatted = formatDateTime(value)
  const timePart = formatted.split(', ').at(-1) ?? formatted
  return timePart.replace(/\s?(AM|PM)$/i, '')
}

export function MonitorRunTimelineChart({
  points,
}: {
  points: MonitorChartPoint[]
  indicators: MonitorIndicator[]
}) {
  const chartData: ChartDatum[] = points.map((point, index) => ({ ...point, index }))
  const durations = chartData.map((point) => point.durationMs).filter(Number.isFinite)
  const minDuration = Math.min(...durations)
  const maxDuration = Math.max(...durations)
  const domainPadding = Math.max((maxDuration - minDuration) * 0.12, 20)
  const domainMin = Math.max(0, minDuration - domainPadding)
  const domainMax = maxDuration + domainPadding
  return (
    <div className="space-y-3">
      <div className="h-64 w-full">
        <ResponsiveContainer height="100%" width="100%">
          <LineChart data={chartData} margin={{ bottom: 8, left: 2, right: 12, top: 12 }}>
            <CartesianGrid stroke="hsl(var(--border))" strokeDasharray="3 8" strokeOpacity={0.4} />
            <XAxis
              axisLine={{ stroke: 'hsl(var(--border))', strokeOpacity: 0.5 }}
              dataKey="startedAt"
              minTickGap={48}
              tick={{ fill: 'hsl(var(--muted-foreground))', fontSize: 10 }}
              tickFormatter={formatTime}
              tickLine={false}
            />
            <YAxis
              axisLine={{ stroke: 'hsl(var(--border))', strokeOpacity: 0.5 }}
              domain={[domainMin, domainMax]}
              tick={{ fill: 'hsl(var(--muted-foreground))', fontSize: 10 }}
              tickFormatter={(value) => formatDuration(Number(value))}
              tickLine={false}
              width={54}
            />
            <Tooltip content={<TimelineTooltip />} cursor={{ stroke: 'hsl(var(--border))' }} />
            <Line
              activeDot={ActiveRunDot}
              dataKey="durationMs"
              dot={RunDot}
              isAnimationActive={false}
              stroke="hsl(var(--primary))"
              strokeLinecap="round"
              strokeWidth={1.5}
              type="monotone"
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
      <div className="flex justify-end border-t border-border/50 pt-3 text-[10px] text-muted-foreground">
        <span className="font-mono">Last {points.length} checks</span>
      </div>
    </div>
  )
}
