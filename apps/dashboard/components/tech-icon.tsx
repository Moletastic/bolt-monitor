import type { ReactElement } from 'react'

import { cn } from '@/lib/utils'
import type { ServiceCategory } from '@/lib/types'

interface TechIconProps {
  category?: ServiceCategory
  className?: string
}

const stroke = 'currentColor'

function Database({ className }: { className: string }) {
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
      <ellipse cx="12" cy="5" rx="8" ry="2.5" />
      <path d="M4 5v6c0 1.4 3.6 2.5 8 2.5s8-1.1 8-2.5V5" />
      <path d="M4 11v6c0 1.4 3.6 2.5 8 2.5s8-1.1 8-2.5v-6" />
    </svg>
  )
}

function GlobeHttp({ className }: { className: string }) {
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
      <circle cx="12" cy="12" r="8" />
      <path d="M4 12h16" />
      <path d="M12 4c2.5 2.7 3.8 5.4 3.8 8s-1.3 5.3-3.8 8" />
      <path d="M12 4c-2.5 2.7-3.8 5.4-3.8 8s1.3 5.3 3.8 8" />
    </svg>
  )
}

function CacheDiamond({ className }: { className: string }) {
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
      <path d="M12 3 4 8v8l8 5 8-5V8z" />
      <path d="M4 8l8 5 8-5" />
      <path d="M12 13v8" />
    </svg>
  )
}

function Queue({ className }: { className: string }) {
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
      <rect x="3" y="6" width="6" height="12" rx="1" />
      <rect x="11" y="9" width="6" height="6" rx="1" />
      <path d="M19 8v8" />
    </svg>
  )
}

function Container({ className }: { className: string }) {
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
      <rect x="3" y="7" width="18" height="11" rx="1.5" />
      <path d="M3 11h18" />
      <path d="M8 7v-2M16 7v-2" />
    </svg>
  )
}

function Function({ className }: { className: string }) {
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
      <path d="M9 18 3 12l6-6" />
      <path d="m15 6 6 6-6 6" />
      <path d="M14 4l-4 16" />
    </svg>
  )
}

function Server({ className }: { className: string }) {
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
      <rect x="4" y="3" width="16" height="7" rx="1.5" />
      <rect x="4" y="14" width="16" height="7" rx="1.5" />
      <path d="M7 6.5h.01M7 17.5h.01" />
    </svg>
  )
}

const CATEGORY_GLYPHS: Record<ServiceCategory, (props: { className: string }) => ReactElement> = {
  server: Server,
  database: Database,
  cache: CacheDiamond,
  http: GlobeHttp,
  queue: Queue,
  container: Container,
  function: Function,
}

export function TechIcon({ category, className }: TechIconProps) {
  const Glyph = category ? CATEGORY_GLYPHS[category] : Server
  return <Glyph className={cn('h-5 w-5', className)} />
}
