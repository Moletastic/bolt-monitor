'use client'

import Link from 'next/link'
import { useSearchParams } from 'next/navigation'

import { cn } from '@/lib/utils'

interface Tab {
  label: string
  href: string
}

interface TabsProps {
  basePath: string
  tabs: Tab[]
  className?: string
}

export function Tabs({ tabs, className }: TabsProps) {
  const searchParams = useSearchParams()
  const currentTab = searchParams.get('tab') ?? tabs[0]?.href.split('tab=')[1] ?? ''

  return (
    <div className={cn('flex gap-1 border-b border-border', className)} role="tablist">
      {tabs.map((tab) => {
        const tabValue = tab.href.split('tab=')[1]
        const isActive = currentTab === tabValue
        return (
          <Link
            aria-selected={isActive}
            key={tab.href}
            href={tab.href}
            role="tab"
            className={cn(
              'relative px-4 py-2 text-sm font-medium transition-colors hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
              isActive ? 'text-foreground' : 'text-muted-foreground'
            )}
          >
            {tab.label}
            {isActive && <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-foreground" />}
          </Link>
        )
      })}
    </div>
  )
}
