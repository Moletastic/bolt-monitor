'use client'

import { History, ScrollText, ShieldAlert } from 'lucide-react'
import Link from 'next/link'
import { useSearchParams } from 'next/navigation'

import { cn } from '@/lib/utils'

const icons = {
  history: History,
  incidents: ShieldAlert,
  audit: ScrollText,
}

interface Tab {
  label: string
  href: string
  iconName?: keyof typeof icons
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
        const Icon = tab.iconName ? icons[tab.iconName] : undefined
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
            {Icon && <Icon aria-hidden="true" className="mr-2 inline h-4 w-4" />}
            {tab.label}
            {isActive && <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-foreground" />}
          </Link>
        )
      })}
    </div>
  )
}
