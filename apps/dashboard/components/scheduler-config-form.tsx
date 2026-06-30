'use client'

import { useActionState } from 'react'

import { SubmitButton } from '@/components/submit-button'
import { idleActionState } from '@/lib/action-state'
import { updateSchedulerConfigStateAction } from '@/lib/actions'
import { schedulerConfigFeedback } from '@/lib/scheduler-config-feedback'

type SchedulerConfigFormProps = {
  recurringEnabled: boolean
}

export function SchedulerConfigForm({ recurringEnabled }: SchedulerConfigFormProps) {
  const [state, formAction] = useActionState(updateSchedulerConfigStateAction, idleActionState)
  const feedback = schedulerConfigFeedback(state)

  return (
    <form action={formAction}>
      {feedback ? (
        <p
          className={
            feedback.tone === 'success'
              ? 'mb-4 rounded-md border border-status-up/30 bg-status-up/10 px-3 py-2 text-sm text-status-up'
              : 'mb-4 rounded-md border border-destructive/30 bg-destructive/10 px-3 py-2 text-sm text-destructive'
          }
        >
          {feedback.message}
          {feedback.tone === 'error' ? (
            <span className="ml-2 text-xs opacity-80">Code: {feedback.code}</span>
          ) : null}
        </p>
      ) : null}
      <div className="mb-4 flex items-center gap-4">
        <label className="text-sm font-medium" htmlFor="recurringEnabled">
          Enable recurring execution
        </label>
        <input
          type="checkbox"
          id="recurringEnabled"
          name="recurringEnabled"
          value="true"
          defaultChecked={recurringEnabled}
        />
      </div>
      <SubmitButton type="submit">Save changes</SubmitButton>
    </form>
  )
}
