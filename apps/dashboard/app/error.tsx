'use client'

import { useState } from 'react'

import { AppShell } from '@/components/app-shell'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function DashboardError({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  const [detailsOpen, setDetailsOpen] = useState(false)

  return (
    <AppShell currentPath="/">
      <div className="grid gap-6">
        <Card className="border-status-warn/30 bg-status-warn/5">
          <CardHeader>
            <CardTitle>Dashboard temporarily unavailable</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4 text-sm text-status-warn">
            <p>
              Something broke while loading this view. The dashboard shell is still usable; retry to
              re-fetch the failed data.
            </p>
            <div className="flex gap-3">
              <button
                className="rounded-md border border-status-warn/40 bg-status-warn/10 px-3 py-2 text-sm font-semibold text-status-warn hover:bg-status-warn/20"
                onClick={() => reset()}
                type="button"
              >
                Retry
              </button>
              <button
                className="rounded-md border border-border bg-surface-low px-3 py-2 text-sm font-semibold text-foreground hover:bg-surface-high"
                onClick={() => setDetailsOpen((open) => !open)}
                type="button"
              >
                {detailsOpen ? 'Hide details' : 'Show details'}
              </button>
            </div>
            {detailsOpen ? (
              <pre className="overflow-x-auto rounded-md border border-status-warn/30 bg-background/50 p-3 font-mono text-xs text-foreground">
                {error.message}
                {error.digest ? `\ndigest: ${error.digest}` : ''}
              </pre>
            ) : null}
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
