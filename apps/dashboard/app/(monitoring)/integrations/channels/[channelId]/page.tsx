import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { ChannelUsageScope, buildChannelUsageMap } from '@/components/channel-usage-scope'
import { DeleteResourceForm } from '@/components/delete-resource-form'
import { NotificationChannelForm } from '@/components/notification-channel-form'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { deleteNotificationChannelAction } from '@/lib/actions'
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
    <AppShell currentPath="/integrations/channels">
      <div className="grid gap-6">
        <div>
          <Link className="text-sm text-primary hover:underline" href="/integrations/channels">
            Back to channels
          </Link>
          <h1 className="mt-2 text-2xl font-semibold tracking-tight">Edit notification channel</h1>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Channel details</CardTitle>
          </CardHeader>
          <CardContent>
            <NotificationChannelForm channel={channel} />
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
