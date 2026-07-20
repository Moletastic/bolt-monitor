import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { ChannelTypeIcon } from '@/components/channel-type-icon'
import { ChannelUsageScope, buildChannelUsageMap } from '@/components/channel-usage-scope'
import { EmptyState } from '@/components/empty-state'
import { FocusOnMount } from '@/components/focus-on-mount'
import { Card, CardContent } from '@/components/ui/card'
import { Feedback } from '@/components/ui/feedback'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { listEscalationPolicies, listNotificationChannels } from '@/lib/api'
import type { EscalationPolicy } from '@/lib/types'
import { formatDateTime } from '@/lib/utils'

export default async function ChannelsPage({
  searchParams,
}: {
  searchParams: Promise<{ deleted?: string }>
}) {
  const query = await searchParams
  const [channels, policiesResult] = await Promise.all([
    listNotificationChannels().catch(() => []),
    listEscalationPolicies().catch(() => [] as EscalationPolicy[]),
  ])
  const usageByChannel = buildChannelUsageMap(policiesResult)
  return (
    <AppShell currentPath="/integrations/channels">
      <div className="grid gap-6">
        <div className="flex items-center justify-between gap-4">
          <div>
            <h1 className="dashboard-display">Notification channels</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Reusable destinations for alerts. Configure once, share across routes.
            </p>
          </div>
          <Link
            className="rounded-md border border-primary/40 bg-primary/10 px-3 py-2 text-sm font-semibold text-primary hover:bg-primary/20"
            data-create-cta
            href="/integrations/channels/new"
          >
            Create channel
          </Link>
        </div>
        {query.deleted ? (
          <FocusOnMount active>
            <Feedback tone="success">Notification channel permanently deleted.</Feedback>
          </FocusOnMount>
        ) : null}
        {channels.length === 0 ? (
          <EmptyState
            actionHref="/integrations/channels/new"
            actionLabel="Create your first channel"
            description="Create a Telegram bot, email sender, or webhook before assigning one to a route."
            title="No channels yet"
          />
        ) : (
          <Card>
            <CardContent className="overflow-x-auto p-0" data-list-root>
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Name</TableHead>
                    <TableHead>Type</TableHead>
                    <TableHead>Target</TableHead>
                    <TableHead>Scope</TableHead>
                    <TableHead>Updated</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {channels.map((channel) => (
                    <TableRow data-list-item key={channel.channelId}>
                      <TableCell className="font-semibold">
                        <Link
                          className="text-primary hover:underline"
                          href={`/integrations/channels/${channel.channelId}`}
                        >
                          {channel.name}
                        </Link>
                      </TableCell>
                      <TableCell className="capitalize">
                        <span className="inline-flex items-center gap-2">
                          <ChannelTypeIcon type={channel.type} />
                          {channel.type}
                        </span>
                      </TableCell>
                      <TableCell>{channel.target}</TableCell>
                      <TableCell>
                        <ChannelUsageScope
                          channelId={channel.channelId}
                          policies={usageByChannel.get(channel.channelId) ?? []}
                        />
                      </TableCell>
                      <TableCell className="font-mono text-xs">
                        {formatDateTime(channel.updatedAt)}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </CardContent>
          </Card>
        )}
      </div>
    </AppShell>
  )
}
