'use client'

import * as React from 'react'

import { Button } from '@/components/ui/button'
import { ConfirmDialog } from '@/components/ui/confirm-dialog'

export function DeleteResourceForm({
  action,
  children,
  confirmMessage,
  disabled,
  focusNextOnSuccess = true,
  label,
}: {
  action: (formData: FormData) => Promise<void>
  children: React.ReactNode
  confirmMessage: string
  disabled?: boolean
  focusNextOnSuccess?: boolean
  label: string
}) {
  const formRef = React.useRef<HTMLFormElement | null>(null)

  return (
    <form action={action} data-focus-next-on-success={focusNextOnSuccess} ref={formRef}>
      {children}
      <ConfirmDialog
        cancelLabel="Cancel"
        confirmLabel={label}
        description={confirmMessage}
        disabled={disabled}
        onConfirm={() => {
          formRef.current?.requestSubmit()
        }}
        title={`Confirm ${label.toLowerCase()}`}
        trigger={
          <Button disabled={disabled} type="button" variant="destructive">
            {label}
          </Button>
        }
      />
    </form>
  )
}
