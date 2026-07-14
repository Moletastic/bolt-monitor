'use client'

import { PlayCircle, Wrench } from 'lucide-react'
import { useFormStatus } from 'react-dom'

import { Button, type ButtonProps } from '@/components/ui/button'

const icons = {
  play: PlayCircle,
  wrench: Wrench,
}

type SubmitButtonProps = ButtonProps & {
  iconName?: keyof typeof icons
  pendingLabel?: string
}

export function SubmitButton({
  children,
  iconName,
  pendingLabel = 'Working...',
  ...props
}: SubmitButtonProps) {
  const { pending } = useFormStatus()
  const Icon = iconName ? icons[iconName] : undefined

  return (
    <Button type="submit" {...props} disabled={pending || props.disabled}>
      {Icon && <Icon aria-hidden="true" className="h-4 w-4" />}
      {pending ? pendingLabel : children}
    </Button>
  )
}
