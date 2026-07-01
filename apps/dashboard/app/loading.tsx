import { AppShell } from '@/components/app-shell'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { TableRowSkeleton } from '@/components/ui/skeleton-row'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export default function Loading() {
  return (
    <AppShell currentPath="/">
      <div className="grid gap-6">
        <section className="grid gap-3">
          <Skeleton className="h-3 w-40" />
          <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
            <div className="grid gap-2">
              <Skeleton className="h-8 w-72" />
              <Skeleton className="h-4 w-[28rem] max-w-full" />
            </div>
            <Skeleton className="h-9 w-32" />
          </div>
        </section>
        <section className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {Array.from({ length: 4 }).map((_, index) => (
            <Card key={index}>
              <CardHeader>
                <CardTitle>
                  <Skeleton className="h-4 w-32" />
                </CardTitle>
              </CardHeader>
              <CardContent className="space-y-3">
                <Skeleton className="h-9 w-20" />
                <Skeleton className="h-4 w-full" />
              </CardContent>
            </Card>
          ))}
        </section>
        <Card>
          <CardHeader>
            <CardTitle>
              <Skeleton className="h-4 w-40" />
            </CardTitle>
          </CardHeader>
          <CardContent className="grid gap-3">
            {Array.from({ length: 3 }).map((_, index) => (
              <div key={index} className="rounded-lg border border-border bg-surface-low p-4">
                <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                  <div className="grid gap-2">
                    <Skeleton className="h-4 w-64" />
                    <Skeleton className="h-3 w-48" />
                  </div>
                  <Skeleton className="h-5 w-24" />
                </div>
              </div>
            ))}
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>
              <Skeleton className="h-4 w-56" />
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="overflow-x-auto">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead>Service</TableHead>
                    <TableHead>Status</TableHead>
                    <TableHead>Lifecycle</TableHead>
                    <TableHead>Coverage</TableHead>
                    <TableHead>Updated</TableHead>
                  </TableRow>
                </TableHeader>
                <TableBody>
                  <TableRowSkeleton columns={5} rows={5} />
                </TableBody>
              </Table>
            </div>
          </CardContent>
        </Card>
        <section className="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
          <Card>
            <CardHeader>
              <CardTitle>
                <Skeleton className="h-4 w-40" />
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3">
              {Array.from({ length: 3 }).map((_, index) => (
                <div key={index} className="rounded-lg border border-border bg-surface-low p-3">
                  <div className="flex items-start justify-between gap-3">
                    <div className="grid gap-2">
                      <Skeleton className="h-4 w-64" />
                      <Skeleton className="h-3 w-32" />
                    </div>
                    <Skeleton className="h-5 w-20" />
                  </div>
                </div>
              ))}
            </CardContent>
          </Card>
          <Card>
            <CardHeader>
              <CardTitle>
                <Skeleton className="h-4 w-32" />
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-3">
              {Array.from({ length: 4 }).map((_, index) => (
                <div
                  key={index}
                  className="flex items-center justify-between gap-3 rounded-lg border border-border bg-surface-low px-3 py-2"
                >
                  <Skeleton className="h-4 w-40" />
                  <Skeleton className="h-4 w-12" />
                </div>
              ))}
            </CardContent>
          </Card>
        </section>
      </div>
    </AppShell>
  )
}
