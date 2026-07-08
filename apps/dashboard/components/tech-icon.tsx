import type { ReactElement, ReactNode } from 'react'

import { isServiceCategory, type ServiceCategory } from '@/lib/types'
import { cn } from '@/lib/utils'

interface TechIconProps {
  category?: ServiceCategory | string
  className?: string
}

const stroke = 'currentColor'

function SvgIcon({ className, children }: { className: string; children: ReactNode }) {
  return (
    <svg
      className={className}
      fill="none"
      stroke={stroke}
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.6"
      viewBox="0 0 24 24"
    >
      {children}
    </svg>
  )
}

function Database({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <ellipse cx="12" cy="5" rx="8" ry="2.5" />
      <path d="M4 5v6c0 1.4 3.6 2.5 8 2.5s8-1.1 8-2.5V5" />
      <path d="M4 11v6c0 1.4 3.6 2.5 8 2.5s8-1.1 8-2.5v-6" />
    </SvgIcon>
  )
}

function GlobeHttp({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="12" cy="12" r="8" />
      <path d="M4 12h16" />
      <path d="M12 4c2.5 2.7 3.8 5.4 3.8 8s-1.3 5.3-3.8 8" />
      <path d="M12 4c-2.5 2.7-3.8 5.4-3.8 8s1.3 5.3 3.8 8" />
    </SvgIcon>
  )
}

function CacheDiamond({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M12 3 4 8v8l8 5 8-5V8z" />
      <path d="M4 8l8 5 8-5" />
      <path d="M12 13v8" />
    </SvgIcon>
  )
}

function Queue({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <rect x="3" y="6" width="6" height="12" rx="1" />
      <rect x="11" y="9" width="6" height="6" rx="1" />
      <path d="M19 8v8" />
    </SvgIcon>
  )
}

function Container({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <rect x="3" y="7" width="18" height="11" rx="1.5" />
      <path d="M3 11h18" />
      <path d="M8 7v-2M16 7v-2" />
    </SvgIcon>
  )
}

function FunctionIcon({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M9 18 3 12l6-6" />
      <path d="m15 6 6 6-6 6" />
      <path d="M14 4l-4 16" />
    </SvgIcon>
  )
}

function Server({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <rect x="4" y="3" width="16" height="7" rx="1.5" />
      <rect x="4" y="14" width="16" height="7" rx="1.5" />
      <path d="M7 6.5h.01M7 17.5h.01" />
    </SvgIcon>
  )
}

function Web({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <rect x="3" y="5" width="18" height="14" rx="2" />
      <path d="M3 9h18" />
      <path d="M7 15h5M15 15h2" />
    </SvgIcon>
  )
}

function API({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M7 8h10M7 12h7M7 16h10" />
      <path d="M4 8h.01M4 12h.01M4 16h.01" />
      <path d="M20 8v8" />
    </SvgIcon>
  )
}

function Worker({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="12" cy="12" r="3" />
      <path d="M12 3v3M12 18v3M3 12h3M18 12h3" />
      <path d="m5.6 5.6 2.1 2.1M16.3 16.3l2.1 2.1M18.4 5.6l-2.1 2.1M7.7 16.3l-2.1 2.1" />
    </SvgIcon>
  )
}

function Scheduler({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="12" cy="12" r="8" />
      <path d="M12 7v5l3 2" />
      <path d="M7 3v3M17 3v3" />
    </SvgIcon>
  )
}

function Storage({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M4 8h5l2 3h9v7.5A1.5 1.5 0 0 1 18.5 20h-13A1.5 1.5 0 0 1 4 18.5z" />
      <path d="M4 8V5.5A1.5 1.5 0 0 1 5.5 4H9l2 3h7.5A1.5 1.5 0 0 1 20 8.5V11" />
    </SvgIcon>
  )
}

function Search({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="10.5" cy="10.5" r="5.5" />
      <path d="m15 15 5 5" />
      <path d="M8 10.5h5" />
    </SvgIcon>
  )
}

function Auth({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <rect x="5" y="10" width="14" height="10" rx="2" />
      <path d="M8 10V7a4 4 0 0 1 8 0v3" />
      <path d="M12 14v2" />
    </SvgIcon>
  )
}

function Payments({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <rect x="3" y="6" width="18" height="12" rx="2" />
      <path d="M3 10h18" />
      <path d="M7 15h4" />
      <circle cx="17" cy="15" r="1" />
    </SvgIcon>
  )
}

function Analytics({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M4 19V5" />
      <path d="M4 19h16" />
      <rect x="7" y="11" width="3" height="5" rx="1" />
      <rect x="12" y="7" width="3" height="9" rx="1" />
      <rect x="17" y="9" width="3" height="7" rx="1" />
    </SvgIcon>
  )
}

function Observability({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M3 12s3-6 9-6 9 6 9 6-3 6-9 6-9-6-9-6" />
      <circle cx="12" cy="12" r="2.5" />
      <path d="M18 19l2 2" />
    </SvgIcon>
  )
}

function AI({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M8 8a4 4 0 0 1 8 0v8a4 4 0 0 1-8 0z" />
      <path d="M8 11h8M8 14h8" />
      <path d="M5 9H3M5 15H3M21 9h-2M21 15h-2" />
    </SvgIcon>
  )
}

function Integration({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="7" cy="7" r="3" />
      <circle cx="17" cy="17" r="3" />
      <path d="M10 7h3a4 4 0 0 1 4 4v3" />
      <path d="M14 17h-3a4 4 0 0 1-4-4v-3" />
    </SvgIcon>
  )
}

function Media({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <rect x="4" y="5" width="16" height="14" rx="2" />
      <path d="m10 9 5 3-5 3z" />
      <path d="M4 17l4-4 3 3" />
    </SvgIcon>
  )
}

function Content({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M7 3h7l4 4v14H7z" />
      <path d="M14 3v5h4" />
      <path d="M10 12h5M10 16h5" />
    </SvgIcon>
  )
}

function Finance({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M5 19V5" />
      <path d="M5 19h14" />
      <path d="M8 16l3-4 3 2 4-6" />
      <circle cx="17" cy="17" r="2" />
    </SvgIcon>
  )
}

function Learning({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M4 6.5c2-1 4-1 8 1 4-2 6-2 8-1v12c-2-1-4-1-8 1-4-2-6-2-8-1z" />
      <path d="M12 7.5v12" />
    </SvgIcon>
  )
}

function Gaming({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M7 15H5a3 3 0 0 1 0-6h14a3 3 0 0 1 0 6h-2" />
      <path d="M8 12h4M10 10v4" />
      <path d="M16 11h.01M18 13h.01" />
    </SvgIcon>
  )
}

function Commerce({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M5 6h2l2 9h8l2-6H9" />
      <circle cx="10" cy="19" r="1.5" />
      <circle cx="17" cy="19" r="1.5" />
    </SvgIcon>
  )
}

function Messaging({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M5 6h14v9H8l-3 3z" />
      <path d="M8 10h8M8 13h5" />
    </SvgIcon>
  )
}

function Support({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="12" cy="12" r="8" />
      <path d="M9 10a3 3 0 1 1 4 2.8c-.7.3-1 .8-1 1.7" />
      <path d="M12 17h.01" />
    </SvgIcon>
  )
}

function Marketing({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M4 13V9l11-4v12z" />
      <path d="M15 8h2a3 3 0 0 1 0 6h-2" />
      <path d="M7 13l2 6" />
    </SvgIcon>
  )
}

function Admin({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="12" cy="8" r="3" />
      <path d="M5 20a7 7 0 0 1 14 0" />
      <path d="M17 4l2 2M19 4l-2 2" />
    </SvgIcon>
  )
}

function Security({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M12 3 5 6v5c0 4.5 2.8 7.7 7 10 4.2-2.3 7-5.5 7-10V6z" />
      <path d="m9 12 2 2 4-4" />
    </SvgIcon>
  )
}

function Location({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <path d="M12 21s6-5.4 6-11a6 6 0 1 0-12 0c0 5.6 6 11 6 11" />
      <circle cx="12" cy="10" r="2" />
    </SvgIcon>
  )
}

function Social({ className }: { className: string }) {
  return (
    <SvgIcon className={className}>
      <circle cx="8" cy="8" r="3" />
      <circle cx="17" cy="10" r="2.5" />
      <path d="M3 20a5 5 0 0 1 10 0" />
      <path d="M14 20a4 4 0 0 1 7 0" />
    </SvgIcon>
  )
}

const CATEGORY_GLYPHS: Record<ServiceCategory, (props: { className: string }) => ReactElement> = {
  server: Server,
  database: Database,
  cache: CacheDiamond,
  http: GlobeHttp,
  queue: Queue,
  container: Container,
  function: FunctionIcon,
  web: Web,
  api: API,
  worker: Worker,
  scheduler: Scheduler,
  storage: Storage,
  search: Search,
  auth: Auth,
  payments: Payments,
  analytics: Analytics,
  observability: Observability,
  ai: AI,
  integration: Integration,
  media: Media,
  content: Content,
  finance: Finance,
  learning: Learning,
  gaming: Gaming,
  commerce: Commerce,
  messaging: Messaging,
  support: Support,
  marketing: Marketing,
  admin: Admin,
  security: Security,
  location: Location,
  social: Social,
}

export function TechIcon({ category, className }: TechIconProps) {
  const Glyph = category && isServiceCategory(category) ? CATEGORY_GLYPHS[category] : Server
  return <Glyph className={cn('h-5 w-5', className)} />
}
