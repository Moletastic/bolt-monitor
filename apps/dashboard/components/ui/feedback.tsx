import { CheckCircle2, CircleAlert, Info, TriangleAlert } from 'lucide-react'
import type { ComponentType, ReactNode, SVGProps } from 'react'

import { cn } from '@/lib/utils'

type FeedbackTone = 'error' | 'info' | 'success' | 'warning'

const feedbackStyles: Record<
  FeedbackTone,
  { icon: ComponentType<SVGProps<SVGSVGElement>>; className: string }
> = {
  error: {
    icon: CircleAlert,
    className: 'border-status-down/30 bg-status-down/10 text-status-down',
  },
  info: { icon: Info, className: 'border-primary/30 bg-primary/10 text-primary' },
  success: { icon: CheckCircle2, className: 'border-status-up/30 bg-status-up/10 text-status-up' },
  warning: {
    icon: TriangleAlert,
    className: 'border-status-warn/30 bg-status-warn/10 text-status-warn',
  },
}

export function Feedback({
  children,
  className,
  tone,
}: {
  children?: ReactNode
  className?: string
  tone: FeedbackTone
}) {
  const { icon: Icon, className: toneClassName } = feedbackStyles[tone]
  return (
    <p
      className={cn(
        'flex items-start gap-2 rounded border px-3 py-2 text-sm',
        toneClassName,
        className
      )}
      role={tone === 'error' ? 'alert' : 'status'}
    >
      <Icon aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
      <span>{children}</span>
    </p>
  )
}

export function Unavailable({ title, message }: { title: string; message: string }) {
  return (
    <section
      aria-live="polite"
      className="rounded-lg border border-status-warn/30 bg-status-warn/5 p-4"
    >
      <div className="flex items-start gap-2 text-status-warn">
        <TriangleAlert aria-hidden="true" className="mt-0.5 size-4 shrink-0" />
        <div>
          <h2 className="text-title-sm">{title}</h2>
          <p className="mt-1 text-sm">{message}</p>
        </div>
      </div>
    </section>
  )
}
