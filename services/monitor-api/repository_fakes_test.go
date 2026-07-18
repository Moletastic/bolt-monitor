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
	"bolt-monitor/shared/resultstatus"
)

// Domain test fakes expose only the interface a handler domain consumes.
// They delegate storage behavior to the shared test state used by legacy
// integration tests, avoiding duplicate fixture data while keeping the
// individual fake method sets narrow.
type fakeServiceStore struct{ state *fakeMonitorRepositoryState }

func (f fakeServiceStore) CreateService(ctx context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	return f.state.CreateService(ctx, service)
}
func (f fakeServiceStore) ListServices(ctx context.Context, tenantID string) ([]monitorconfig.Service, error) {
	return f.state.ListServices(ctx, tenantID)
}
func (f fakeServiceStore) GetService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, bool, error) {
	return f.state.GetService(ctx, tenantID, serviceID)
}
func (f fakeServiceStore) UpdateService(ctx context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	return f.state.UpdateService(ctx, service)
}
func (f fakeServiceStore) DeleteService(ctx context.Context, tenantID, serviceID string) (bool, error) {
	return f.state.DeleteService(ctx, tenantID, serviceID)
}
func (f fakeServiceStore) ArchiveService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	return f.state.ArchiveService(ctx, tenantID, serviceID)
}
func (f fakeServiceStore) ReactivateService(ctx context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	return f.state.ReactivateService(ctx, tenantID, serviceID)
}
func (f fakeServiceStore) ServiceReferencesEscalationPolicy(ctx context.Context, tenantID, policyID string) (bool, error) {
	return f.state.ServiceReferencesEscalationPolicy(ctx, tenantID, policyID)
}

type fakeMonitorStore struct{ state *fakeMonitorRepositoryState }

func (f fakeMonitorStore) CreateMonitor(ctx context.Context, monitor monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	return f.state.CreateMonitor(ctx, monitor)
}
func (f fakeMonitorStore) ListMonitors(ctx context.Context, tenantID, serviceID string) ([]monitorconfig.Monitor, error) {
	return f.state.ListMonitors(ctx, tenantID, serviceID)
}
func (f fakeMonitorStore) GetMonitor(ctx context.Context, tenantID, serviceID, monitorID string) (monitorconfig.Monitor, bool, error) {
	return f.state.GetMonitor(ctx, tenantID, serviceID, monitorID)
}
func (f fakeMonitorStore) GetMonitorByRef(ctx context.Context, ref domainvalues.MonitorRef) (monitorconfig.Monitor, bool, error) {
	return f.state.GetMonitor(ctx, string(ref.Tenant), string(ref.Service), string(ref.Monitor))
}
func (f fakeMonitorStore) UpdateMonitor(ctx context.Context, monitor monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	return f.state.UpdateMonitor(ctx, monitor)
}
func (f fakeMonitorStore) DeleteMonitor(ctx context.Context, tenantID, serviceID, monitorID string) (bool, error) {
	return f.state.DeleteMonitor(ctx, tenantID, serviceID, monitorID)
}
func (f fakeMonitorStore) SetMonitorEnabled(ctx context.Context, tenantID, serviceID, monitorID string, enabled bool) (monitorconfig.Monitor, bool, error) {
	return f.state.SetMonitorEnabled(ctx, tenantID, serviceID, monitorID, enabled)
}
func (f fakeMonitorStore) SetMonitorMaintenance(ctx context.Context, tenantID, serviceID, monitorID string, enabled bool) (resultstatus.MonitorStatus, bool, error) {
	return f.state.SetMonitorMaintenance(ctx, tenantID, serviceID, monitorID, enabled)
}
func (f fakeMonitorStore) GetMonitorStatus(ctx context.Context, tenantID, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	return f.state.GetMonitorStatus(ctx, tenantID, serviceID, monitorID)
}
func (f fakeMonitorStore) ListMonitorRuns(ctx context.Context, tenantID, serviceID, monitorID string, limit int32) ([]resultstatus.CheckRun, error) {
	return f.state.ListMonitorRuns(ctx, tenantID, serviceID, monitorID, limit)
}
func (f fakeMonitorStore) ListMonitorRunsPage(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, cursor map[string]sharedaws.AttributeValue) (historyPage[resultstatus.CheckRun], error) {
	return f.state.ListMonitorRunsPage(ctx, tenantID, serviceID, monitorID, limit, cursor)
}
func (f fakeMonitorStore) GetServiceCardMetrics(ctx context.Context, tenantID, serviceID string) (serviceCardMetricsResponse, error) {
	return f.state.GetServiceCardMetrics(ctx, tenantID, serviceID)
}
func (f fakeMonitorStore) CreateManualRun(ctx context.Context, monitor monitorconfig.Monitor, now time.Time) (manualRunRequestRecord, error) {
	return f.state.CreateManualRun(ctx, monitor, now)
}
func (f fakeMonitorStore) RecordExecutionResult(ctx context.Context, monitor monitorconfig.Monitor, runID string, result checkexecution.ExecutionResult) error {
	return f.state.RecordExecutionResult(ctx, monitor, runID, result)
}
func (f fakeMonitorStore) ListMonitorIncidents(ctx context.Context, tenantID, serviceID, monitorID string) ([]dynamodbrecord.IncidentRecord, error) {
	return f.state.ListMonitorIncidents(ctx, tenantID, serviceID, monitorID)
}
func (f fakeMonitorStore) ListMonitorIncidentsPage(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, cursor map[string]sharedaws.AttributeValue) (historyPage[dynamodbrecord.IncidentRecord], error) {
	return f.state.ListMonitorIncidentsPage(ctx, tenantID, serviceID, monitorID, limit, cursor)
}
func (f fakeMonitorStore) ListServiceIncidents(ctx context.Context, tenantID, serviceID string, limit int32) ([]dynamodbrecord.IncidentRecord, error) {
	return f.state.ListServiceIncidents(ctx, tenantID, serviceID, limit)
}

type fakeIncidentStore struct{ state *fakeMonitorRepositoryState }

func (f fakeIncidentStore) ListIncidents(ctx context.Context, tenantID, status string) ([]dynamodbrecord.IncidentRecord, error) {
	return f.state.ListIncidents(ctx, tenantID, status)
}
func (f fakeIncidentStore) GetIncident(ctx context.Context, tenantID, incidentID string) (dynamodbrecord.IncidentRecord, bool, error) {
	return f.state.GetIncident(ctx, tenantID, incidentID)
}
func (f fakeIncidentStore) ListIncidentActivities(ctx context.Context, tenantID, incidentID string) ([]dynamodbrecord.IncidentActivityRecord, error) {
	return f.state.ListIncidentActivities(ctx, tenantID, incidentID)
}
func (f fakeIncidentStore) AcknowledgeIncident(ctx context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	return f.state.AcknowledgeIncident(ctx, tenantID, incidentID, now)
}
func (f fakeIncidentStore) ResolveIncident(ctx context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	return f.state.ResolveIncident(ctx, tenantID, incidentID, now)
}

type fakeSchedulerStore struct{ state *fakeMonitorRepositoryState }

func (f fakeSchedulerStore) GetSchedulerConfig(ctx context.Context, tenantID string) (dynamodbrecord.SchedulerConfigRecord, error) {
	return f.state.GetSchedulerConfig(ctx, tenantID)
}
func (f fakeSchedulerStore) UpdateSchedulerConfig(ctx context.Context, tenantID string, config checkexecution.SchedulerConfig, now time.Time) (dynamodbrecord.SchedulerConfigRecord, error) {
	return f.state.UpdateSchedulerConfig(ctx, tenantID, config, now)
}

type fakeAuditStore struct{ state *fakeMonitorRepositoryState }

func (f fakeAuditStore) ListMonitorAuditEvents(ctx context.Context, tenantID, serviceID, monitorID string) ([]auditEventView, error) {
	return f.state.ListMonitorAuditEvents(ctx, tenantID, serviceID, monitorID)
}
func (f fakeAuditStore) ListMonitorAuditEventsPage(ctx context.Context, tenantID, serviceID, monitorID string, limit int32, cursor map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	return f.state.ListMonitorAuditEventsPage(ctx, tenantID, serviceID, monitorID, limit, cursor)
}
func (f fakeAuditStore) ListServiceAuditEvents(ctx context.Context, tenantID, serviceID string) ([]auditEventView, error) {
	return f.state.ListServiceAuditEvents(ctx, tenantID, serviceID)
}
func (f fakeAuditStore) ListServiceAuditEventsPage(ctx context.Context, tenantID, serviceID string, limit int32, cursor map[string]sharedaws.AttributeValue) (historyPage[auditEventView], error) {
	return f.state.ListServiceAuditEventsPage(ctx, tenantID, serviceID, limit, cursor)
}

type fakeSearchStore struct{ state *fakeMonitorRepositoryState }

func (f fakeSearchStore) SearchResources(ctx context.Context, tenantID, query string, limit int, types map[string]struct{}) ([]searchResult, error) {
	return f.state.SearchResources(ctx, tenantID, query, limit, types)
}

type fakeEscalationStore struct{ state *fakeMonitorRepositoryState }

func (f fakeEscalationStore) CreateEscalationPolicy(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	return f.state.CreateEscalationPolicy(ctx, policy)
}
func (f fakeEscalationStore) ListEscalationPolicies(ctx context.Context, tenantID string) ([]escalation.EscalationPolicy, error) {
	return f.state.ListEscalationPolicies(ctx, tenantID)
}
func (f fakeEscalationStore) GetEscalationPolicy(ctx context.Context, tenantID, policyID string) (*escalation.EscalationPolicy, error) {
	return f.state.GetEscalationPolicy(ctx, tenantID, policyID)
}
func (f fakeEscalationStore) UpdateEscalationPolicy(ctx context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	return f.state.UpdateEscalationPolicy(ctx, policy)
}
func (f fakeEscalationStore) DeleteEscalationPolicy(ctx context.Context, tenantID, policyID string) error {
	return f.state.DeleteEscalationPolicy(ctx, tenantID, policyID)
}
func (f fakeEscalationStore) CreateNotificationChannel(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	return f.state.CreateNotificationChannel(ctx, channel)
}
func (f fakeEscalationStore) ListNotificationChannels(ctx context.Context, tenantID string) ([]escalation.NotificationChannel, error) {
	return f.state.ListNotificationChannels(ctx, tenantID)
}
func (f fakeEscalationStore) GetNotificationChannel(ctx context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	return f.state.GetNotificationChannel(ctx, tenantID, channelID)
}
func (f fakeEscalationStore) UpdateNotificationChannel(ctx context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	return f.state.UpdateNotificationChannel(ctx, channel)
}
func (f fakeEscalationStore) DeleteNotificationChannel(ctx context.Context, tenantID, channelID string) error {
	return f.state.DeleteNotificationChannel(ctx, tenantID, channelID)
}
func (f fakeEscalationStore) ChannelsReferencedByRoutes(ctx context.Context, tenantID, channelID string) ([]routeReference, error) {
	return f.state.ChannelsReferencedByRoutes(ctx, tenantID, channelID)
}
func (f fakeEscalationStore) RecordNotificationChannelTestAudit(ctx context.Context, tenantID, channelID, channelType, outcome, reason string, now time.Time) error {
	return f.state.RecordNotificationChannelTestAudit(ctx, tenantID, channelID, channelType, outcome, reason, now)
}
func (f fakeEscalationStore) GetEscalationState(ctx context.Context, tenantID, incidentID string) (*escalation.EscalationState, error) {
	return f.state.GetEscalationState(ctx, tenantID, incidentID)
}

// Compile-time contracts prevent a domain fake from accumulating unrelated
// methods to satisfy a broader handler surface.
var (
	_ ServiceStore    = fakeServiceStore{}
	_ MonitorStore    = fakeMonitorStore{}
	_ IncidentStore   = fakeIncidentStore{}
	_ SchedulerStore  = fakeSchedulerStore{}
	_ AuditStore      = fakeAuditStore{}
	_ EscalationStore = fakeEscalationStore{}
	_ SearchStore     = fakeSearchStore{}
)
