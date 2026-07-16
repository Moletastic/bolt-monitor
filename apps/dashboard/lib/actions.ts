'use server'

import { revalidatePath } from 'next/cache'
import { redirect } from 'next/navigation'

import {
  acknowledgeIncident,
  createService,
  createMonitor,
  createEscalationPolicy,
  updateEscalationPolicy,
  deleteEscalationPolicy,
  deleteService,
  deleteMonitor,
  resolveIncident,
  setMonitorEnabled,
  setMonitorMaintenance,
  triggerManualRun,
  updateService,
  updateMonitor,
  updateSchedulerConfig,
  createNotificationChannel,
  updateNotificationChannel,
  deleteNotificationChannel,
  testNotificationChannel,
  getMonitorRuns,
  getMonitorIncidents,
  getMonitorAudit,
  searchResources,
} from '@/lib/api'
import { parseJson, runServerAction } from '@/lib/io/server-action'
import { err, isErr, ok, type Result } from '@/lib/result'
import { actionErr, actionOk, type ActionState } from '@/lib/action-state'
import { ApiErrorCode, messageFor } from '@/lib/errors'
import { requireDashboardSession } from '@/lib/auth/session-guard'
import { requireDashboardCsrf } from '@/lib/auth/csrf'
import type {
  CreateServicePayload,
  CreateMonitorPayload,
  EscalationPath,
  EscalationStep,
  EscalationChannelType,
  BusinessHoursConfig,
  HttpConfiguration,
  CheckRun,
  Incident,
  AuditEvent,
  MonitorHistoryPage,
  GlobalSearchResourceType,
  GlobalSearchResult,
  UpdateServicePayload,
  UpdateMonitorPayload,
} from '@/lib/types'

export type MonitorHistoryKind = 'runs' | 'incidents' | 'audit'
export type MonitorHistoryActionData = {
  kind: MonitorHistoryKind
  page: MonitorHistoryPage<CheckRun | Incident | AuditEvent>
}

export type GlobalSearchActionResult = Result<
  GlobalSearchResult[],
  { readonly code: ApiErrorCode; readonly message: string }
>

/** Keeps interactive global search behind the same server-only API token boundary. */
export async function searchResourcesAction(input: {
  readonly query: string
  readonly limit?: number
  readonly types?: GlobalSearchResourceType[]
}): Promise<GlobalSearchActionResult> {
  await requireDashboardSession()
  const result = await runServerAction(() => searchResources(input))
  return result.ok ? result : err({ code: result.error.code, message: messageFor(result.error) })
}

export async function loadMonitorHistoryPageAction(
  _previousState: ActionState<MonitorHistoryActionData>,
  formData: FormData
): Promise<ActionState<MonitorHistoryActionData>> {
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const monitorId = String(formData.get('monitorId') ?? '').trim()
  const cursor = String(formData.get('cursor') ?? '').trim() || undefined
  const kind = String(formData.get('kind') ?? '') as MonitorHistoryKind

  const result = await runServerAction(async () => {
    switch (kind) {
      case 'runs':
        return { kind, page: await getMonitorRuns(serviceId, monitorId, cursor) }
      case 'incidents':
        return { kind, page: await getMonitorIncidents(serviceId, monitorId, cursor) }
      case 'audit':
        return { kind, page: await getMonitorAudit(serviceId, monitorId, cursor) }
      default:
        throw new Error('Unknown monitor history kind.')
    }
  })
  return isErr(result)
    ? actionErr<MonitorHistoryActionData>(result.error)
    : actionOk<MonitorHistoryActionData>(result.value)
}

function parseHeaders(raw: string): Record<string, string> {
  const headers: Record<string, string> = {}
  for (const line of raw.split('\n')) {
    const trimmed = line.trim()
    if (!trimmed) {
      continue
    }
    const separator = trimmed.indexOf(':')
    if (separator === -1) {
      throw new Error(`Invalid header format: ${trimmed}`)
    }
    const key = trimmed.slice(0, separator).trim()
    const value = trimmed.slice(separator + 1).trim()
    if (!key || !value) {
      throw new Error(`Invalid header format: ${trimmed}`)
    }
    headers[key] = value
  }
  return headers
}

function parseStatusCodes(raw: string): number[] {
  const values = raw
    .split(',')
    .map((part) => part.trim())
    .filter(Boolean)
    .map((part) => Number(part))

  if (values.some((value) => Number.isNaN(value))) {
    throw new Error('Expected status codes must be comma-separated numbers.')
  }

  return values
}

function parsePath(raw: string): Result<EscalationPath, string> {
  const parsedResult = parseJson<EscalationPath>(raw)
  if (isErr(parsedResult)) {
    return parsedResult
  }
  const parsed = parsedResult.value
  if (!parsed || !Array.isArray(parsed.steps)) {
    return err('Escalation path payload was malformed.')
  }
  const steps: EscalationStep[] = []
  for (let index = 0; index < parsed.steps.length; index += 1) {
    const step = parsed.steps[index] as EscalationStep
    const legacyStep = step as EscalationStep & {
      target?: unknown
      config?: unknown
      channels?: unknown[]
    }
    if (
      'target' in legacyStep ||
      'config' in legacyStep ||
      (legacyStep.channels?.length ?? 0) > 0
    ) {
      return err('steps must reference channels by channelId; remove target and config')
    }
    if (!step.channelId) {
      return err(`Pick a channel for step ${index + 1}`)
    }
    if (step.delayMinutes < 0) {
      return err('Step delay cannot be negative.')
    }
    steps.push({ channelId: step.channelId, delayMinutes: step.delayMinutes })
  }
  return ok({ steps })
}

function parseBusinessHours(raw: string): Result<BusinessHoursConfig, string> {
  const parsedResult = parseJson<BusinessHoursConfig>(raw)
  if (isErr(parsedResult)) {
    return parsedResult
  }
  const parsed = parsedResult.value
  if (!parsed || typeof parsed.timezone !== 'string' || !Array.isArray(parsed.daysOfWeek)) {
    return err('Business hours payload was malformed.')
  }
  if (parsed.startHour < 0 || parsed.startHour > 24 || parsed.endHour < 0 || parsed.endHour > 24) {
    return err('Business hours must be between 0 and 24.')
  }
  if (parsed.daysOfWeek.length === 0) {
    return err('Select at least one business day.')
  }
  return ok(parsed)
}

function buildHttpConfiguration(formData: FormData): HttpConfiguration {
  const headers = parseHeaders(String(formData.get('headers') ?? ''))
  const expectedStatusCodes = parseStatusCodes(String(formData.get('expectedStatusCodes') ?? '200'))
  const expectedBodyContains = String(formData.get('expectedBodyContains') ?? '').trim()

  return {
    target: String(formData.get('target') ?? '').trim(),
    method: String(formData.get('method') ?? 'GET')
      .trim()
      .toUpperCase(),
    headers,
    timeoutMs: Number(formData.get('timeoutMs') ?? '5000'),
    expectedStatusCodes,
    expectedBodyContains: expectedBodyContains || undefined,
  }
}

function getReturnTo(formData: FormData, fallback: string) {
  return String(formData.get('returnTo') ?? fallback)
}

function appendError(returnTo: string, message: string): string {
  return `${returnTo}${returnTo.includes('?') ? '&' : '?'}error=${encodeURIComponent(message)}`
}

function appendInlineFeedback(path: string): string {
  return `${path}${path.includes('?') ? '&' : '?'}feedback=inline`
}

export async function createMonitorAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/services/${serviceId}/monitors/new`)

  const payload: CreateMonitorPayload = {
    name: String(formData.get('name') ?? '').trim(),
    type: 'http',
    intervalSeconds: Number(formData.get('intervalSeconds') ?? '60'),
    enabled: formData.get('enabled') === 'on',
    http: buildHttpConfiguration(formData),
  }

  const result = await runServerAction(() => createMonitor(serviceId, payload))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  redirect(
    appendInlineFeedback(`/services/${serviceId}/monitors/${result.value.monitorId}?created=1`)
  )
}

export async function updateMonitorAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const monitorId = String(formData.get('monitorId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/services/${serviceId}/monitors/${monitorId}`)

  const payload: UpdateMonitorPayload = {
    name: String(formData.get('name') ?? '').trim(),
    intervalSeconds: Number(formData.get('intervalSeconds') ?? '60'),
    http: buildHttpConfiguration(formData),
  }

  const result = await runServerAction(() => updateMonitor(serviceId, monitorId, payload))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  revalidatePath(`/services/${serviceId}/monitors/${monitorId}`)
  redirect(appendInlineFeedback(`/services/${serviceId}/monitors/${monitorId}?updated=1`))
}

export async function toggleMonitorAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const monitorId = String(formData.get('monitorId') ?? '').trim()
  const enabled = formData.get('enabled') === 'true'
  const returnTo = getReturnTo(formData, `/services/${serviceId}/monitors/${monitorId}`)

  const result = await runServerAction(() => setMonitorEnabled(serviceId, monitorId, enabled))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  revalidatePath(`/services/${serviceId}/monitors/${monitorId}`)
  redirect(returnTo)
}

export async function toggleMonitorStateAction(
  _previousState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const monitorId = String(formData.get('monitorId') ?? '').trim()
  const enabled = formData.get('enabled') === 'true'

  const result = await runServerAction(() => setMonitorEnabled(serviceId, monitorId, enabled))
  if (isErr(result)) {
    return actionErr(result.error)
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  revalidatePath(`/services/${serviceId}/monitors/${monitorId}`)
  return actionOk(undefined, enabled ? 'Monitor enabled.' : 'Monitor disabled.')
}

export async function toggleMaintenanceModeAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const monitorId = String(formData.get('monitorId') ?? '').trim()
  const enabled = formData.get('enabled') === 'true'
  const returnTo = getReturnTo(formData, `/services/${serviceId}/monitors/${monitorId}`)

  const result = await runServerAction(() => setMonitorMaintenance(serviceId, monitorId, enabled))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  revalidatePath(`/services/${serviceId}/monitors/${monitorId}`)
  redirect(returnTo)
}

export async function deleteMonitorAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const monitorId = String(formData.get('monitorId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/services/${serviceId}/monitors/${monitorId}`)

  const result = await runServerAction(() => deleteMonitor(serviceId, monitorId))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  revalidatePath(`/services/${serviceId}/monitors/${monitorId}`)
  redirect(appendInlineFeedback(`/services/${serviceId}?deletedMonitor=1`))
}

export async function acknowledgeIncidentAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const incidentId = String(formData.get('incidentId') ?? '').trim()
  const returnTo = getReturnTo(formData, '/incidents')

  const result = await runServerAction(() => acknowledgeIncident(incidentId))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/incidents')
  revalidatePath(`/incidents/${incidentId}`)
  redirect(returnTo)
}

export async function acknowledgeIncidentStateAction(
  _previousState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const incidentId = String(formData.get('incidentId') ?? '').trim()

  const result = await runServerAction(() => acknowledgeIncident(incidentId))
  if (isErr(result)) {
    return actionErr(result.error)
  }

  revalidatePath('/incidents')
  revalidatePath(`/incidents/${incidentId}`)
  return actionOk(undefined, 'Incident acknowledged.')
}

export async function resolveIncidentAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const incidentId = String(formData.get('incidentId') ?? '').trim()
  const returnTo = getReturnTo(formData, '/incidents')

  const result = await runServerAction(() => resolveIncident(incidentId))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/incidents')
  revalidatePath(`/incidents/${incidentId}`)
  redirect(returnTo)
}

export async function resolveIncidentStateAction(
  _previousState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const incidentId = String(formData.get('incidentId') ?? '').trim()

  const result = await runServerAction(() => resolveIncident(incidentId))
  if (isErr(result)) {
    return actionErr(result.error)
  }

  revalidatePath('/incidents')
  revalidatePath(`/incidents/${incidentId}`)
  return actionOk(undefined, 'Incident resolved.')
}

export async function updateSchedulerConfigAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const returnTo = getReturnTo(formData, '/admin/scheduler')

  const recurringEnabled = formData.get('recurringEnabled') === 'true'
  const stopControlMode = formData.get('stopControlMode') as string | undefined

  const result = await runServerAction(() =>
    updateSchedulerConfig({ recurringEnabled, stopControlMode })
  )
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/admin/scheduler')
  redirect(`${returnTo}?updated=1`)
}

export async function updateSchedulerConfigStateAction(
  _previousState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const recurringEnabled = formData.get('recurringEnabled') === 'true'
  const stopControlMode = formData.get('stopControlMode') as string | undefined

  const result = await runServerAction(() =>
    updateSchedulerConfig({ recurringEnabled, stopControlMode })
  )
  if (isErr(result)) {
    return actionErr(result.error)
  }

  revalidatePath('/admin/scheduler')
  return actionOk(undefined, 'Scheduler configuration updated.')
}

export async function triggerManualRunAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const monitorId = String(formData.get('monitorId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/services/${serviceId}/monitors/${monitorId}`)

  const result = await runServerAction(() => triggerManualRun(serviceId, monitorId))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath(`/services/${serviceId}`)
  revalidatePath(`/services/${serviceId}/monitors/${monitorId}`)
  redirect(appendInlineFeedback(`${returnTo}?run=triggered`))
}

export async function createServiceAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const returnTo = getReturnTo(formData, '/services/new')

  const serviceCategory = String(formData.get('serviceCategory') ?? '').trim()
  const payload: CreateServicePayload = {
    name: String(formData.get('name') ?? '').trim(),
    description: String(formData.get('description') ?? '').trim() || undefined,
    lifecycleState: String(
      formData.get('lifecycleState') ?? 'draft'
    ).trim() as CreateServicePayload['lifecycleState'],
    serviceCategory: serviceCategory
      ? (serviceCategory as CreateServicePayload['serviceCategory'])
      : undefined,
    escalationPolicyId: stringOrNull(formData.get('escalationPolicyId')) ?? undefined,
    businessHours: parseServiceBusinessHours(formData.get('businessHoursPayload')) ?? undefined,
  }

  const result = await runServerAction(() => createService(payload))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  redirect(appendInlineFeedback(`/services/${result.value.serviceId}?created=1`))
}

export async function updateServiceAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/services/${serviceId}`)

  const serviceCategory = String(formData.get('serviceCategory') ?? '').trim()
  const payload: UpdateServicePayload = {
    name: String(formData.get('name') ?? '').trim(),
    description: String(formData.get('description') ?? '').trim() || undefined,
    lifecycleState: String(
      formData.get('lifecycleState') ?? 'draft'
    ).trim() as UpdateServicePayload['lifecycleState'],
    serviceCategory: serviceCategory
      ? (serviceCategory as UpdateServicePayload['serviceCategory'])
      : undefined,
    escalationPolicyId: stringOrNull(formData.get('escalationPolicyId')),
    businessHours: parseServiceBusinessHours(formData.get('businessHoursPayload')),
  }

  const result = await runServerAction(() => updateService(serviceId, payload))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  redirect(appendInlineFeedback(`/services/${serviceId}?updated=1`))
}

function stringOrNull(value: FormDataEntryValue | null) {
  if (value === null) {
    return null
  }
  const trimmed = String(value).trim()
  return trimmed === '' ? null : trimmed
}

function parseServiceBusinessHours(value: FormDataEntryValue | null): BusinessHoursConfig | null {
  if (value === null) {
    return null
  }
  const raw = String(value).trim()
  if (!raw) {
    return null
  }
  const result = parseBusinessHours(raw)
  return isErr(result) ? null : result.value
}

export async function archiveServiceAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/services/${serviceId}`)

  const result = await runServerAction(() =>
    updateService(serviceId, { lifecycleState: 'archived' })
  )
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  redirect(appendInlineFeedback(`/services/${serviceId}?archived=1`))
}

export async function deleteServiceAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const serviceId = String(formData.get('serviceId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/services/${serviceId}`)

  const result = await runServerAction(() => deleteService(serviceId))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }

  revalidatePath('/services')
  revalidatePath(`/services/${serviceId}`)
  redirect(appendInlineFeedback('/services?deletedService=1'))
}

export async function createEscalationPolicyAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const errorHref = String(formData.get('errorHref') ?? '/policies/new').trim()
  const returnTo = String(formData.get('returnTo') ?? '/policies').trim()
  const name = String(formData.get('name') ?? '').trim()
  const description = String(formData.get('description') ?? '').trim()

  const businessHoursPathResult = parsePath(
    String(formData.get('businessHoursPathPayload') ?? '[]')
  )
  if (isErr(businessHoursPathResult)) {
    redirect(`${errorHref}?error=${encodeURIComponent(businessHoursPathResult.error)}`)
  }
  const businessHoursPath = businessHoursPathResult.value
  const offHoursPathResult = parsePath(String(formData.get('offHoursPathPayload') ?? '[]'))
  if (isErr(offHoursPathResult)) {
    redirect(`${errorHref}?error=${encodeURIComponent(offHoursPathResult.error)}`)
  }
  const offHoursPath = offHoursPathResult.value

  const businessHoursRaw = String(formData.get('businessHoursPayload') ?? '').trim()
  if (businessHoursRaw && process.env.NODE_ENV !== 'production') {
    console.warn(
      '[escalation-policy] businessHoursPayload was submitted on create; escalation policies do not own service-scoped business hours and the value is being ignored.'
    )
  }

  const result = await runServerAction(() =>
    createEscalationPolicy({
      name,
      description: description || undefined,
      businessHoursPath,
      offHoursPath,
    })
  )
  if (isErr(result)) {
    redirect(`${errorHref}?error=${encodeURIComponent(messageFor(result.error))}`)
  }

  revalidatePath('/policies')
  redirect(`${returnTo}?created=1`)
}

export async function updateEscalationPolicyAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const errorHref = String(formData.get('errorHref') ?? '/policies').trim()
  const returnTo = String(formData.get('returnTo') ?? '/policies').trim()
  const policyId = String(formData.get('policyId') ?? '').trim()
  const name = String(formData.get('name') ?? '').trim()
  const description = String(formData.get('description') ?? '').trim()

  const businessHoursPathResult = parsePath(
    String(formData.get('businessHoursPathPayload') ?? '[]')
  )
  if (isErr(businessHoursPathResult)) {
    redirect(`${errorHref}?error=${encodeURIComponent(businessHoursPathResult.error)}`)
  }
  const businessHoursPath = businessHoursPathResult.value
  const offHoursPathResult = parsePath(String(formData.get('offHoursPathPayload') ?? '[]'))
  if (isErr(offHoursPathResult)) {
    redirect(`${errorHref}?error=${encodeURIComponent(offHoursPathResult.error)}`)
  }
  const offHoursPath = offHoursPathResult.value

  const businessHoursRaw = String(formData.get('businessHoursPayload') ?? '').trim()
  if (businessHoursRaw && process.env.NODE_ENV !== 'production') {
    console.warn(
      '[escalation-policy] businessHoursPayload was submitted on update; escalation policies do not own service-scoped business hours and the value is being ignored.'
    )
  }

  const result = await runServerAction(() =>
    updateEscalationPolicy(policyId, {
      name,
      description: description || undefined,
      businessHoursPath,
      offHoursPath,
    })
  )
  if (isErr(result)) {
    redirect(`${errorHref}?error=${encodeURIComponent(messageFor(result.error))}`)
  }

  revalidatePath('/policies')
  revalidatePath(`/policies/${policyId}`)
  redirect(`${returnTo}?updated=1`)
}

export async function deleteEscalationPolicyAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const policyId = String(formData.get('policyId') ?? '').trim()
  const returnTo = String(formData.get('returnTo') ?? '/policies').trim()

  const result = await runServerAction(() => deleteEscalationPolicy(policyId))
  if (isErr(result)) {
    redirect(`${returnTo}?error=${encodeURIComponent(messageFor(result.error))}`)
  }

  revalidatePath('/policies')
  redirect(`${returnTo}?deleted=1`)
}

function buildNotificationChannelPayload(formData: FormData) {
  const type = String(formData.get('type') ?? 'telegram').trim() as EscalationChannelType
  const config: Record<string, string> = {}
  for (const [key, value] of formData.entries()) {
    if (key.startsWith('config.')) {
      const trimmed = String(value).trim()
      if (trimmed && trimmed !== '••••••') {
        config[key.replace('config.', '')] = trimmed
      }
    }
  }
  return {
    name: String(formData.get('name') ?? '').trim(),
    type,
    target: String(formData.get('target') ?? '').trim(),
    config,
  }
}

export async function createNotificationChannelAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const returnTo = getReturnTo(formData, '/integrations/channels/new')
  const result = await runServerAction(() =>
    createNotificationChannel(buildNotificationChannelPayload(formData))
  )
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }
  revalidatePath('/integrations/channels')
  redirect('/integrations/channels?created=1')
}

export async function updateNotificationChannelAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const channelId = String(formData.get('channelId') ?? '').trim()
  const returnTo = getReturnTo(formData, `/integrations/channels/${channelId}`)
  const result = await runServerAction(() =>
    updateNotificationChannel(channelId, buildNotificationChannelPayload(formData))
  )
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }
  revalidatePath('/integrations/channels')
  revalidatePath(`/integrations/channels/${channelId}`)
  redirect(`${returnTo}?updated=1`)
}

export async function deleteNotificationChannelAction(formData: FormData) {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const channelId = String(formData.get('channelId') ?? '').trim()
  const returnTo = getReturnTo(formData, '/integrations/channels')
  const result = await runServerAction(() => deleteNotificationChannel(channelId))
  if (isErr(result)) {
    redirect(appendError(returnTo, messageFor(result.error)))
  }
  revalidatePath('/integrations/channels')
  redirect('/integrations/channels?deleted=1')
}

export async function testNotificationChannelStateAction(
  _previousState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireDashboardCsrf()
  await requireDashboardSession()
  const channelId = String(formData.get('channelId') ?? '').trim()
  const result = await runServerAction(() => testNotificationChannel(channelId))
  if (isErr(result)) {
    if (result.error.code === ApiErrorCode.NotificationDeliveryFailed) {
      const reason = result.error.details.reason
      return {
        status: 'error',
        error: {
          code: result.error.code,
          details: result.error.details,
          message:
            typeof reason === 'string' && reason.trim()
              ? `Test notification failed: ${reason}`
              : 'Test notification could not be delivered.',
        },
      }
    }
    return actionErr(result.error)
  }
  return actionOk(undefined, 'Test notification sent.')
}
