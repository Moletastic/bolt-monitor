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

  function moveFocusAfterDelete() {
    if (!focusNextOnSuccess) return
    const form = formRef.current
    if (!form) return
    const root = form.closest('[data-list-root]') as HTMLElement | null
    if (root) {
      const items = Array.from(root.querySelectorAll<HTMLElement>('[data-list-item]'))
      const currentItem = form.closest('[data-list-item]') as HTMLElement | null
      if (currentItem) {
        const idx = items.indexOf(currentItem)
        const next = items[idx + 1] ?? items[idx - 1] ?? null
        if (next) {
          const focusable = next.querySelector<HTMLElement>(
            'a[href], button:not([disabled]), [tabindex]:not([tabindex="-1"])'
          )
          focusable?.focus()
          return
        }
      }
      const createCta = root.querySelector<HTMLElement>('[data-create-cta]')
      if (createCta) {
        createCta.focus()
        return
      }
    }
    const fallback = document.querySelector<HTMLElement>('[data-create-cta]')
    if (fallback) fallback.focus()
  }

  return (
    <form
      action={async (formData) => {
        await action(formData)
        moveFocusAfterDelete()
      }}
      ref={formRef}
    >
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
