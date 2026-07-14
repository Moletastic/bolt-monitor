'use client'

import * as AlertDialog from '@radix-ui/react-alert-dialog'
import { AlertOctagon, MoreVertical, Power, PowerOff, Trash2, Wrench, X } from 'lucide-react'
import { useActionState, useEffect, useRef, useState } from 'react'

import { samePageActionStartEvent } from '@/components/query-feedback-banner'
import { Button } from '@/components/ui/button'
import {
  deleteMonitorAction,
  toggleMaintenanceModeAction,
  toggleMonitorStateAction,
} from '@/lib/actions'
import { actionErrorMessage, idleActionState } from '@/lib/action-state'

export function MonitorDetailActionsMenu({
  serviceId,
  monitorId,
  monitorName,
  enabled,
  inMaintenance,
  returnTo,
  deleteDisabled,
  deleteDisabledReason,
}: {
  serviceId: string
  monitorId: string
  monitorName: string
  enabled: boolean
  inMaintenance: boolean
  returnTo: string
  deleteDisabled?: boolean
  deleteDisabledReason?: string
}) {
  const [open, setOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [typed, setTyped] = useState('')
  const [state, formAction, pending] = useActionState(toggleMonitorStateAction, idleActionState)
  const [locallyPending, setLocallyPending] = useState(false)
  const containerRef = useRef<HTMLDivElement | null>(null)
  const expected = monitorName.trim()
  const typedMatches = typed.trim() === expected && expected.length > 0
  const toggleLabel = enabled ? 'Disable monitor' : 'Enable monitor'
  const ToggleIcon = enabled ? PowerOff : Power
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
    if (!open) {
      return
    }
    document.addEventListener('mousedown', handleClickOutside)
    document.addEventListener('keydown', handleEscape)
    return () => {
      document.removeEventListener('mousedown', handleClickOutside)
      document.removeEventListener('keydown', handleEscape)
    }
  }, [open])

  useEffect(() => {
    if (state.status !== 'idle') {
      setLocallyPending(false)
      setOpen(false)
    }
  }, [state.status])

  useEffect(() => {
    if (!deleteOpen) {
      setTyped('')
    }
  }, [deleteOpen])

  return (
    <div className="relative inline-block" ref={containerRef}>
      <Button
        aria-expanded={open}
        aria-haspopup="menu"
        aria-label="More monitor actions"
        className="px-3"
        onClick={() => setOpen((value) => !value)}
        type="button"
        variant="outline"
      >
        <MoreVertical aria-hidden="true" className="h-4 w-4" />
      </Button>
      {open ? (
        <div
          className="absolute right-0 z-20 mt-2 min-w-[14rem] overflow-hidden rounded-md border border-border bg-surface py-1 shadow-lg"
          role="menu"
        >
          <form action={toggleMaintenanceModeAction} className="contents">
            <input name="serviceId" type="hidden" value={serviceId} />
            <input name="monitorId" type="hidden" value={monitorId} />
            <input name="enabled" type="hidden" value={inMaintenance ? 'false' : 'true'} />
            <input name="returnTo" type="hidden" value={returnTo} />
            <button
              className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm font-medium text-foreground transition-colors hover:bg-surface-low"
              role="menuitem"
              type="submit"
            >
              <Wrench aria-hidden="true" className="h-4 w-4 flex-shrink-0 text-muted-foreground" />
              {inMaintenance ? 'Exit maintenance' : 'Enter maintenance'}
            </button>
          </form>
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
              disabled={isPending}
              role="menuitem"
              type="submit"
            >
              <ToggleIcon
                aria-hidden="true"
                className={`h-4 w-4 flex-shrink-0 ${enabled ? 'text-status-down' : 'text-status-up'}`}
              />
              {isPending ? 'Updating...' : toggleLabel}
            </button>
          </form>
          <button
            className="flex w-full items-center gap-2 border-t border-border px-3 py-2 text-left text-sm font-medium text-status-down transition-colors hover:bg-status-down/10 disabled:pointer-events-none disabled:opacity-50"
            disabled={deleteDisabled}
            onClick={() => {
              setOpen(false)
              setDeleteOpen(true)
            }}
            role="menuitem"
            type="button"
          >
            <Trash2 aria-hidden="true" className="h-4 w-4 flex-shrink-0" />
            Delete monitor
          </button>
          {deleteDisabled && deleteDisabledReason ? (
            <p className="border-t border-border px-3 py-2 text-xs text-status-warn">
              {deleteDisabledReason}
            </p>
          ) : null}
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

      <AlertDialog.Root onOpenChange={setDeleteOpen} open={deleteOpen}>
        <AlertDialog.Portal>
          <AlertDialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" />
          <AlertDialog.Content className="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-surface p-6 shadow-xl">
            <div className="flex items-start justify-between gap-4">
              <div className="flex items-center gap-2">
                <AlertOctagon aria-hidden="true" className="h-5 w-5 text-status-down" />
                <AlertDialog.Title className="text-lg font-semibold text-foreground">
                  Delete monitor
                </AlertDialog.Title>
              </div>
              <AlertDialog.Cancel asChild>
                <button
                  aria-label="Close delete monitor dialog"
                  className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-surface-low hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                  type="button"
                >
                  <X aria-hidden="true" className="h-5 w-5" />
                </button>
              </AlertDialog.Cancel>
            </div>
            <AlertDialog.Description className="mt-3 text-sm text-muted-foreground">
              Deleting <strong className="text-foreground">{monitorName}</strong> is permanent and
              cannot be undone. Type the monitor name to confirm deletion.
            </AlertDialog.Description>
            <form action={deleteMonitorAction} className="mt-4 space-y-4">
              <input name="serviceId" type="hidden" value={serviceId} />
              <input name="monitorId" type="hidden" value={monitorId} />
              <input name="returnTo" type="hidden" value={returnTo} />
              <div className="space-y-2">
                <label
                  className="block text-sm font-medium text-foreground"
                  htmlFor={`delete-monitor-${monitorId}`}
                >
                  Type <span className="font-mono text-foreground">{monitorName}</span> to confirm
                </label>
                <input
                  aria-describedby={`delete-monitor-${monitorId}-hint`}
                  autoComplete="off"
                  className="w-full rounded-md border border-border bg-surface-low px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:ring-offset-background"
                  id={`delete-monitor-${monitorId}`}
                  onChange={(event) => setTyped(event.target.value)}
                  placeholder={monitorName}
                  type="text"
                  value={typed}
                />
                <p
                  className="text-xs text-muted-foreground"
                  id={`delete-monitor-${monitorId}-hint`}
                >
                  Confirm button enables once the name matches exactly.
                </p>
              </div>
              <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
                <AlertDialog.Cancel asChild>
                  <Button type="button" variant="outline">
                    Cancel
                  </Button>
                </AlertDialog.Cancel>
                <Button
                  disabled={!typedMatches || deleteDisabled}
                  onClick={(event) => {
                    if (!typedMatches || deleteDisabled) event.preventDefault()
                  }}
                  type="submit"
                  variant="destructive"
                >
                  Delete monitor
                </Button>
              </div>
            </form>
          </AlertDialog.Content>
        </AlertDialog.Portal>
      </AlertDialog.Root>
    </div>
  )
}
