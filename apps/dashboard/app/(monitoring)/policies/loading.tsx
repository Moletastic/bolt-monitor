import { AppShell } from '@/components/app-shell'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export default function Loading() {
  return (
    <AppShell currentPath="/policies">
      <div className="grid gap-6">
        <div className="flex items-center justify-between">
          <div className="grid gap-2">
            <Skeleton className="h-7 w-64" />
            <Skeleton className="h-4 w-[28rem] max-w-full" />
          </div>
          <Skeleton className="h-9 w-32" />
        </div>
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
          {Array.from({ length: 6 }).map((_, index) => (
            <Card key={index}>
              <CardHeader>
                <CardTitle>
                  <Skeleton className="h-5 w-40" />
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3 text-sm">
                <Skeleton className="h-4 w-full" />
                <dl className="grid grid-cols-2 gap-3">
                  {Array.from({ length: 3 }).map((_, item) => (
                    <div className={item === 2 ? 'col-span-2 grid gap-1' : 'grid gap-1'} key={item}>
                      <Skeleton className="h-3 w-20" />
                      <Skeleton className="h-4 w-24" />
                    </div>
                  ))}
                </dl>
              </CardContent>
            </Card>
          ))}
        </div>
      </div>
    </AppShell>
  )
}
