import { type ComponentType } from 'react'
import * as Devicon from 'devicons-react'
import { Server } from 'lucide-react'
import { cn } from '@/lib/utils'
import type { TechnologyKey } from '@/lib/types'

interface ServiceIconProps {
  technologyKey?: TechnologyKey
  className?: string
  size?: 'sm' | 'md' | 'lg'
}

type DeviconComponent = ComponentType<{ className?: string; size?: number | string }>

const iconMap: Record<string, DeviconComponent> = {
  golang: Devicon.GoOriginal as DeviconComponent,
  mariadb: Devicon.MariadbOriginal as DeviconComponent,
  mysql: Devicon.MysqlOriginal as DeviconComponent,
  nginx: Devicon.NginxOriginal as DeviconComponent,
  postgres: Devicon.PostgresqlOriginal as DeviconComponent,
  python: Devicon.PythonOriginal as DeviconComponent,
  typescript: Devicon.TypescriptOriginal as DeviconComponent,
  mongodb: Devicon.MongodbOriginal as DeviconComponent,
  redis: Devicon.RedisOriginal as DeviconComponent,
  kafka: Devicon.ApachekafkaOriginal as DeviconComponent,
  docker: Devicon.DockerOriginal as DeviconComponent,
  apache: Devicon.ApacheOriginal as DeviconComponent,
  javascript: Devicon.JavascriptOriginal as DeviconComponent,
  rabbitmq: Devicon.RabbitmqOriginal as DeviconComponent,
}

const sizeClass = {
  sm: { frame: 'h-9 w-9', icon: 'h-6 w-6', px: 36 },
  md: { frame: 'h-11 w-11', icon: 'h-8 w-8', px: 44 },
  lg: { frame: 'h-14 w-14', icon: 'h-10 w-10', px: 56 },
}

export function ServiceIcon({ technologyKey, className, size = 'md' }: ServiceIconProps) {
  const classes = sizeClass[size]
  const iconClassName = cn('brightness-0 invert', classes.icon)

  if (!technologyKey) {
    return (
      <span
        className={cn(
          'inline-flex shrink-0 items-center justify-center rounded-xl bg-surface-high',
          classes.frame,
          className
        )}
      >
        <Server className={iconClassName} size={classes.px} />
      </span>
    )
  }

  const IconComponent = iconMap[technologyKey]

  if (!IconComponent) {
    return (
      <span
        className={cn(
          'inline-flex shrink-0 items-center justify-center rounded-xl bg-surface-high',
          classes.frame,
          className
        )}
      >
        <Server className={iconClassName} size={classes.px} />
      </span>
    )
  }

  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center justify-center rounded-xl bg-surface-high',
        classes.frame,
        className
      )}
    >
      <IconComponent className={iconClassName} size={classes.px} />
    </span>
  )
}
