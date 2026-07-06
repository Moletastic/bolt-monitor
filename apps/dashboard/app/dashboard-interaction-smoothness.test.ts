import { readFile } from 'node:fs/promises'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

const root = process.cwd()

async function source(path: string) {
  return readFile(join(root, path), 'utf8')
}

describe('dashboard interaction smoothness guards', () => {
  it('schedules polling refreshes as transition work inside the polling provider', async () => {
    const pollingProvider = await source('components/polling-provider.tsx')

    expect(pollingProvider).toContain('import { startTransition')
    expect(pollingProvider).toContain('startTransition(() => {')
    expect(pollingProvider).toContain('router.refresh()')
    expect(pollingProvider).toContain('lastVisibilityStateRef')
    expect(pollingProvider).toContain(
      'if (lastVisibilityStateRef.current === nextVisibilityState) return'
    )
    expect(pollingProvider).toContain('if (intervalRef.current) return')
    expect(pollingProvider).toContain('clearInterval(intervalRef.current)')
  })

  it('keeps notification channel test sends on same-page action state', async () => {
    const channelPage = await source('app/(monitoring)/integrations/channels/[channelId]/page.tsx')
    const actions = await source('lib/actions.ts')
    const api = await source('lib/api.ts')
    const infra = await readFile(join(root, '../../infra/stacks/bootstrap.ts'), 'utf8')

    expect(channelPage).toContain('buttonLabel="Send test"')
    expect(channelPage).toContain('pendingLabel="Sending test..."')
    expect(channelPage).toContain('testNotificationChannelStateAction')
    expect(channelPage).toContain('<NotificationChannelForm channel={channel} />')
    expect(actions).toContain('testNotificationChannelStateAction')
    expect(actions).toContain("return actionOk(undefined, 'Test notification sent.')")
    expect(api).toContain('/api/v1/notification-channels/${channelId}/test')
    expect(infra).toContain('POST /api/v1/notification-channels/{channelId}/test')
  })

  it('explains Telegram chat ID setup before test sends fail', async () => {
    const channelForm = await source('components/notification-channel-form.tsx')

    expect(channelForm).toContain('Use the numeric chat ID')
    expect(channelForm).toContain('message the bot first')
    expect(channelForm).toContain('add the bot')
  })

  it('keeps imperative router navigation out of dashboard components', async () => {
    const files = [
      'components/same-page-action-form.tsx',
      'components/monitor-table.tsx',
      'app/(monitoring)/incidents/[id]/page.tsx',
      'app/(monitoring)/services/[serviceId]/monitors/[monitorId]/page.tsx',
    ]

    for (const file of files) {
      const text = await source(file)
      expect(text).not.toContain('useRouter')
      expect(text).not.toContain('router.push')
      expect(text).not.toContain('router.replace')
    }
  })

  it('lets inline-owned query feedback suppress duplicate toasts', async () => {
    const toastWatcher = await source('components/toast-watcher.tsx')
    const actions = await source('lib/actions.ts')
    const feedbackOwnership = await source('lib/feedback-ownership.ts')

    expect(toastWatcher).toContain("feedbackOwner === 'inline'")
    expect(toastWatcher).toContain('hasInlineOwnedQueryFeedback')
    expect(feedbackOwnership).toContain("deletedService: 'inline'")
    expect(feedbackOwnership).toContain("deletedMonitor: 'inline'")
    expect(feedbackOwnership).toContain("created: 'toast'")
    expect(actions).toContain('appendInlineFeedback')
  })

  it('keeps same-page mutation flows on typed action state with local feedback', async () => {
    const actionForm = await source('components/same-page-action-form.tsx')
    const actions = await source('lib/actions.ts')
    const queryFeedbackBanner = await source('components/query-feedback-banner.tsx')

    expect(actionForm).toContain('useActionState')
    expect(actionForm).toContain('const [state, formAction, pending]')
    expect(actionForm).toContain('setLocallyPending(true)')
    expect(actionForm).toContain('const isPending = pending || locallyPending')
    expect(actionForm).toContain('samePageActionStartEvent')
    expect(actionForm).toContain('role="status"')
    expect(actionForm).toContain('role="alert"')
    expect(queryFeedbackBanner).toContain('window.addEventListener(samePageActionStartEvent')
    expect(actions).toContain('toggleMonitorStateAction')
    expect(actions).toContain('acknowledgeIncidentStateAction')
    expect(actions).toContain('resolveIncidentStateAction')
    expect(actions).toContain('return actionOk(undefined, enabled ?')
    expect(actions).toContain('return actionErr(result.error)')
    expect(actions).toContain("revalidatePath('/incidents')")
    expect(actions).toContain("revalidatePath('/services')")
  })

  it('keeps mobile monitor card actions outside row-level links', async () => {
    const monitorTable = await source('components/monitor-table.tsx')

    expect(monitorTable).toContain('<SamePageActionForm')
    expect(monitorTable).not.toContain('</SamePageActionForm>\n            </Link>')
  })

  it('keeps destructive delete focus from falling back to the document body', async () => {
    const deleteForm = await source('components/delete-resource-form.tsx')
    const api = await source('lib/api.ts')
    const servicesPage = await source('app/(monitoring)/services/page.tsx')
    const policiesPage = await source('app/(monitoring)/policies/page.tsx')
    const policyDetailPage = await source('app/(monitoring)/policies/[policyId]/page.tsx')
    const channelsPage = await source('app/(monitoring)/integrations/channels/page.tsx')

    expect(deleteForm).toContain('action={action}')
    expect(api).toContain('apiRequestVoid')
    expect(api).toContain('deleteMonitor(serviceId: string, monitorId: string)')
    expect(api).toContain('deleteEscalationPolicy(policyId: string)')
    expect(servicesPage).toContain('function DeletedServiceFeedback')
    expect(servicesPage).toContain('query.deletedService ? <DeletedServiceFeedback /> : null')
    expect(policyDetailPage).toContain('deleteEscalationPolicyAction')
    expect(policyDetailPage).toContain('label="Delete route"')
    expect(policyDetailPage).toContain('name="returnTo"')
    expect(policyDetailPage).toContain('value={`/policies/${policy.policyId}`}')
    expect(deleteForm).not.toContain('javascript:throw')
    expect(deleteForm).not.toContain('document.body.focus')
    expect(servicesPage).toContain('<FocusOnMount active')
    expect(policiesPage).toContain('<FocusOnMount active')
    expect(channelsPage).toContain('<FocusOnMount active')
  })
})
