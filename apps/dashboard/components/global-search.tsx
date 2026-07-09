'use client'

import Link from 'next/link'
import { useEffect, useRef, useState } from 'react'
import { Bell, RadioTower, Search, Server, ShieldAlert, X } from 'lucide-react'

import { searchResources } from '@/lib/api'
import { ApiError, messageFor } from '@/lib/errors'
import type { GlobalSearchResult } from '@/lib/types'
import { cn } from '@/lib/utils'

const minQueryLength = 2
const debounceMs = 275

const iconByType = {
  service: Server,
  monitor: ShieldAlert,
  policy: Bell,
  channel: RadioTower,
}

export function GlobalSearch() {
  const [query, setQuery] = useState('')
  const [results, setResults] = useState<GlobalSearchResult[]>([])
  const [status, setStatus] = useState<'idle' | 'loading' | 'ready' | 'empty' | 'error'>('idle')
  const [errorMessage, setErrorMessage] = useState('')
  const [open, setOpen] = useState(false)
  const requestSeq = useRef(0)
  const normalizedQuery = query.trim().replace(/\s+/g, ' ')

  useEffect(() => {
    if (normalizedQuery.length < minQueryLength) {
      requestSeq.current += 1
      setResults([])
      setStatus('idle')
      setErrorMessage('')
      return
    }

    const seq = requestSeq.current + 1
    requestSeq.current = seq
    setStatus('loading')
    setOpen(true)
    const timeout = window.setTimeout(() => {
      void searchResources({ query: normalizedQuery, limit: 8 }).then(
        (nextResults) => {
          if (requestSeq.current !== seq) return
          setResults(nextResults)
          setStatus(nextResults.length > 0 ? 'ready' : 'empty')
          setErrorMessage('')
        },
        (error: unknown) => {
          if (requestSeq.current !== seq) return
          setResults([])
          setStatus('error')
          setErrorMessage(error instanceof ApiError ? messageFor(error) : 'Try again.')
        }
      )
    }, debounceMs)

    return () => window.clearTimeout(timeout)
  }, [normalizedQuery])

  const showPanel =
    open && status !== 'loading' && (normalizedQuery.length >= minQueryLength || status === 'idle')

  return (
    <div className="relative w-full max-w-2xl">
      <label className="sr-only" htmlFor="global-search">
        Search services, monitors, routes, and channels
      </label>
      <div className="relative">
        <Search className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <input
          aria-controls="global-search-results"
          aria-expanded={showPanel}
          aria-haspopup="listbox"
          autoComplete="off"
          className={cn(
            'h-11 w-full rounded-xl border border-border bg-surface-lowest pl-10 pr-20 text-sm shadow-sm outline-none transition focus:border-primary/50 focus:ring-2 focus:ring-primary/20',
            showPanel ? 'rounded-b-none' : null
          )}
          id="global-search"
          onBlur={() => window.setTimeout(() => setOpen(false), 120)}
          onChange={(event) => setQuery(event.target.value)}
          onFocus={() => setOpen(true)}
          onKeyDown={(event) => {
            if (event.key === 'Escape') setOpen(false)
          }}
          placeholder="Search services, monitors, routes, and channels..."
          role="combobox"
          type="text"
          value={query}
        />
        {status === 'loading' ? (
          <span className="pointer-events-none absolute right-10 top-1/2 flex h-5 w-5 -translate-y-1/2 items-center justify-center text-muted-foreground">
            <span className="global-search-spinner h-4 w-4 rounded-full border-2 border-current border-r-transparent" />
            <span className="sr-only">Searching</span>
          </span>
        ) : null}
        {query ? (
          <button
            aria-label="Clear search"
            className="absolute right-3 top-1/2 flex h-6 w-6 -translate-y-1/2 items-center justify-center rounded-md text-muted-foreground transition hover:bg-surface-high hover:text-foreground focus:bg-surface-high focus:text-foreground focus:outline-none focus:ring-2 focus:ring-primary/30"
            onClick={() => {
              setQuery('')
              setResults([])
              setStatus('idle')
              setErrorMessage('')
            }}
            type="button"
          >
            <X className="h-4 w-4" />
          </button>
        ) : null}
      </div>
      {showPanel ? (
        <div
          className="absolute left-0 right-0 z-30 mt-1 overflow-hidden rounded-b-xl border border-t-0 border-border bg-surface-lowest shadow-xl"
          id="global-search-results"
          role="listbox"
        >
          {status === 'idle' ? (
            <p className="px-4 py-3 text-sm text-muted-foreground">
              Type at least 2 characters to jump to resources.
            </p>
          ) : null}
          {status === 'empty' ? (
            <p className="px-4 py-3 text-sm text-muted-foreground">No matching resources found.</p>
          ) : null}
          {status === 'error' ? (
            <p className="px-4 py-3 text-sm text-destructive">
              Search failed. {errorMessage || 'Try again.'}
            </p>
          ) : null}
          {status === 'ready' ? (
            <div className="max-h-96 overflow-y-auto py-2">
              {results.map((result) => {
                const Icon = iconByType[result.type]
                return (
                  <Link
                    className={cn(
                      'flex gap-3 px-4 py-3 text-sm transition-colors hover:bg-surface-low focus:bg-surface-low focus:outline-none'
                    )}
                    href={result.href}
                    key={`${result.type}:${result.serviceId ?? ''}:${result.id}`}
                    onClick={() => setOpen(false)}
                    role="option"
                  >
                    <span className="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
                      <Icon className="h-4 w-4" />
                    </span>
                    <span className="min-w-0">
                      <span className="block truncate font-semibold text-foreground">
                        {result.label}
                      </span>
                      <span className="block truncate text-muted-foreground">
                        {result.description}
                      </span>
                    </span>
                  </Link>
                )
              })}
            </div>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
