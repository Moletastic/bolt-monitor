export type MonitorType = 'http' | 'tcp' | 'grpc' | 'dns'
export type ServiceLifecycle = 'draft' | 'active' | 'archived'
export const SERVICE_CATEGORIES = [
  'server',
  'database',
  'cache',
  'http',
  'queue',
  'container',
  'function',
  'web',
  'api',
  'worker',
  'scheduler',
  'storage',
  'search',
  'auth',
  'payments',
  'analytics',
  'observability',
  'ai',
  'integration',
  'media',
  'content',
  'finance',
  'learning',
  'gaming',
  'commerce',
  'messaging',
  'support',
  'marketing',
  'admin',
  'security',
  'location',
  'social',
]

export type ServiceCategory = (typeof SERVICE_CATEGORIES)[number]

export const SERVICE_CATEGORY_LABELS: Record<ServiceCategory, string> = {
  server: 'Server',
  database: 'Database',
  cache: 'Cache',
  http: 'HTTP',
  queue: 'Queue',
  container: 'Container',
  function: 'Function',
  web: 'Web app',
  api: 'API',
  worker: 'Worker',
  scheduler: 'Scheduler',
  storage: 'Storage',
  search: 'Search',
  auth: 'Authentication',
  payments: 'Payments',
  analytics: 'Analytics',
  observability: 'Observability',
  ai: 'AI',
  integration: 'Integration',
  media: 'Media',
  content: 'Content',
  finance: 'Finance',
  learning: 'Learning',
  gaming: 'Gaming',
  commerce: 'Commerce',
  messaging: 'Messaging',
  support: 'Support',
  marketing: 'Marketing',
  admin: 'Admin',
  security: 'Security',
  location: 'Location',
  social: 'Social',
}

export function isServiceCategory(value: string): value is ServiceCategory {
  return (SERVICE_CATEGORIES as readonly string[]).includes(value)
}

export function formatServiceCategoryLabel(category: ServiceCategory) {
  return SERVICE_CATEGORY_LABELS[category]
}

export interface HttpConfiguration {
  target: string
  method: string
  headers?: Record<string, string>
  timeoutMs: number
  expectedStatusCodes?: number[]
  expectedBodyContains?: string
}

export interface MonitorStatus {
  serviceId: string
  monitorId: string
  currentStatus: string
  consecutiveFailures: number
  consecutiveSuccesses: number
  lastCheckedAt: string
  lastDurationMs: number
  lastError?: string
  lastOutcome: string
}

export interface Monitor {
  monitorId: string
  serviceId: string
  tenantId: string
  name: string
  type: MonitorType
  intervalSeconds: number
  enabled: boolean
  failureThreshold: number
  recoveryThreshold: number
  http?: HttpConfiguration
  status?: MonitorStatus
}

export interface MonitorSummary {
  tenantId: string
  serviceId: string
  monitorId: string
  name: string
  type: MonitorType
  enabled: boolean
  intervalSeconds: number
  currentStatus?: string
  lastCheckedAt?: string
  lastDurationMs?: number
  lastError?: string
  updatedAt?: string
}

export interface Service {
  tenantId: string
  serviceId: string
  name: string
  description?: string
  lifecycleState: ServiceLifecycle
  serviceCategory?: ServiceCategory
  monitorCount?: number
  enabledMonitorCount?: number
  rollupStatus?: string
  escalationPolicyId?: string
  businessHours?: BusinessHoursConfig
  createdAt?: string
  updatedAt?: string
  cardMetrics?: ServiceCardMetrics
  monitors?: Monitor[]
}

export type ServiceCardMetricState = 'ready' | 'no_monitors' | 'no_data'

export interface ServiceCardTrendPoint {
  monitorId: string
  startedAt: string
  durationMs: number
  outcome: string
  success: boolean
}

export interface ServiceCardMetrics {
  state: ServiceCardMetricState
  sampleCount: number
  successCount: number
  monitorCount: number
  upMonitorCount: number
  avgLatencyMs?: number
  p99LatencyMs?: number
  recentUptimePct?: number
  trend?: ServiceCardTrendPoint[]
}

export interface CheckRun {
  runId: string
  type: string
  trigger: string
  startedAt: string
  finishedAt: string
  durationMs: number
  outcome: string
  statusCode?: number
  error?: string
}

export interface ListServicesResponse {
  services: Service[]
}

export type GlobalSearchResourceType = 'service' | 'monitor' | 'policy' | 'channel'

export interface GlobalSearchResult {
  type: GlobalSearchResourceType
  id: string
  serviceId?: string
  label: string
  description: string
  href: string
  iconKey: string
  matchText: string
}

export interface GlobalSearchResponse {
  results: GlobalSearchResult[]
}

export interface MonitorRunsResponse {
  runs: CheckRun[]
}

export interface CreateMonitorPayload {
  name: string
  type: MonitorType
  intervalSeconds: number
  enabled: boolean
  http: HttpConfiguration
}

export interface UpdateMonitorPayload {
  name?: string
  intervalSeconds?: number
  http?: HttpConfiguration
}

export interface CreateServicePayload {
  name: string
  description?: string
  lifecycleState?: ServiceLifecycle
  serviceCategory?: ServiceCategory
  escalationPolicyId?: string
  businessHours?: BusinessHoursConfig
}

export interface UpdateServicePayload {
  name?: string
  description?: string
  lifecycleState?: ServiceLifecycle
  serviceCategory?: ServiceCategory
  escalationPolicyId?: string | null
  businessHours?: BusinessHoursConfig | null
}

export interface Incident {
  incidentId: string
  serviceId?: string
  monitorId: string
  type?: string
  summary: string
  status: string
  openedAt: string
  acknowledgedAt?: string
  resolvedAt?: string
  updatedAt: string
  origin?: string
  originalIncidentId?: string
}

export interface IncidentListResponse {
  incidents: Incident[]
}

export interface IncidentActivity {
  activityId: string
  incidentId: string
  action: string
  timestamp: string
}

export interface IncidentActivityListResponse {
  activities: IncidentActivity[]
}

export interface SchedulerConfig {
  recurringEnabled: boolean
  stopControlMode?: string
  updatedAt: string
}

export interface SchedulerConfigResponse {
  recurringEnabled: boolean
  stopControlMode?: string
  updatedAt: string
}

export interface UpdateSchedulerConfigPayload {
  recurringEnabled: boolean
  stopControlMode?: string
}

export interface AuditEvent {
  auditId: string
  eventType: string
  occurredAt: string
  actor?: string
  origin?: string
}

export interface MonitorAuditResponse {
  events: AuditEvent[]
}

export interface ServiceAuditEventsResponse {
  events: AuditEvent[]
}

export interface ManualRunResponse {
  runId: string
  monitorId: string
  trigger: string
  acceptedAt: string
}

export type EscalationChannelType = 'telegram' | 'email' | 'sms' | 'webhook' | 'pagerduty'

export interface NotificationChannel {
  channelId: string
  tenantId?: string
  name: string
  type: EscalationChannelType
  target: string
  config?: Record<string, unknown>
  createdAt?: string
  updatedAt?: string
}

export interface ListNotificationChannelsResponse {
  channels: NotificationChannel[]
}

export interface CreateNotificationChannelPayload {
  name: string
  type: EscalationChannelType
  target: string
  config?: Record<string, unknown>
}

export interface UpdateNotificationChannelPayload {
  name?: string
  type?: EscalationChannelType
  target?: string
  config?: Record<string, unknown>
}

export interface EscalationStep {
  channelId: string
  delayMinutes: number
}

export interface EscalationPath {
  steps: EscalationStep[]
}

export interface BusinessHoursConfig {
  timezone: string
  startHour: number
  endHour: number
  daysOfWeek: number[]
}

export interface EscalationPolicy {
  tenantId?: string
  policyId: string
  name: string
  description?: string
  businessHoursPath: EscalationPath
  offHoursPath: EscalationPath
  createdAt?: string
  updatedAt?: string
}

export interface ListEscalationPoliciesResponse {
  policies: EscalationPolicy[]
}

export interface CreateEscalationPolicyPayload {
  name: string
  description?: string
  businessHoursPath: EscalationPath
  offHoursPath: EscalationPath
}

export interface UpdateEscalationPolicyPayload {
  name: string
  description?: string
  businessHoursPath: EscalationPath
  offHoursPath: EscalationPath
}
