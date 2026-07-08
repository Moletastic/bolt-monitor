import Link from 'next/link'
import type { ReactNode } from 'react'
import { Server, AlertTriangle, Settings, LayoutDashboard, Bell, RadioTower } from 'lucide-react'

import { Breadcrumbs, type BreadcrumbItem } from '@/components/breadcrumbs'
import { cn } from '@/lib/utils'

const navItems = [
  { href: '/', label: 'Dashboard', matches: ['/'], icon: LayoutDashboard },
  { href: '/services', label: 'Services', matches: ['/services', '/monitors'], icon: Server },
  { href: '/policies', label: 'Notification routes', matches: ['/policies'], icon: Bell },
  {
    href: '/integrations/channels',
    label: 'Channels',
    matches: ['/integrations'],
    icon: RadioTower,
  },
  {
    href: '/incidents',
    label: 'Incidents',
    matches: ['/audit-trail', '/incidents'],
    icon: AlertTriangle,
  },
  { href: '/config', label: 'Settings', matches: ['/config', '/admin'], icon: Settings },
]

function isNavItemActive(currentPath: string, item: (typeof navItems)[number]) {
  return item.matches.some((match) =>
    match === '/'
      ? currentPath === '/'
      : currentPath === match || currentPath.startsWith(`${match}/`)
  )
}

export function AppShell({
  children,
  currentPath,
  breadcrumbs,
}: {
  children: ReactNode
  currentPath: string
  breadcrumbs?: BreadcrumbItem[]
}) {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <div className="flex min-h-screen flex-col lg:flex-row">
        <aside className="border-b border-border bg-surface-lowest/95 lg:sticky lg:top-0 lg:h-screen lg:w-72 lg:border-b-0 lg:border-r">
          <div className="flex h-full flex-col px-5 py-6">
            <div className="mb-8 space-y-2">
              <p className="text-[11px] font-bold uppercase tracking-[0.3em] text-primary">
                Bolt Monitor
              </p>
              <span className="block text-2xl font-semibold tracking-tight text-foreground">
                Operator Console
              </span>
              <p className="text-sm text-muted-foreground">
                Module-oriented workspace for dashboard landing, services, notification routes, and
                config.
              </p>
            </div>
            <nav className="grid gap-2">
              {navItems.map((item) => {
                const active = isNavItemActive(currentPath, item)
                const Icon = item.icon
                return (
                  <Link
                    className={cn(
                      'rounded-lg border px-4 py-3 text-sm font-semibold transition-colors',
                      active
                        ? 'border-primary/40 bg-primary/10 text-primary'
                        : 'border-transparent bg-transparent text-muted-foreground hover:border-border hover:bg-surface-low hover:text-foreground'
                    )}
                    href={item.href}
                    key={item.href}
                  >
                    <span className="flex items-center gap-2">
                      {Icon && <Icon className="h-4 w-4" />}
                      {item.label}
                    </span>
                  </Link>
                )
              })}
            </nav>
            <div className="mt-auto" />
          </div>
        </aside>
        <div className="flex-1">
          <main className="data-grid min-h-screen px-5 py-6 lg:px-8" id="main-content">
            {breadcrumbs && breadcrumbs.length > 0 ? (
              <Breadcrumbs items={breadcrumbs} />
            ) : null}
            {children}
          </main>
        </div>
      </div>
    </div>
  )
}
