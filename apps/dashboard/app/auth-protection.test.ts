import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const source = (path: string) => readFileSync(join(process.cwd(), path), 'utf8')
const actionSource = source('lib/actions.ts')

const protectedActions = [
  'loadMonitorHistoryPageAction',
  'createMonitorAction',
  'updateMonitorAction',
  'toggleMonitorAction',
  'toggleMonitorStateAction',
  'toggleMaintenanceModeAction',
  'deleteMonitorAction',
  'acknowledgeIncidentAction',
  'acknowledgeIncidentStateAction',
  'resolveIncidentAction',
  'resolveIncidentStateAction',
  'updateSchedulerConfigAction',
  'updateSchedulerConfigStateAction',
  'triggerManualRunAction',
  'createServiceAction',
  'updateServiceAction',
  'archiveServiceAction',
  'deleteServiceAction',
  'createEscalationPolicyAction',
  'updateEscalationPolicyAction',
  'deleteEscalationPolicyAction',
  'createNotificationChannelAction',
  'updateNotificationChannelAction',
  'deleteNotificationChannelAction',
  'testNotificationChannelStateAction',
]

const stateChangingActions = protectedActions.filter(
  (action) => action !== 'loadMonitorHistoryPageAction'
)

describe('dashboard server protection', () => {
  it('validates the authoritative session before monitoring content renders', () => {
    const layout = source('app/(monitoring)/layout.tsx')
    expect(layout).toContain('await requireDashboardSession()')

    for (const route of [
      'app/(monitoring)/page.tsx',
      'app/(monitoring)/config/page.tsx',
      'app/(monitoring)/audit-trail/page.tsx',
      'app/(monitoring)/admin/scheduler/page.tsx',
      'app/(monitoring)/services/new/page.tsx',
      'app/(monitoring)/services/[serviceId]/edit/page.tsx',
      'app/(monitoring)/services/[serviceId]/monitors/new/page.tsx',
      'app/(monitoring)/services/[serviceId]/monitors/[monitorId]/edit/page.tsx',
    ]) {
      expect(source(route)).not.toContain('requireDashboardSession')
    }
  })

  it('guards every operator server action before it can invoke an API operation', () => {
    for (const action of protectedActions) {
      const start = actionSource.indexOf(`export async function ${action}`)
      const next = actionSource.indexOf('export async function ', start + 1)
      const body = actionSource.slice(start, next === -1 ? undefined : next)

      expect(start).toBeGreaterThan(-1)
      expect(body).toContain('await requireDashboardSession()')
    }
  })

  it('validates CSRF evidence before every state-changing dashboard action', () => {
    for (const action of stateChangingActions) {
      const start = actionSource.indexOf(`export async function ${action}`)
      const next = actionSource.indexOf('export async function ', start + 1)
      const body = actionSource.slice(start, next === -1 ? undefined : next)

      expect(body).toContain('await requireDashboardCsrf()')
      expect(body.indexOf('await requireDashboardCsrf()')).toBeLessThan(
        body.indexOf('await requireDashboardSession()')
      )
    }
  })

  it('redirects established sessions away from auth pages and auth route handlers', () => {
    expect(source('app/(auth)/layout.tsx')).toContain('await redirectIfDashboardSession()')
    const enrollmentRoute = source('app/(auth)/totp/enroll/route.ts')
    expect(enrollmentRoute).toContain('await redirectIfDashboardSession()')

    for (const action of [
      'app/(auth)/login/actions.ts',
      'app/(auth)/activate/actions.ts',
      'app/(auth)/forgot-password/actions.ts',
      'app/(auth)/reset-password/actions.ts',
      'app/(auth)/totp/challenge/actions.ts',
    ]) {
      expect(source(action)).toContain('await redirectIfDashboardSession()')
      expect(source(action)).toContain('await requireDashboardCsrf()')
    }
    expect(enrollmentRoute).toContain('await requireDashboardCsrf()')
  })
})
