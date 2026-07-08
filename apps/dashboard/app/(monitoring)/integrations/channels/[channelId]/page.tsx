import { AppShell } from '@/components/app-shell'
import { ChannelUsageScope, buildChannelUsageMap } from '@/components/channel-usage-scope'
import { DeleteResourceForm } from '@/components/delete-resource-form'
import { NotificationChannelForm } from '@/components/notification-channel-form'
import { SamePageActionForm } from '@/components/same-page-action-form'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { deleteNotificationChannelAction, testNotificationChannelStateAction } from '@/lib/actions'
import { getNotificationChannel, listEscalationPolicies } from '@/lib/api'

export default async function EditChannelPage({
  params,
}: {
  params: Promise<{ channelId: string }>
}) {
  const { channelId } = await params
  const [channel, policies] = await Promise.all([
    getNotificationChannel(channelId),
    listEscalationPolicies().catch(() => []),
  ])
  const usage = buildChannelUsageMap(policies).get(channel.channelId) ?? []
  const deleteBlocked = usage.length > 0
  return (
    <AppShell
      breadcrumbs={[
        { label: 'Channels', href: '/integrations/channels' },
        { label: channel.name || 'Channel' },
      ]}
      currentPath="/integrations/channels"
    >
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Edit notification channel</h1>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Channel details</CardTitle>
          </CardHeader>
          <CardContent className="space-y-5">
            <NotificationChannelForm channel={channel} />
            <div className="rounded-lg border border-border bg-surface-low p-4">
              <h2 className="text-sm font-semibold text-foreground">Test delivery</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Send a real test notification to this registered destination. No incident will be
                created.
              </p>
              <SamePageActionForm
                action={testNotificationChannelStateAction}
                buttonLabel="Send test"
                className="mt-4"
                pendingLabel="Sending test..."
                variant="secondary"
              >
                <input name="channelId" type="hidden" value={channel.channelId} />
              </SamePageActionForm>
            </div>
          </CardContent>
        </Card>
        <Card className="border-status-down/30">
          <CardHeader>
            <CardTitle>Delete channel</CardTitle>
          </CardHeader>
          <CardContent>
            {deleteBlocked ? (
              <div
                className="mb-4 rounded-md border border-status-warn/30 bg-status-warn/10 px-3 py-2 text-sm text-status-warn"
                role="status"
              >
                This channel is used by notification routes and cannot be deleted until those
                references are removed. Current usage:{' '}
                <ChannelUsageScope channelId={channel.channelId} policies={usage} />
              </div>
            ) : null}
            <DeleteResourceForm
              action={deleteNotificationChannelAction}
              confirmMessage={`Delete ${channel.name}? Routes using this channel will stop firing. This cannot be undone.`}
              disabled={deleteBlocked}
              label="Delete channel"
            >
              <input name="channelId" type="hidden" value={channel.channelId} />
              <input name="returnTo" type="hidden" value="/integrations/channels" />
            </DeleteResourceForm>
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
