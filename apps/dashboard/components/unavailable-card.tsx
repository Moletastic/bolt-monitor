import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'

export function UnavailableCard({ title, message }: { title: string; message: string }) {
  return (
    <Card className="border-status-warn/30 bg-status-warn/5">
      <CardHeader>
        <CardTitle>{title}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="text-sm text-status-warn">{message}</p>
      </CardContent>
    </Card>
  )
}
