'use client'

import * as React from 'react'
import * as AlertDialog from '@radix-ui/react-alert-dialog'
import { AlertOctagon, X } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { deleteServiceAction } from '@/lib/actions'

export function DeleteServiceConfirmDialog({
  serviceId,
  serviceName,
  returnTo,
  disabled,
}: {
  serviceId: string
  serviceName: string
  returnTo: string
  disabled?: boolean
}) {
  const [open, setOpen] = React.useState(false)
  const [typed, setTyped] = React.useState('')
  const expected = serviceName.trim()
  const matches = typed.trim() === expected && expected.length > 0

  React.useEffect(() => {
    if (!open) {
      setTyped('')
    }
  }, [open])

  return (
    <AlertDialog.Root onOpenChange={setOpen} open={open}>
      <AlertDialog.Trigger asChild>
        <Button disabled={disabled} type="button" variant="destructive">
          Delete service
        </Button>
      </AlertDialog.Trigger>
      <AlertDialog.Portal>
        <AlertDialog.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" />
        <AlertDialog.Content className="fixed left-1/2 top-1/2 z-50 w-full max-w-md -translate-x-1/2 -translate-y-1/2 rounded-lg border border-border bg-surface p-6 shadow-xl">
          <div className="flex items-start justify-between gap-4">
            <div className="flex items-center gap-2">
              <AlertOctagon aria-hidden="true" className="h-5 w-5 text-status-down" />
              <AlertDialog.Title className="text-lg font-semibold text-foreground">
                Delete service
              </AlertDialog.Title>
            </div>
            <AlertDialog.Cancel asChild>
              <button
                aria-label="Close delete service dialog"
                className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-surface-low hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background"
                type="button"
              >
                <X aria-hidden="true" className="h-5 w-5" />
              </button>
            </AlertDialog.Cancel>
          </div>
          <AlertDialog.Description className="mt-3 text-sm text-muted-foreground">
            Deleting <strong className="text-foreground">{serviceName}</strong> is permanent and
            cannot be undone. All monitor configuration for this service is removed from active
            management views.
          </AlertDialog.Description>
          <form action={deleteServiceAction} className="mt-4 space-y-4">
            <input name="serviceId" type="hidden" value={serviceId} />
            <input name="returnTo" type="hidden" value={returnTo} />
            <div className="space-y-2">
              <label
                className="block text-sm font-medium text-foreground"
                htmlFor={`delete-confirm-${serviceId}`}
              >
                Type <span className="font-mono text-foreground">{serviceName}</span> to confirm
              </label>
              <input
                aria-describedby={`delete-confirm-${serviceId}-hint`}
                autoComplete="off"
                className="w-full rounded-md border border-border bg-surface-low px-3 py-2 text-sm text-foreground focus:border-primary focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 focus:ring-offset-background"
                id={`delete-confirm-${serviceId}`}
                onChange={(event) => setTyped(event.target.value)}
                placeholder={serviceName}
                type="text"
                value={typed}
              />
              <p className="text-xs text-muted-foreground" id={`delete-confirm-${serviceId}-hint`}>
                The confirm button enables once the name matches exactly.
              </p>
            </div>
            <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
              <AlertDialog.Cancel asChild>
                <Button type="button" variant="outline">
                  Cancel
                </Button>
              </AlertDialog.Cancel>
              <Button
                disabled={!matches}
                onClick={(event) => {
                  if (!matches) event.preventDefault()
                }}
                type="submit"
                variant="destructive"
              >
                Delete service
              </Button>
            </div>
          </form>
        </AlertDialog.Content>
      </AlertDialog.Portal>
    </AlertDialog.Root>
  )
}
