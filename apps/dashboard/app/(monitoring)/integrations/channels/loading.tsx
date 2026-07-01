import { AppShell } from '@/components/app-shell'
import { Skeleton } from '@/components/ui/skeleton'
import { Card, CardContent } from '@/components/ui/card'
import { TableRowSkeleton } from '@/components/ui/skeleton-row'
import { Table, TableBody, TableHead, TableHeader, TableRow } from '@/components/ui/table'

export default function Loading() {
  return (
    <AppShell currentPath="/integrations/channels">
      <div className="grid gap-6">
        <div className="flex items-center justify-between">
          <div className="grid gap-2">
            <Skeleton className="h-7 w-64" />
            <Skeleton className="h-4 w-96 max-w-full" />
          </div>
          <Skeleton className="h-9 w-36" />
        </div>
        <Card>
          <CardContent className="overflow-x-auto p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Type</TableHead>
                  <TableHead>Target</TableHead>
                  <TableHead>Updated</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRowSkeleton columns={4} rows={5} />
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
