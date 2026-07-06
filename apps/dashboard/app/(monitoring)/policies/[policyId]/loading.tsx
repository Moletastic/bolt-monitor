import { AppShell } from '@/components/app-shell'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function Loading() {
  return (
    <AppShell currentPath="/policies">
      <div className="grid gap-6">
        <div className="grid gap-2">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-[30rem] max-w-full" />
        </div>
        <Card>
          <CardHeader>
            <CardTitle>
              <Skeleton className="h-5 w-44" />
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <Skeleton className="h-10" />
            <Skeleton className="h-24" />
            <div className="grid gap-4 md:grid-cols-2">
              <Skeleton className="h-40" />
              <Skeleton className="h-40" />
            </div>
            <Skeleton className="h-10 w-32" />
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
