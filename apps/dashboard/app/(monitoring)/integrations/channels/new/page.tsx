import { AppShell } from '@/components/app-shell'
import { NotificationChannelForm } from '@/components/notification-channel-form'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function NewChannelPage() {
  return (
    <AppShell
      breadcrumbs={[
        { label: 'Channels', href: '/integrations/channels' },
        { label: 'Create channel' },
      ]}
      currentPath="/integrations/channels/new"
    >
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">New notification channel</h1>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Channel details</CardTitle>
          </CardHeader>
          <CardContent>
            <NotificationChannelForm />
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}