import Link from 'next/link'

import type { EscalationPolicy } from '@/lib/types'

export function buildChannelUsageMap(policies: EscalationPolicy[]): Map<string, EscalationPolicy[]> {
  const usage = new Map<string, EscalationPolicy[]>()
  for (const policy of policies) {
    const seen = new Set<string>()
    const paths = [policy.businessHoursPath, policy.offHoursPath]
    for (const path of paths) {
      for (const step of path.steps) {
        if (seen.has(step.channelId)) continue
        seen.add(step.channelId)
        const list = usage.get(step.channelId) ?? []
        list.push(policy)
        usage.set(step.channelId, list)
      }
    }
  }
  return usage
}

export function ChannelUsageScope({
  policies,
}: {
  channelId: string
  policies: EscalationPolicy[]
}) {
  if (policies.length === 0) {
    return (
      <span className="text-xs font-medium text-muted-foreground">Unused</span>
    )
  }

  return (
    <details className="group">
      <summary className="cursor-pointer list-none text-xs font-semibold text-primary marker:hidden hover:underline">
        Used by {policies.length} route{policies.length === 1 ? '' : 's'}
      </summary>
      <ul className="mt-2 grid gap-1 pl-1">
        {policies.map((policy) => (
          <li key={policy.policyId}>
            <Link
              className="text-xs text-primary hover:underline"
              href={`/policies/${policy.policyId}`}
            >
              {policy.name}
            </Link>
          </li>
        ))}
      </ul>
    </details>
  )
}
