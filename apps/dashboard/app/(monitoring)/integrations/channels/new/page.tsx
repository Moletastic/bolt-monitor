import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { NotificationChannelForm } from '@/components/notification-channel-form'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function NewChannelPage() {
  return (
    <AppShell currentPath="/integrations/channels">
      <div className="grid gap-6">
        <div>
          <Link className="text-sm text-primary hover:underline" href="/integrations/channels">
            Back to channels
          </Link>
          <h1 className="mt-2 text-2xl font-semibold tracking-tight">New notification channel</h1>
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
