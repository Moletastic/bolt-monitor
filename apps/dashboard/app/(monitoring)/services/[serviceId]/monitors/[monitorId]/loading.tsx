import { AppShell } from '@/components/app-shell'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function Loading() {
  return (
    <AppShell currentPath="/services">
      <div className="grid gap-6">
        <Card>
          <CardContent className="space-y-5 pt-6">
            <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div className="flex items-start gap-4">
                <Skeleton className="mt-3 h-3 w-3 rounded-full" />
                <div className="grid gap-2">
                  <div className="flex items-center gap-3">
                    <Skeleton className="h-8 w-64" />
                  </div>
                  <Skeleton className="h-5 w-80" />
                  <div className="flex gap-2">
                    <Skeleton className="h-7 w-32" />
                    <Skeleton className="h-7 w-24" />
                  </div>
                </div>
              </div>
              <div className="flex gap-2">
                <Skeleton className="h-10 w-28" />
                <Skeleton className="h-10 w-10 lg:w-20" />
                <Skeleton className="h-10 w-10" />
              </div>
            </div>
          </CardContent>
        </Card>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <Skeleton className="h-36" key={index} />
          ))}
        </div>
        <section>
          <Card>
            <CardHeader>
              <CardTitle>
                <Skeleton className="h-5 w-40" />
              </CardTitle>
            </CardHeader>
            <CardContent>
              <Skeleton className="h-64" />
            </CardContent>
          </Card>
        </section>
        <Card>
          <CardHeader>
            <CardTitle>
              <Skeleton className="h-5 w-32" />
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-3">
            {Array.from({ length: 5 }).map((_, index) => (
              <Skeleton className="h-11" key={index} />
            ))}
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
