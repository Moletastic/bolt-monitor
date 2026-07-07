import { AppShell } from '@/components/app-shell'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent } from '@/components/ui/card'

export default function Loading() {
  return (
    <AppShell currentPath="/services">
      <div className="grid gap-6">
        <h1 className="sr-only">Services</h1>
        <section className="grid gap-4 md:grid-cols-[minmax(0,1fr)_auto] md:items-end">
          <div className="hidden space-y-2 md:block">
            <Skeleton className="h-9 w-36" />
            <Skeleton className="h-4 w-[28rem] max-w-full" />
          </div>
          <div className="grid grid-cols-3 gap-2 md:gap-3">
            {Array.from({ length: 3 }).map((_, index) => (
              <Card key={index}>
                <CardContent className="space-y-2 p-3 md:p-4">
                  <Skeleton className="h-3 w-16" />
                  <Skeleton className="h-8 w-10" />
                </CardContent>
              </Card>
            ))}
          </div>
        </section>
        <section className="grid gap-3 md:grid-cols-[1fr_auto_auto]">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full md:w-36" />
          <Skeleton className="hidden h-10 w-32 md:block" />
        </section>
        <section className="grid gap-4 pb-24 md:grid-cols-2 md:pb-0 xl:grid-cols-4">
          {Array.from({ length: 6 }).map((_, index) => (
            <Card key={index}>
              <CardContent className="space-y-4 pt-6">
                <div className="flex items-start justify-between gap-4">
                  <div className="flex items-start gap-3">
                    <Skeleton className="h-11 w-11 rounded-xl" />
                    <div className="grid gap-1">
                      <Skeleton className="h-5 w-32" />
                      <Skeleton className="h-3 w-16" />
                    </div>
                  </div>
                  <Skeleton className="h-5 w-16" />
                </div>
                <Skeleton className="h-4 w-full" />
                <dl className="grid grid-cols-2 gap-3">
                  {Array.from({ length: 4 }).map((_, item) => (
                    <div className="grid gap-1" key={item}>
                      <Skeleton className="h-3 w-20" />
                      <Skeleton className="h-4 w-24" />
                    </div>
                  ))}
                </dl>
              </CardContent>
            </Card>
          ))}
        </section>
      </div>
    </AppShell>
  )
}
