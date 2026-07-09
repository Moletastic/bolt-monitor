'use client'

import * as React from 'react'
import { Archive } from 'lucide-react'

import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'
import { archiveServiceAction } from '@/lib/actions'

export function ArchiveServiceButton({
  serviceId,
  serviceName,
  disabled,
}: {
  serviceId: string
  serviceName: string
  disabled?: boolean
}) {
  const formRef = React.useRef<HTMLFormElement | null>(null)

  return (
    <form action={archiveServiceAction} ref={formRef}>
      <input name="serviceId" type="hidden" value={serviceId} />
      <input name="returnTo" type="hidden" value={`/services/${serviceId}`} />
      <ConfirmDialog
        cancelLabel="Cancel"
        confirmLabel="Archive service"
        description={`Archive "${serviceName}"? The service becomes read-only and editing, monitor creation, and monitor enable/disable are disabled.`}
        disabled={disabled}
        onConfirm={() => {
          formRef.current?.requestSubmit()
        }}
        title="Archive service?"
        trigger={
          <Button className="gap-2" disabled={disabled} type="button" variant="outline">
            <Archive aria-hidden="true" className="h-4 w-4" />
            Archive service
          </Button>
        }
      />
    </form>
  )
}
