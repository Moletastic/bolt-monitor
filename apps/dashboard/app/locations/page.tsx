import { AppShell } from '@/components/app-shell'
import { EmptyState } from '@/components/empty-state'
import { UnavailableCard } from '@/components/unavailable-card'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { listProbeLocations } from '@/lib/api'

export default async function LocationsPage() {
  let locations
  try {
    locations = await listProbeLocations()
  } catch (error) {
    const message = error instanceof Error ? error.message : 'Unable to load probe locations.'
    return (
      <AppShell currentPath="/locations">
        <div className="grid gap-6">
          <div>
            <h1 className="text-2xl font-semibold tracking-tight">Probe Locations</h1>
            <p className="mt-1 text-sm text-muted-foreground">
              Available probe regions for monitor execution.
            </p>
          </div>
          <UnavailableCard message={message} title="Probe locations unavailable" />
        </div>
      </AppShell>
    )
  }

  return (
    <AppShell currentPath="/locations">
      <div className="grid gap-6">
        <div>
          <h1 className="text-2xl font-semibold tracking-tight">Probe Locations</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Available probe regions for monitor execution.
          </p>
        </div>
        <Card>
          <CardHeader>
            <CardTitle>Locations</CardTitle>
          </CardHeader>
          <CardContent>
            {locations.length === 0 ? (
              <EmptyState description="No probe locations configured." title="No locations" />
            ) : (
              <div className="overflow-x-auto">
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>Location ID</TableHead>
                      <TableHead>Display name</TableHead>
                      <TableHead>Enabled</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {locations.map((location) => (
                      <TableRow key={location.locationId}>
                        <TableCell className="font-mono font-medium">
                          {location.locationId}
                        </TableCell>
                        <TableCell>{location.displayName}</TableCell>
                        <TableCell>{location.enabled ? 'Yes' : 'No'}</TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}
          </CardContent>
        </Card>
      </div>
    </AppShell>
  )
}
