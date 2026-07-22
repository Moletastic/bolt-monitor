package main

import (
	"context"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/domainvalues"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/notifications"
	"bolt-monitor/shared/resultstatus"
)

// Domain test fakes expose only the interface a handler domain consumes.
// They delegate storage behavior to the shared test state used by legacy
// integration tests, avoiding duplicate fixture data while keeping the
// individual fake method sets narrow.
type fakeServiceStore struct{ state *fakeMonitorRepositoryState }

func (f fakeServiceStore) CreateService(c context.Context, v monitorconfig.Service) (monitorconfig.Service, error) {
	return f.state.CreateService(c, v)
}
func (f fakeServiceStore) ListServices(c context.Context, t string) ([]monitorconfig.Service, error) {
	return f.state.ListServices(c, t)
}
func (f fakeServiceStore) GetService(c context.Context, t, id string) (monitorconfig.Service, bool, error) {
	return f.state.GetService(c, t, id)
}
func (f fakeServiceStore) UpdateService(c context.Context, v monitorconfig.Service) (monitorconfig.Service, error) {
	return f.state.UpdateService(c, v)
}
func (f fakeServiceStore) DeleteService(c context.Context, t, id string) (bool, error) {
	return f.state.DeleteService(c, t, id)
}
func (f fakeServiceStore) ArchiveService(c context.Context, t, id string) (monitorconfig.Service, error) {
	return f.state.ArchiveService(c, t, id)
}
func (f fakeServiceStore) ReactivateService(c context.Context, t, id string) (monitorconfig.Service, error) {
	return f.state.ReactivateService(c, t, id)
}
func (f fakeServiceStore) ServiceReferencesEscalationPolicy(c context.Context, t, id string) (bool, error) {
	return f.state.ServiceReferencesEscalationPolicy(c, t, id)
}
func (f fakeServiceStore) GetServiceCardMetrics(c context.Context, t, id string) (serviceCardMetricsResponse, error) {
	return f.state.GetServiceCardMetrics(c, t, id)
}
func (f fakeServiceStore) GetMonitorStatus(c context.Context, t, s, id string) (resultstatus.MonitorStatus, bool, error) {
	return f.state.GetMonitorStatus(c, t, s, id)
}
func (f fakeServiceStore) ListMonitors(c context.Context, t, id string) ([]monitorconfig.Monitor, error) {
	return f.state.ListMonitors(c, t, id)
}
func (f fakeServiceStore) ListServiceAuditEventsPage(c context.Context, t, id string, n int32, k map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	return f.state.ListServiceAuditEventsPage(c, t, id, n, k)
}
func (f fakeServiceStore) ListServiceIncidents(c context.Context, t, id string, n int32) ([]dynamodbrecord.IncidentRecord, error) {
	return f.state.ListServiceIncidents(c, t, id, n)
}

type fakeMonitorStore struct{ state *fakeMonitorRepositoryState }

func (f fakeMonitorStore) CreateMonitor(c context.Context, v monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	return f.state.CreateMonitor(c, v)
}
func (f fakeMonitorStore) ListMonitors(c context.Context, t, s string) ([]monitorconfig.Monitor, error) {
	return f.state.ListMonitors(c, t, s)
}
func (f fakeMonitorStore) GetMonitor(c context.Context, t, s, id string) (monitorconfig.Monitor, bool, error) {
	return f.state.GetMonitor(c, t, s, id)
}
func (f fakeMonitorStore) GetMonitorByRef(c context.Context, ref domainvalues.MonitorRef) (monitorconfig.Monitor, bool, error) {
	return f.state.GetMonitor(c, string(ref.Tenant), string(ref.Service), string(ref.Monitor))
}
func (f fakeMonitorStore) UpdateMonitor(c context.Context, v monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	return f.state.UpdateMonitor(c, v)
}
func (f fakeMonitorStore) DeleteMonitor(c context.Context, t, s, id string) (bool, error) {
	return f.state.DeleteMonitor(c, t, s, id)
}
func (f fakeMonitorStore) SetMonitorEnabled(c context.Context, t, s, id string, e bool) (monitorconfig.Monitor, bool, error) {
	return f.state.SetMonitorEnabled(c, t, s, id, e)
}
func (f fakeMonitorStore) SetMonitorMaintenance(c context.Context, t, s, id string, e bool) (resultstatus.MonitorStatus, bool, error) {
	return f.state.SetMonitorMaintenance(c, t, s, id, e)
}
func (f fakeMonitorStore) GetMonitorStatus(c context.Context, t, s, id string) (resultstatus.MonitorStatus, bool, error) {
	return f.state.GetMonitorStatus(c, t, s, id)
}
func (f fakeMonitorStore) ListMonitorRuns(c context.Context, t, s, id string, n int32) ([]resultstatus.CheckRun, error) {
	return f.state.ListMonitorRuns(c, t, s, id, n)
}
func (f fakeMonitorStore) ListMonitorRunsPage(c context.Context, t, s, id string, n int32, k map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error) {
	return f.state.ListMonitorRunsPage(c, t, s, id, n, k)
}
func (f fakeMonitorStore) GetServiceCardMetrics(c context.Context, t, s string) (serviceCardMetricsResponse, error) {
	return f.state.GetServiceCardMetrics(c, t, s)
}
func (f fakeMonitorStore) CreateManualRun(c context.Context, m monitorconfig.Monitor, n time.Time) (manualRunRequestRecord, error) {
	return f.state.CreateManualRun(c, m, n)
}
func (f fakeMonitorStore) RecordExecutionResult(c context.Context, m monitorconfig.Monitor, id string, r checkexecution.ExecutionResult) error {
	return f.state.RecordExecutionResult(c, m, id, r)
}
func (f fakeMonitorStore) ReserveManualIdempotency(c context.Context, r manualIdempotencyRecord) (manualIdempotencyRecord, error) {
	return f.state.ReserveManualIdempotency(c, r)
}
func (f fakeMonitorStore) LoadManualIdempotency(c context.Context, t, s, id, k string) (manualIdempotencyRecord, bool, error) {
	return f.state.LoadManualIdempotency(c, t, s, id, k)
}
func (f fakeMonitorStore) ListMonitorIncidents(c context.Context, t, s, id string) ([]dynamodbrecord.IncidentRecord, error) {
	return f.state.ListMonitorIncidents(c, t, s, id)
}
func (f fakeMonitorStore) ListMonitorIncidentsPage(c context.Context, t, s, id string, n int32, k map[string]sharedaws.AttributeValue) (historyPage[dynamodbrecord.IncidentRecord], error) {
	return f.state.ListMonitorIncidentsPage(c, t, s, id, n, k)
}
func (f fakeMonitorStore) ListServiceIncidents(c context.Context, t, s string, n int32) ([]dynamodbrecord.IncidentRecord, error) {
	return f.state.ListServiceIncidents(c, t, s, n)
}
func (f fakeMonitorStore) ListMonitorAuditEventsPage(c context.Context, t, s, id string, n int32, k map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	return f.state.ListMonitorAuditEventsPage(c, t, s, id, n, k)
}

type fakeIncidentStore struct{ state *fakeMonitorRepositoryState }

func (f fakeIncidentStore) ListIncidents(c context.Context, t, s string) ([]dynamodbrecord.IncidentRecord, error) {
	return f.state.ListIncidents(c, t, s)
}
func (f fakeIncidentStore) GetIncident(c context.Context, t, id string) (dynamodbrecord.IncidentRecord, bool, error) {
	return f.state.GetIncident(c, t, id)
}
func (f fakeIncidentStore) ListIncidentActivities(c context.Context, t, id string) ([]dynamodbrecord.IncidentActivityRecord, error) {
	return f.state.ListIncidentActivities(c, t, id)
}
func (f fakeIncidentStore) AcknowledgeIncident(c context.Context, t, id string, n time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	return f.state.AcknowledgeIncident(c, t, id, n)
}
func (f fakeIncidentStore) ResolveIncident(c context.Context, t, id string, n time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	return f.state.ResolveIncident(c, t, id, n)
}
func (f fakeIncidentStore) GetEscalationState(c context.Context, t, id string) (*escalation.EscalationState, error) {
	return f.state.GetEscalationState(c, t, id)
}
func (f fakeIncidentStore) ListIncidentDeliveries(c context.Context, t, id string) ([]notifications.DeliveryRecord, error) {
	return f.state.ListIncidentDeliveries(c, t, id)
}
func (f fakeIncidentStore) PrepareDeliveryReplay(c context.Context, command notifications.ReplayCommand, fingerprint string, now time.Time, retention time.Duration) (string, error) {
	return f.state.PrepareDeliveryReplay(c, command, fingerprint, now, retention)
}
func (f fakeIncidentStore) LookupReplayIdempotency(c context.Context, tenantID, incidentID, deliveryID, key string) (*notifications.ReplayIdempotencyRecord, error) {
	return f.state.LookupReplayIdempotency(c, tenantID, incidentID, deliveryID, key)
}

type fakeSchedulerStore struct{ state *fakeMonitorRepositoryState }

func (f fakeSchedulerStore) GetSchedulerConfig(c context.Context, t string) (dynamodbrecord.SchedulerConfigRecord, error) {
	return f.state.GetSchedulerConfig(c, t)
}
func (f fakeSchedulerStore) UpdateSchedulerConfig(c context.Context, t string, v checkexecution.SchedulerConfig, n time.Time) (dynamodbrecord.SchedulerConfigRecord, error) {
	return f.state.UpdateSchedulerConfig(c, t, v, n)
}

type fakeAuditStore struct{ state *fakeMonitorRepositoryState }

func (f fakeAuditStore) ListMonitorAuditEvents(c context.Context, t, s, id string) ([]auditEventView, error) {
	return f.state.ListMonitorAuditEvents(c, t, s, id)
}
func (f fakeAuditStore) ListMonitorAuditEventsPage(c context.Context, t, s, id string, n int32, k map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	return f.state.ListMonitorAuditEventsPage(c, t, s, id, n, k)
}
func (f fakeAuditStore) ListServiceAuditEvents(c context.Context, t, s string) ([]auditEventView, error) {
	return f.state.ListServiceAuditEvents(c, t, s)
}
func (f fakeAuditStore) ListServiceAuditEventsPage(c context.Context, t, s string, n int32, k map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	return f.state.ListServiceAuditEventsPage(c, t, s, n, k)
}

type fakeSearchStore struct{ state *fakeMonitorRepositoryState }

func (f fakeSearchStore) SearchResources(c context.Context, t, q string, n int, types map[string]struct{}) ([]searchResult, error) {
	return f.state.SearchResources(c, t, q, n, types)
}

type fakeEscalationStore struct{ state *fakeMonitorRepositoryState }

func (f fakeEscalationStore) CreateEscalationPolicy(c context.Context, v escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	return f.state.CreateEscalationPolicy(c, v)
}
func (f fakeEscalationStore) ListEscalationPolicies(c context.Context, t string) ([]escalation.EscalationPolicy, error) {
	return f.state.ListEscalationPolicies(c, t)
}
func (f fakeEscalationStore) GetEscalationPolicy(c context.Context, t, id string) (*escalation.EscalationPolicy, error) {
	return f.state.GetEscalationPolicy(c, t, id)
}
func (f fakeEscalationStore) UpdateEscalationPolicy(c context.Context, v escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	return f.state.UpdateEscalationPolicy(c, v)
}
func (f fakeEscalationStore) DeleteEscalationPolicy(c context.Context, t, id string) error {
	return f.state.DeleteEscalationPolicy(c, t, id)
}
func (f fakeEscalationStore) CreateNotificationChannel(c context.Context, v escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	return f.state.CreateNotificationChannel(c, v)
}
func (f fakeEscalationStore) ListNotificationChannels(c context.Context, t string) ([]escalation.NotificationChannel, error) {
	return f.state.ListNotificationChannels(c, t)
}
func (f fakeEscalationStore) GetNotificationChannel(c context.Context, t, id string) (*escalation.NotificationChannel, error) {
	return f.state.GetNotificationChannel(c, t, id)
}
func (f fakeEscalationStore) UpdateNotificationChannel(c context.Context, v escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	return f.state.UpdateNotificationChannel(c, v)
}
func (f fakeEscalationStore) DeleteNotificationChannel(c context.Context, t, id string) error {
	return f.state.DeleteNotificationChannel(c, t, id)
}
func (f fakeEscalationStore) ChannelsReferencedByRoutes(c context.Context, t, id string) ([]routeReference, error) {
	return f.state.ChannelsReferencedByRoutes(c, t, id)
}
func (f fakeEscalationStore) RecordNotificationChannelTestAudit(c context.Context, t, id, typ, outcome, reason string, n time.Time) error {
	return f.state.RecordNotificationChannelTestAudit(c, t, id, typ, outcome, reason, n)
}
func (f fakeEscalationStore) GetEscalationState(c context.Context, t, id string) (*escalation.EscalationState, error) {
	return f.state.GetEscalationState(c, t, id)
}

var (
	_ MonitorStore    = fakeMonitorStore{}
	_ SchedulerStore  = fakeSchedulerStore{}
	_ AuditStore      = fakeAuditStore{}
	_ EscalationStore = fakeEscalationStore{}
	_ SearchStore     = fakeSearchStore{}
)
