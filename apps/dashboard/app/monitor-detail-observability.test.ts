import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const pageSource = readFileSync(
  join(process.cwd(), 'app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx'),
  'utf8'
)
const pickerSource = readFileSync(
  join(process.cwd(), 'components/monitor-indicator-picker.tsx'),
  'utf8'
)
const chartSource = readFileSync(
  join(process.cwd(), 'components/monitor-run-timeline-chart.tsx'),
  'utf8'
)
const actionsMenuSource = readFileSync(
  join(process.cwd(), 'components/monitor-detail-actions-menu.tsx'),
  'utf8'
)

describe('monitor detail observability layout guards', () => {
  it('keeps compact mobile action controls accessible', () => {
    expect(pageSource).toContain('Run now')
    expect(pageSource).toContain('aria-label="Edit monitor"')
    expect(pageSource).toContain('href={`${returnTo}/edit`}')
    expect(pageSource).toContain('<MonitorDetailActionsMenu')
    expect(actionsMenuSource).toContain('aria-label="More monitor actions"')
    expect(actionsMenuSource).toContain('Enter maintenance')
    expect(actionsMenuSource).toContain('Delete monitor')
  })

  it('moves status and request metadata into the monitor header', () => {
    expect(pageSource).toContain('getStatusDotClass')
    expect(pageSource).toContain('role="img"')
    expect(pageSource).toContain('<MonitorProtocolBadge type={monitor.type} />')
    expect(pageSource).toContain('{formatMonitorCadence(monitor.intervalSeconds ?? 0)}')
    expect(pageSource).toContain('{formatDuration(monitor.http?.timeoutMs ?? 0)} timeout')
    expect(pageSource).not.toContain('Check configuration')
  })

  it('renders the chart as a library-backed run timeline', () => {
    expect(pageSource).toContain('<CardTitle>Run timeline</CardTitle>')
    expect(pageSource).toContain(
      '<MonitorRunTimelineChart indicators={indicators} points={chartPoints} />'
    )
    expect(chartSource).toContain('ResponsiveContainer')
    expect(chartSource).toContain('CartesianGrid')
    expect(chartSource).toContain('TimelineTooltip')
    expect(chartSource).toContain('tickFormatter={formatTime}')
    expect(chartSource).toContain('Last {points.length} checks')
    expect(chartSource).not.toContain('uptime')
    expect(chartSource).not.toContain('P99')
  })

  it('moves destructive delete into a typed confirmation menu flow', () => {
    expect(pageSource).not.toContain('<Card className="border-status-down/30">')
    expect(actionsMenuSource).toContain(
      'Type <span className="font-mono text-foreground">{monitorName}</span> to confirm'
    )
    expect(actionsMenuSource).toContain('typed.trim() === expected')
    expect(actionsMenuSource).toContain('deleteMonitorAction')
  })

  it('adds icon-backed evidence tabs', () => {
    expect(pageSource).toContain("iconName: 'history'")
    expect(pageSource).toContain("iconName: 'incidents'")
    expect(pageSource).toContain("iconName: 'audit'")
    expect(pageSource).toContain('Runs')
    expect(pageSource).toContain('Incidents')
    expect(pageSource).toContain('Audit')
  })

  it('defaults the mobile indicator picker to current state', () => {
    expect(pickerSource).toContain("useState<MonitorIndicatorKey>('state')")
    expect(pickerSource).toContain('role="tablist"')
    expect(pickerSource).toContain('role="tab"')
  })
})
