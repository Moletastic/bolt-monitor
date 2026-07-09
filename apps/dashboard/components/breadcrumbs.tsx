import { ChevronRight } from 'lucide-react'
import Link from 'next/link'
import { Fragment } from 'react'

import { cn } from '@/lib/utils'

export type BreadcrumbItem = {
  label: string
  href?: string
}

export function Breadcrumbs({ items, className }: { items: BreadcrumbItem[]; className?: string }) {
  if (items.length === 0) {
    return null
  }

  const lastIndex = items.length - 1

  return (
    <nav aria-label="Breadcrumb" className={cn('mb-6', className)}>
      <ol className="flex flex-wrap items-center gap-x-1 gap-y-1 text-xs text-muted-foreground">
        {items.map((item, index) => {
          const isCurrent = index === lastIndex
          const isLast = isCurrent
          return (
            <Fragment key={`${item.label}-${index}`}>
              <li className="flex min-w-0 items-center">
                {isCurrent || !item.href ? (
                  <span
                    aria-current="page"
                    className="block max-w-[16rem] truncate font-semibold text-foreground sm:max-w-[20rem]"
                    title={item.label}
                  >
                    {item.label}
                  </span>
                ) : (
                  <Link
                    className="block max-w-[16rem] truncate rounded-sm font-medium text-muted-foreground outline-none transition-colors hover:text-primary focus-visible:ring-2 focus-visible:ring-primary/50 focus-visible:ring-offset-2 focus-visible:ring-offset-background sm:max-w-[20rem]"
                    href={item.href}
                    title={item.label}
                  >
                    {item.label}
                  </Link>
                )}
              </li>
              {!isLast ? (
                <li
                  aria-hidden="true"
                  className="flex select-none items-center px-1 text-muted-foreground/60"
                >
                  <ChevronRight className="h-3 w-3" aria-hidden="true" />
                </li>
              ) : null}
            </Fragment>
          )
        })}
      </ol>
    </nav>
  )
}
