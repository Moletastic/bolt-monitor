import { Unavailable } from '@/components/ui/feedback'

export function UnavailableCard({ title, message }: { title: string; message: string }) {
  return <Unavailable message={message} title={title} />
}
