import Link from 'next/link'

import { AppShell } from '@/components/app-shell'
import { buttonVariants } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function NotFound() {
  return (
    <AppShell currentPath="/services">
      <Card>
        <CardHeader>
          <CardTitle>Monitor not found</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <p className="text-sm text-muted-foreground">
            Requested monitor does not exist in current tenant context.
          </p>
          <Link className={buttonVariants()} href="/services">
            Back to services overview
          </Link>
        </CardContent>
      </Card>
    </AppShell>
  )
}
