'use client'

import { useActionState, useEffect, useRef, useState } from 'react'
import { MoreVertical, Power, PowerOff } from 'lucide-react'

import { samePageActionStartEvent } from '@/components/query-feedback-banner'
import { toggleMonitorStateAction } from '@/lib/actions'
import { actionErrorMessage, idleActionState } from '@/lib/action-state'

export function MonitorActionsMenu({
  serviceId,
  monitorId,
  enabled,
  returnTo,
  disabled,
}: {
  serviceId: string
  monitorId: string
  enabled: boolean
  returnTo: string
  disabled?: boolean
}) {
  const [open, setOpen] = useState(false)
  const [openAbove, setOpenAbove] = useState(false)
  const containerRef = useRef<HTMLDivElement | null>(null)
  const triggerRef = useRef<HTMLButtonElement | null>(null)
  const [state, formAction, pending] = useActionState(toggleMonitorStateAction, idleActionState)
  const [locallyPending, setLocallyPending] = useState(false)
  const isPending = pending || locallyPending

  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }
    function handleEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        setOpen(false)
      }
    }
    if (open) {
      document.addEventListener('mousedown', handleClickOutside)
      document.addEventListener('keydown', handleEscape)
      return () => {
        document.removeEventListener('mousedown', handleClickOutside)
        document.removeEventListener('keydown', handleEscape)
      }
    }
  }, [open])

  useEffect(() => {
    if (!open) {
      return
    }
    function recompute() {
      const trigger = triggerRef.current
      if (!trigger) {
        return
      }
      const rect = trigger.getBoundingClientRect()
      const spaceBelow = window.innerHeight - rect.bottom
      setOpenAbove(spaceBelow < 96)
    }
    recompute()
    window.addEventListener('resize', recompute)
    window.addEventListener('scroll', recompute, true)
    return () => {
      window.removeEventListener('resize', recompute)
      window.removeEventListener('scroll', recompute, true)
    }
  }, [open])

  useEffect(() => {
    if (state.status !== 'idle') {
      setLocallyPending(false)
      setOpen(false)
    }
  }, [state])

  const label = enabled ? 'Disable monitor' : 'Enable monitor'
  const pendingLabel = enabled ? 'Disabling...' : 'Enabling...'
  const ToggleIcon = enabled ? PowerOff : Power
  const iconClass = enabled ? 'text-status-down' : 'text-status-up'

  return (
    <div className="relative inline-block" ref={containerRef}>
      <button
        aria-expanded={open}
        aria-haspopup="menu"
        aria-label="Monitor actions"
        className="inline-flex h-8 w-8 items-center justify-center rounded-md border border-border bg-transparent text-foreground transition-colors hover:bg-surface-low focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50"
        disabled={disabled}
        onClick={(event) => {
          event.preventDefault()
          event.stopPropagation()
          setOpen((value) => !value)
        }}
        ref={triggerRef}
        type="button"
      >
        <MoreVertical aria-hidden="true" className="h-4 w-4" />
      </button>
      {open ? (
        <div
          className={`absolute right-0 z-20 min-w-[12rem] overflow-hidden rounded-md border border-border bg-surface py-1 shadow-md ${openAbove ? 'bottom-full mb-1' : 'top-full mt-1'}`}
          role="menu"
        >
          <form
            action={formAction}
            className="contents"
            onSubmit={() => {
              setLocallyPending(true)
              window.dispatchEvent(new Event(samePageActionStartEvent))
            }}
          >
            <input name="serviceId" type="hidden" value={serviceId} />
            <input name="monitorId" type="hidden" value={monitorId} />
            <input name="enabled" type="hidden" value={enabled ? 'false' : 'true'} />
            <input name="returnTo" type="hidden" value={returnTo} />
            <button
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-foreground transition-colors hover:bg-surface-low disabled:pointer-events-none disabled:opacity-50"
              disabled={isPending || disabled}
              role="menuitem"
              type="submit"
            >
              <ToggleIcon aria-hidden="true" className={`h-4 w-4 flex-shrink-0 ${iconClass}`} />
              {isPending ? pendingLabel : label}
            </button>
          </form>
          {state.status === 'success' && state.message ? (
            <p
              aria-live="polite"
              className="border-t border-border px-3 py-2 text-xs text-status-up"
              role="status"
            >
              {state.message}
            </p>
          ) : null}
          {state.status === 'error' ? (
            <p className="border-t border-border px-3 py-2 text-xs text-status-down" role="alert">
              {actionErrorMessage(state.error)}
            </p>
          ) : null}
        </div>
      ) : null}
    </div>
  )
}
