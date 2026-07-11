import { clsx, type ClassValue } from 'clsx'
import { isValid, parseISO } from 'date-fns'
import { twMerge } from 'tailwind-merge'

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}

export function formatDateTime(value?: string) {
  if (!value) {
    return 'Never'
  }

  const date = parseISO(value)
  if (!isValid(date)) {
    return value
  }

  return new Intl.DateTimeFormat('en-US', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

export function formatDuration(durationMs?: number) {
  if (durationMs === undefined || durationMs === null) {
    return 'n/a'
  }
  if (durationMs < 1000) {
    return `${durationMs}ms`
  }
  if (durationMs < 60_000) {
    return `${(durationMs / 1000).toFixed(1)}s`
  }
  return `${(durationMs / 60_000).toFixed(1)}m`
}

export const monitorCadenceOptions = [
  { value: 60, label: 'Every 1 minute' },
  { value: 120, label: 'Every 2 minutes' },
  { value: 180, label: 'Every 3 minutes' },
  { value: 300, label: 'Every 5 minutes' },
  { value: 600, label: 'Every 10 minutes' },
  { value: 900, label: 'Every 15 minutes' },
  { value: 1800, label: 'Every 30 minutes' },
  { value: 3600, label: 'Every 1 hour' },
]

export function formatMonitorCadence(intervalSeconds: number) {
  return (
    monitorCadenceOptions.find((option) => option.value === intervalSeconds)?.label ??
    `Every ${intervalSeconds}s`
  )
}

export function formatOutcome(status?: string) {
  if (!status) {
    return 'UNKNOWN'
  }
  return status.toUpperCase()
}

export function slugify(value: string) {
  return value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .replace(/-{2,}/g, '-')
}

export function getServiceIconLabel(technologyKey?: string) {
  switch (technologyKey) {
    case 'golang':
      return 'GO'
    case 'mariadb':
      return 'MA'
    case 'mysql':
      return 'MY'
    case 'nginx':
      return 'NG'
    case 'postgres':
      return 'PG'
    case 'python':
      return 'PY'
    case 'typescript':
      return 'TS'
    default:
      return 'SV'
  }
}

export function getMonitorIconLabel(type?: string, target?: string) {
  if (type === 'http' && target?.startsWith('https://')) {
    return 'HS'
  }
  if (type === 'http') {
    return 'HT'
  }
  return 'MN'
}
