export type MonitorType = 'http' | 'tcp' | 'grpc' | 'dns'
export type ServiceLifecycle = 'draft' | 'active' | 'archived'

export const TECHNOLOGY_KEYS = [
  'golang',
  'mariadb',
  'mysql',
  'nginx',
  'postgres',
  'python',
  'typescript',
  'mongodb',
  'redis',
  'kafka',
  'docker',
  'apache',
  'javascript',
  'rabbitmq',
] as const
export type TechnologyKey = (typeof TECHNOLOGY_KEYS)[number]

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
  lastProbeLocationId: string
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
  probeLocations: string[]
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
  probeLocations: string[]
  currentStatus?: string
  lastCheckedAt?: string
  lastDurationMs?: number
  lastProbeLocationId?: string
  lastError?: string
  updatedAt?: string
}

export interface Service {
  tenantId: string
  serviceId: string
  name: string
  description?: string
  lifecycleState: ServiceLifecycle
  technologyKey?: TechnologyKey
  monitorCount?: number
  enabledMonitorCount?: number
  rollupStatus?: string
  escalationPolicyId?: string
  businessHours?: BusinessHoursConfig
  createdAt?: string
  updatedAt?: string
  monitors?: Monitor[]
}

export interface CheckRun {
  runId: string
  type: string
  probeLocationId: string
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

export interface MonitorRunsResponse {
  runs: CheckRun[]
}

export interface CreateMonitorPayload {
  name: string
  type: MonitorType
  intervalSeconds: number
  probeLocations: string[]
  enabled: boolean
  http: HttpConfiguration
}

export interface UpdateMonitorPayload {
  name?: string
  intervalSeconds?: number
  probeLocations?: string[]
  http?: HttpConfiguration
}

export interface CreateServicePayload {
  name: string
  description?: string
  lifecycleState?: ServiceLifecycle
  technologyKey?: TechnologyKey
  escalationPolicyId?: string
  businessHours?: BusinessHoursConfig
}

export interface UpdateServicePayload {
  name?: string
  description?: string
  lifecycleState?: ServiceLifecycle
  technologyKey?: TechnologyKey
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

export interface ProbeLocation {
  locationId: string
  displayName: string
  enabled: boolean
}

export interface ProbeLocationListResponse {
  probeLocations: ProbeLocation[]
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
