import {
  type CreateServicePayload,
  type CreateMonitorPayload,
  type ListServicesResponse,
  type Monitor,
  type MonitorRunsResponse,
  type MonitorStatus,
  type Service,
  type UpdateServicePayload,
  type UpdateMonitorPayload,
  type Incident,
  type IncidentActivityListResponse,
  type IncidentListResponse,
  type SchedulerConfigResponse,
  type ProbeLocationListResponse,
  type MonitorAuditResponse,
  type ServiceAuditEventsResponse,
  type ManualRunResponse,
  type EscalationPolicy,
  type ListEscalationPoliciesResponse,
  type CreateEscalationPolicyPayload,
  type UpdateEscalationPolicyPayload,
  type ListNotificationChannelsResponse,
  type NotificationChannel,
  type CreateNotificationChannelPayload,
  type UpdateNotificationChannelPayload,
} from '@/lib/types'
import { type ApiResponse, isError, Status } from '@/lib/api-response'
import { ApiError, ApiErrorCode, fromEnvelope, type ApiReasonPayload } from '@/lib/errors'

export { ApiError }
export type { ApiErrorCode, ApiReasonPayload }

function getApiBaseUrl() {
  const baseUrl = process.env.NEXT_PUBLIC_MONITOR_API_BASE_URL
  if (!baseUrl) {
    throw new Error('NEXT_PUBLIC_MONITOR_API_BASE_URL is not configured.')
  }

  return baseUrl.replace(/\/$/, '')
}

function unwrap<T>(response: ApiResponse<T>, httpStatus: number): T {
  if (isError(response)) {
    throw attachMessage(fromEnvelope(response.reason, httpStatus), response.message)
  }
  if (response.status !== Status.Success || response.data === undefined) {
    throw new ApiError(ApiErrorCode.Internal, httpStatus, { body: response })
  }
  return response.data
}

function attachMessage(err: ApiError, message: string | undefined): ApiError {
  if (!message) return err
  return new ApiError(err.code, err.status, err.details, message)
}

async function apiRequest<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    ...init,
    cache: 'no-store',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  const text = await response.text()
  const parsed = text ? (JSON.parse(text) as ApiResponse<T>) : undefined

  if (!response.ok) {
    const reason = parsed && isError(parsed) ? parsed.reason : undefined
    if (reason) {
      throw attachMessage(fromEnvelope(reason, response.status), parsed?.message)
    }
    throw new ApiError(ApiErrorCode.Internal, response.status, {}, parsed?.message)
  }

  return unwrap(parsed as ApiResponse<T>, response.status)
}

async function apiRequestVoid(path: string, init?: RequestInit): Promise<void> {
  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    ...init,
    cache: 'no-store',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })

  const text = await response.text()
  const parsed = text ? (JSON.parse(text) as ApiResponse<unknown>) : undefined

  if (!response.ok) {
    const reason = parsed && isError(parsed) ? parsed.reason : undefined
    if (reason) {
      throw attachMessage(fromEnvelope(reason, response.status), parsed?.message)
    }
    throw new ApiError(ApiErrorCode.Internal, response.status, {}, parsed?.message)
  }
}

export async function listServices() {
  const response = await apiRequest<ListServicesResponse>('/api/v1/services')
  return response.services
}

export async function getService(serviceId: string) {
  return apiRequest<Service>(`/api/v1/services/${serviceId}`)
}

export async function createService(payload: CreateServicePayload) {
  return apiRequest<Service>('/api/v1/services', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function updateService(serviceId: string, payload: UpdateServicePayload) {
  return apiRequest<Service>(`/api/v1/services/${serviceId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export async function deleteService(serviceId: string) {
  await apiRequestVoid(`/api/v1/services/${serviceId}`, {
    method: 'DELETE',
  })
}

export async function getMonitor(serviceId: string, monitorId: string) {
  return apiRequest<Monitor>(`/api/v1/services/${serviceId}/monitors/${monitorId}`)
}

export async function getMonitorStatus(serviceId: string, monitorId: string) {
  return apiRequest<MonitorStatus>(`/api/v1/services/${serviceId}/monitors/${monitorId}/status`)
}

export async function listMonitorRuns(serviceId: string, monitorId: string) {
  const response = await apiRequest<MonitorRunsResponse>(
    `/api/v1/services/${serviceId}/monitors/${monitorId}/runs`
  )
  return response.runs
}

export async function getMonitorRuns(serviceId: string, monitorId: string) {
  return listMonitorRuns(serviceId, monitorId)
}

export async function createMonitor(serviceId: string, payload: CreateMonitorPayload) {
  return apiRequest<Monitor>(`/api/v1/services/${serviceId}/monitors`, {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function updateMonitor(
  serviceId: string,
  monitorId: string,
  payload: UpdateMonitorPayload
) {
  return apiRequest<Monitor>(`/api/v1/services/${serviceId}/monitors/${monitorId}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export async function deleteMonitor(serviceId: string, monitorId: string) {
  await apiRequestVoid(`/api/v1/services/${serviceId}/monitors/${monitorId}`, {
    method: 'DELETE',
  })
}

export async function setMonitorEnabled(serviceId: string, monitorId: string, enabled: boolean) {
  return apiRequest<Monitor>(
    `/api/v1/services/${serviceId}/monitors/${monitorId}/${enabled ? 'enable' : 'disable'}`,
    {
      method: 'POST',
      body: '{}',
    }
  )
}

export async function setMonitorMaintenance(
  serviceId: string,
  monitorId: string,
  enabled: boolean
) {
  return apiRequest<MonitorStatus>(
    `/api/v1/services/${serviceId}/monitors/${monitorId}/maintenance/${enabled ? 'enable' : 'disable'}`,
    {
      method: 'POST',
      body: '{}',
    }
  )
}

export async function listIncidents(status?: string) {
  const path = status ? `/api/v1/incidents?status=${status}` : '/api/v1/incidents'
  const response = await apiRequest<IncidentListResponse>(path)
  return response.incidents
}

export async function getIncident(incidentId: string) {
  return apiRequest<Incident>(`/api/v1/incidents/${incidentId}`)
}

export async function getIncidentActivities(incidentId: string) {
  const response = await apiRequest<IncidentActivityListResponse>(
    `/api/v1/incidents/${incidentId}/activities`
  )
  return response.activities
}

export async function acknowledgeIncident(incidentId: string) {
  return apiRequest<Incident>(`/api/v1/incidents/${incidentId}/ack`, { method: 'POST' })
}

export async function resolveIncident(incidentId: string) {
  return apiRequest<Incident>(`/api/v1/incidents/${incidentId}/resolve`, { method: 'POST' })
}

export async function getSchedulerConfig() {
  return apiRequest<SchedulerConfigResponse>('/api/v1/admin/scheduler-config')
}

export async function updateSchedulerConfig(payload: {
  recurringEnabled: boolean
  stopControlMode?: string
}) {
  return apiRequest<SchedulerConfigResponse>('/api/v1/admin/scheduler-config', {
    method: 'PATCH',
    body: JSON.stringify(payload),
  })
}

export async function listProbeLocations() {
  const response = await apiRequest<ProbeLocationListResponse>('/api/v1/probe-locations')
  return response.probeLocations
}

export async function getMonitorIncidents(serviceId: string, monitorId: string) {
  const response = await apiRequest<IncidentListResponse>(
    `/api/v1/services/${serviceId}/monitors/${monitorId}/incidents`
  )
  return response.incidents
}

export async function listServiceIncidents(serviceId: string, limit?: number) {
  const path =
    limit === undefined
      ? `/api/v1/services/${serviceId}/incidents`
      : `/api/v1/services/${serviceId}/incidents?limit=${limit}`
  const response = await apiRequest<IncidentListResponse>(path)
  return response.incidents
}

export async function listMonitorAuditEvents(serviceId: string, monitorId: string) {
  const response = await apiRequest<MonitorAuditResponse>(
    `/api/v1/services/${serviceId}/monitors/${monitorId}/audit`
  )
  return response.events
}

export async function getMonitorAudit(serviceId: string, monitorId: string) {
  return listMonitorAuditEvents(serviceId, monitorId)
}

export async function listServiceAuditEvents(serviceId: string) {
  const response = await apiRequest<ServiceAuditEventsResponse>(
    `/api/v1/services/${serviceId}/audit`
  )
  return response.events
}

export async function triggerManualRun(serviceId: string, monitorId: string) {
  return apiRequest<ManualRunResponse>(`/api/v1/services/${serviceId}/monitors/${monitorId}/run`, {
    method: 'POST',
  })
}

export async function listEscalationPolicies() {
  const response = await apiRequest<ListEscalationPoliciesResponse>('/api/v1/escalation-policies')
  return response.policies
}

export async function listNotificationChannels() {
  const response = await apiRequest<ListNotificationChannelsResponse>(
    '/api/v1/notification-channels'
  )
  return response.channels
}

export async function getNotificationChannel(channelId: string) {
  return apiRequest<NotificationChannel>(`/api/v1/notification-channels/${channelId}`)
}

export async function createNotificationChannel(payload: CreateNotificationChannelPayload) {
  return apiRequest<NotificationChannel>('/api/v1/notification-channels', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function updateNotificationChannel(
  channelId: string,
  payload: UpdateNotificationChannelPayload
) {
  return apiRequest<NotificationChannel>(`/api/v1/notification-channels/${channelId}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export async function deleteNotificationChannel(channelId: string) {
  await apiRequestVoid(`/api/v1/notification-channels/${channelId}`, { method: 'DELETE' })
}

export async function testNotificationChannel(channelId: string) {
  return apiRequest<{ channelId: string; sentAt: string }>(
    `/api/v1/notification-channels/${channelId}/test`,
    { method: 'POST' }
  )
}

export async function getEscalationPolicy(policyId: string) {
  return apiRequest<EscalationPolicy>(`/api/v1/escalation-policies/${policyId}`)
}

export async function createEscalationPolicy(payload: CreateEscalationPolicyPayload) {
  return apiRequest<EscalationPolicy>('/api/v1/escalation-policies', {
    method: 'POST',
    body: JSON.stringify(payload),
  })
}

export async function updateEscalationPolicy(
  policyId: string,
  payload: UpdateEscalationPolicyPayload
) {
  return apiRequest<EscalationPolicy>(`/api/v1/escalation-policies/${policyId}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  })
}

export async function deleteEscalationPolicy(policyId: string) {
  await apiRequestVoid(`/api/v1/escalation-policies/${policyId}`, {
    method: 'DELETE',
  })
}

export async function getServiceEscalationPolicy(serviceId: string) {
  return apiRequest<EscalationPolicy>(`/api/v1/services/${serviceId}/escalation-policy`)
}

export type EscalationState = {
  exists: boolean
  policyId?: string
  serviceId?: string
  monitorId?: string
  currentStep?: number
  stepsFired?: number[]
  selectedPath?: string
  scheduledFor?: string
  status?: 'ACTIVE' | 'SUPPRESSED' | 'EXHAUSTED'
  createdAt?: string
  updatedAt?: string
}

export async function getIncidentEscalationState(incidentId: string) {
  return apiRequest<EscalationState>(`/api/v1/incidents/${incidentId}/escalation-state`)
}
