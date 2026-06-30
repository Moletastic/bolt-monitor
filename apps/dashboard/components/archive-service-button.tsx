'use client'

import * as React from 'react'

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
          <Button disabled={disabled} type="button" variant="outline">
            Archive service
          </Button>
        }
      />
    </form>
  )
}
