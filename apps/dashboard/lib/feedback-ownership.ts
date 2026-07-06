export const queryFeedbackOwners = {
  created: 'toast',
  updated: 'toast',
  run: 'toast',
  error: 'toast',
  deletedService: 'inline',
  deletedMonitor: 'inline',
  deleted: 'inline',
  archived: 'inline',
} as const

export type QueryFeedbackKey = keyof typeof queryFeedbackOwners
export type FeedbackOwner = (typeof queryFeedbackOwners)[QueryFeedbackKey]

export function feedbackOwnerFor(key: string | null): FeedbackOwner | undefined {
  if (!key) return undefined
  return queryFeedbackOwners[key as QueryFeedbackKey]
}
