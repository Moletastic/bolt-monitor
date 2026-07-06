package main

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"strings"
	"testing"
	"time"

	sharedaws "bolt-monitor/shared/aws"
	"bolt-monitor/shared/checkexecution"
	"bolt-monitor/shared/dynamodbrecord"
	"bolt-monitor/shared/dynamodbschema"
	sharederrors "bolt-monitor/shared/errors"
	"bolt-monitor/shared/escalation"
	"bolt-monitor/shared/monitorconfig"
	"bolt-monitor/shared/resultstatus"
	"github.com/aws/aws-lambda-go/events"
)

type fakeMonitorRepository struct {
	services         map[string]monitorconfig.Service
	monitors         map[string]monitorconfig.Monitor
	statuses         map[string]resultstatus.MonitorStatus
	runs             map[string][]resultstatus.CheckRun
	manual           map[string]manualRunRequestRecord
	policies         map[string]escalation.EscalationPolicy
	channels         map[string]escalation.NotificationChannel
	escalationStates map[string]escalation.EscalationState
	incidents        map[string]dynamodbrecord.IncidentRecord
	activities       map[string][]dynamodbrecord.IncidentActivityRecord
	audit            map[string][]auditEventView
	scheduler        dynamodbrecord.SchedulerConfigRecord
}

type fakeDynamoClient struct {
	transactInput  *sharedaws.DynamoDBTransactWriteItemsInput
	transactInputs []*sharedaws.DynamoDBTransactWriteItemsInput
	getItemOutput  *sharedaws.DynamoDBGetItemOutput
	queryOutput    *sharedaws.DynamoDBQueryOutput
	items          map[string]map[string]sharedaws.AttributeValue
}

func (f *fakeDynamoClient) GetItem(_ context.Context, input *sharedaws.DynamoDBGetItemInput) (*sharedaws.DynamoDBGetItemOutput, error) {
	if f.items != nil {
		pk := input.Key["PK"].(*sharedaws.AttributeValueMemberS).Value
		sk := input.Key["SK"].(*sharedaws.AttributeValueMemberS).Value
		if item, ok := f.items[pk+"|"+sk]; ok {
			return &sharedaws.DynamoDBGetItemOutput{Item: item}, nil
		}
		return &sharedaws.DynamoDBGetItemOutput{}, nil
	}
	if f.getItemOutput != nil {
		return f.getItemOutput, nil
	}
	return &sharedaws.DynamoDBGetItemOutput{}, nil
}

func (f *fakeDynamoClient) Query(_ context.Context, input *sharedaws.DynamoDBQueryInput) (*sharedaws.DynamoDBQueryOutput, error) {
	if f.items != nil {
		pk := input.ExpressionAttributeValues[":pk"].(*sharedaws.AttributeValueMemberS).Value
		prefix := ""
		if value, ok := input.ExpressionAttributeValues[":prefix"]; ok {
			prefix = value.(*sharedaws.AttributeValueMemberS).Value
		}
		items := []map[string]sharedaws.AttributeValue{}
		for _, item := range f.items {
			itemPK, ok1 := item["PK"].(*sharedaws.AttributeValueMemberS)
			itemSK, ok2 := item["SK"].(*sharedaws.AttributeValueMemberS)
			if !ok1 || !ok2 || itemPK.Value != pk {
				continue
			}
			if prefix != "" && !strings.HasPrefix(itemSK.Value, prefix) {
				continue
			}
			items = append(items, item)
		}
		return &sharedaws.DynamoDBQueryOutput{Items: items}, nil
	}
	if f.queryOutput != nil {
		return f.queryOutput, nil
	}
	return &sharedaws.DynamoDBQueryOutput{}, nil
}

func (f *fakeDynamoClient) TransactWriteItems(_ context.Context, input *sharedaws.DynamoDBTransactWriteItemsInput) (*sharedaws.DynamoDBTransactWriteItemsOutput, error) {
	f.transactInput = input
	f.transactInputs = append(f.transactInputs, input)
	return &sharedaws.DynamoDBTransactWriteItemsOutput{}, nil
}

func (f *fakeDynamoClient) PutItem(_ context.Context, input *sharedaws.DynamoDBPutItemInput) (*sharedaws.DynamoDBPutItemOutput, error) {
	if f.items != nil {
		pk := input.Item["PK"].(*sharedaws.AttributeValueMemberS).Value
		sk := input.Item["SK"].(*sharedaws.AttributeValueMemberS).Value
		f.items[pk+"|"+sk] = input.Item
	}
	return &sharedaws.DynamoDBPutItemOutput{}, nil
}

func (f *fakeDynamoClient) UpdateItem(_ context.Context, _ *sharedaws.DynamoDBUpdateItemInput) (*sharedaws.DynamoDBUpdateItemOutput, error) {
	return &sharedaws.DynamoDBUpdateItemOutput{}, nil
}

func (f *fakeDynamoClient) DeleteItem(_ context.Context, input *sharedaws.DynamoDBDeleteItemInput) (*sharedaws.DynamoDBDeleteItemOutput, error) {
	if f.items != nil {
		pk := input.Key["PK"].(*sharedaws.AttributeValueMemberS).Value
		sk := input.Key["SK"].(*sharedaws.AttributeValueMemberS).Value
		delete(f.items, pk+"|"+sk)
	}
	return &sharedaws.DynamoDBDeleteItemOutput{}, nil
}

func (f *fakeDynamoClient) Scan(_ context.Context, input *sharedaws.DynamoDBScanInput) (*sharedaws.DynamoDBScanOutput, error) {
	return &sharedaws.DynamoDBScanOutput{}, nil
}

func newFakeMonitorRepository() *fakeMonitorRepository {
	return &fakeMonitorRepository{
		services:         map[string]monitorconfig.Service{},
		monitors:         map[string]monitorconfig.Monitor{},
		statuses:         map[string]resultstatus.MonitorStatus{},
		runs:             map[string][]resultstatus.CheckRun{},
		manual:           map[string]manualRunRequestRecord{},
		policies:         map[string]escalation.EscalationPolicy{},
		channels:         map[string]escalation.NotificationChannel{},
		escalationStates: map[string]escalation.EscalationState{},
		incidents:        map[string]dynamodbrecord.IncidentRecord{},
		activities:       map[string][]dynamodbrecord.IncidentActivityRecord{},
		audit:            map[string][]auditEventView{},
	}
}

func serviceKey(serviceID string) string            { return serviceID }
func monitorKey(serviceID, monitorID string) string { return serviceID + "/" + monitorID }

func (r *fakeMonitorRepository) CreateService(_ context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	if _, ok := r.services[serviceKey(service.ServiceID)]; ok {
		return monitorconfig.Service{}, errServiceAlreadyExists
	}
	r.services[serviceKey(service.ServiceID)] = service
	return service, nil
}

func (r *fakeMonitorRepository) ListServices(_ context.Context, tenantID string) ([]monitorconfig.Service, error) {
	out := []monitorconfig.Service{}
	for _, service := range r.services {
		if service.TenantID == tenantID {
			out = append(out, service)
		}
	}
	return out, nil
}

func (r *fakeMonitorRepository) GetService(_ context.Context, tenantID, serviceID string) (monitorconfig.Service, bool, error) {
	service, ok := r.services[serviceKey(serviceID)]
	if !ok || service.TenantID != tenantID {
		return monitorconfig.Service{}, false, nil
	}
	return service, true, nil
}

func (r *fakeMonitorRepository) UpdateService(_ context.Context, service monitorconfig.Service) (monitorconfig.Service, error) {
	r.services[serviceKey(service.ServiceID)] = service
	return service, nil
}

func (r *fakeMonitorRepository) DeleteService(_ context.Context, tenantID, serviceID string) (bool, error) {
	service, ok := r.services[serviceKey(serviceID)]
	if !ok || service.TenantID != tenantID {
		return false, nil
	}
	if service.LifecycleState == monitorconfig.ServiceLifecycleActive {
		return true, errCannotDeleteActiveService
	}
	delete(r.services, serviceKey(serviceID))
	for key, monitor := range r.monitors {
		if monitor.ServiceID == serviceID {
			delete(r.monitors, key)
		}
	}
	return true, nil
}

func (r *fakeMonitorRepository) CreateMonitor(_ context.Context, monitor monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	key := monitorKey(monitor.ServiceID, monitor.MonitorID)
	if _, ok := r.monitors[key]; ok {
		return monitorconfig.Monitor{}, errMonitorAlreadyExists
	}
	r.monitors[key] = monitor
	r.statuses[key] = resultstatus.MonitorStatus{ServiceID: monitor.ServiceID, MonitorID: monitor.MonitorID, TenantID: monitor.TenantID, CurrentStatus: "UNKNOWN", LastCheckedAt: time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC), LastOutcome: checkexecution.Outcome(rollupUnknown)}
	service := r.services[serviceKey(monitor.ServiceID)]
	service.MonitorCount++
	if monitor.Enabled {
		service.EnabledCount++
	}
	r.services[serviceKey(monitor.ServiceID)] = service
	return monitor, nil
}

func (r *fakeMonitorRepository) ListMonitors(_ context.Context, tenantID, serviceID string) ([]monitorconfig.Monitor, error) {
	out := []monitorconfig.Monitor{}
	for _, monitor := range r.monitors {
		if monitor.TenantID == tenantID && monitor.ServiceID == serviceID {
			out = append(out, monitor)
		}
	}
	return out, nil
}

func (r *fakeMonitorRepository) GetMonitor(_ context.Context, tenantID, serviceID, monitorID string) (monitorconfig.Monitor, bool, error) {
	monitor, ok := r.monitors[monitorKey(serviceID, monitorID)]
	if !ok || monitor.TenantID != tenantID {
		return monitorconfig.Monitor{}, false, nil
	}
	return monitor, true, nil
}

func (r *fakeMonitorRepository) UpdateMonitor(_ context.Context, monitor monitorconfig.Monitor) (monitorconfig.Monitor, error) {
	r.monitors[monitorKey(monitor.ServiceID, monitor.MonitorID)] = monitor
	return monitor, nil
}

func (r *fakeMonitorRepository) DeleteMonitor(_ context.Context, tenantID, serviceID, monitorID string) (bool, error) {
	service := r.services[serviceKey(serviceID)]
	count := 0
	enabledCount := 0
	for _, monitor := range r.monitors {
		if monitor.ServiceID == serviceID {
			count++
			if monitor.Enabled {
				enabledCount++
			}
		}
	}
	if service.LifecycleState == monitorconfig.ServiceLifecycleActive && count == 1 {
		return true, errCannotDeleteLastMonitorFromActiveService
	}
	key := monitorKey(serviceID, monitorID)
	monitor, ok := r.monitors[key]
	if !ok || monitor.TenantID != tenantID {
		return false, nil
	}
	delete(r.monitors, key)
	delete(r.statuses, key)
	service.MonitorCount = count - 1
	if monitor.Enabled {
		enabledCount--
	}
	service.EnabledCount = enabledCount
	if service.LifecycleState != monitorconfig.ServiceLifecycleArchived {
		if enabledCount > 0 {
			service.LifecycleState = monitorconfig.ServiceLifecycleActive
		} else {
			service.LifecycleState = monitorconfig.ServiceLifecycleDraft
		}
	}
	r.services[serviceKey(serviceID)] = service
	return true, nil
}

func (r *fakeMonitorRepository) SetMonitorEnabled(_ context.Context, tenantID, serviceID, monitorID string, enabled bool) (monitorconfig.Monitor, bool, error) {
	monitor, ok := r.monitors[monitorKey(serviceID, monitorID)]
	if !ok || monitor.TenantID != tenantID {
		return monitorconfig.Monitor{}, false, nil
	}
	monitor.Enabled = enabled
	r.monitors[monitorKey(serviceID, monitorID)] = monitor
	return monitor, true, nil
}

func (r *fakeMonitorRepository) SetMonitorMaintenance(_ context.Context, tenantID, serviceID, monitorID string, enabled bool) (resultstatus.MonitorStatus, bool, error) {
	now := time.Now()
	status := resultstatus.MonitorStatus{
		ServiceID:     strings.ToLower(serviceID),
		MonitorID:     strings.ToLower(monitorID),
		TenantID:      strings.ToUpper(tenantID),
		LastCheckedAt: now,
	}
	if enabled {
		status.CurrentStatus = string(resultstatus.MonitorStateMaintenance)
	} else {
		status.CurrentStatus = string(resultstatus.MonitorStateUp)
	}
	r.statuses[monitorKey(serviceID, monitorID)] = status
	return status, true, nil
}

func (r *fakeMonitorRepository) ArchiveService(_ context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	service, ok := r.services[serviceKey(serviceID)]
	if !ok || service.TenantID != tenantID {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	service.LifecycleState = monitorconfig.ServiceLifecycleArchived
	r.services[serviceKey(serviceID)] = service
	return service, nil
}

func (r *fakeMonitorRepository) ReactivateService(_ context.Context, tenantID, serviceID string) (monitorconfig.Service, error) {
	service, ok := r.services[serviceKey(serviceID)]
	if !ok || service.TenantID != tenantID {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeServiceNotFound, nil)
	}
	if service.LifecycleState != monitorconfig.ServiceLifecycleArchived {
		return monitorconfig.Service{}, sharederrors.New(sharederrors.CodeServiceNotArchived, nil)
	}
	if service.EnabledCount > 0 {
		service.LifecycleState = monitorconfig.ServiceLifecycleActive
	} else {
		service.LifecycleState = monitorconfig.ServiceLifecycleDraft
	}
	r.services[serviceKey(serviceID)] = service
	return service, nil
}

func (r *fakeMonitorRepository) GetMonitorStatus(_ context.Context, tenantID, serviceID, monitorID string) (resultstatus.MonitorStatus, bool, error) {
	status, ok := r.statuses[monitorKey(serviceID, monitorID)]
	if !ok || status.TenantID != tenantID {
		return resultstatus.MonitorStatus{}, false, nil
	}
	return status, true, nil
}

func (r *fakeMonitorRepository) ListMonitorRuns(_ context.Context, tenantID, serviceID, monitorID string, _ int32) ([]resultstatus.CheckRun, error) {
	runs := r.runs[monitorKey(serviceID, monitorID)]
	filtered := make([]resultstatus.CheckRun, 0, len(runs))
	for _, run := range runs {
		if run.TenantID == tenantID {
			filtered = append(filtered, run)
		}
	}
	return filtered, nil
}

func (r *fakeMonitorRepository) CreateManualRun(_ context.Context, monitor monitorconfig.Monitor, now time.Time) (manualRunRequestRecord, error) {
	run := manualRunRequestRecord{RunID: newRunID(now), ServiceID: monitor.ServiceID, MonitorID: monitor.MonitorID, TenantID: monitor.TenantID, Trigger: checkexecution.TriggerTypeManual, AcceptedAt: now.UTC().Format(time.RFC3339)}
	r.manual[run.RunID] = run
	return run, nil
}

func (r *fakeMonitorRepository) RecordExecutionResult(_ context.Context, monitor monitorconfig.Monitor, runID string, result checkexecution.ExecutionResult) error {
	key := monitorKey(monitor.ServiceID, monitor.MonitorID)
	status := resultstatus.NewMonitorStatus(result)
	r.statuses[key] = status
	return nil
}

func (r *fakeMonitorRepository) ListIncidents(_ context.Context, tenantID, status string) ([]dynamodbrecord.IncidentRecord, error) {
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(r.incidents))
	for _, incident := range r.incidents {
		if incident.TenantID == tenantID && matchesIncidentFilter(incident.Status, status) {
			incidents = append(incidents, incident)
		}
	}
	return incidents, nil
}

func (r *fakeMonitorRepository) GetIncident(_ context.Context, tenantID, incidentID string) (dynamodbrecord.IncidentRecord, bool, error) {
	incident, ok := r.incidents[incidentID]
	if !ok || incident.TenantID != tenantID {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	return incident, true, nil
}

func (r *fakeMonitorRepository) ListIncidentActivities(_ context.Context, tenantID, incidentID string) ([]dynamodbrecord.IncidentActivityRecord, error) {
	incident, ok := r.incidents[incidentID]
	if !ok || incident.TenantID != tenantID {
		return nil, nil
	}
	return append([]dynamodbrecord.IncidentActivityRecord(nil), r.activities[incidentID]...), nil
}

func (r *fakeMonitorRepository) ListMonitorIncidents(_ context.Context, tenantID, serviceID, monitorID string) ([]dynamodbrecord.IncidentRecord, error) {
	incidents := make([]dynamodbrecord.IncidentRecord, 0, len(r.incidents))
	for _, incident := range r.incidents {
		if incident.TenantID == tenantID && incident.ServiceID == serviceID && incident.MonitorID == monitorID {
			incidents = append(incidents, incident)
		}
	}
	return incidents, nil
}

func (r *fakeMonitorRepository) AcknowledgeIncident(_ context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	incident, ok := r.incidents[incidentID]
	if !ok || incident.TenantID != tenantID {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	if incident.Status != incidentStatusOpen {
		return dynamodbrecord.IncidentRecord{}, true, errIncidentNotActionable
	}
	incident.Status = incidentStatusAcknowledged
	incident.AcknowledgedAt = now.UTC().Format(time.RFC3339)
	incident.UpdatedAt = incident.AcknowledgedAt
	r.incidents[incidentID] = incident
	return incident, true, nil
}

func (r *fakeMonitorRepository) ResolveIncident(_ context.Context, tenantID, incidentID string, now time.Time) (dynamodbrecord.IncidentRecord, bool, error) {
	incident, ok := r.incidents[incidentID]
	if !ok || incident.TenantID != tenantID {
		return dynamodbrecord.IncidentRecord{}, false, nil
	}
	if incident.Status == incidentStatusResolved {
		return dynamodbrecord.IncidentRecord{}, true, errIncidentNotActionable
	}
	incident.Status = incidentStatusResolved
	incident.ResolvedAt = now.UTC().Format(time.RFC3339)
	incident.UpdatedAt = incident.ResolvedAt
	r.incidents[incidentID] = incident
	return incident, true, nil
}

func (r *fakeMonitorRepository) GetSchedulerConfig(_ context.Context, _ string) (dynamodbrecord.SchedulerConfigRecord, error) {
	return r.scheduler, nil
}

func (r *fakeMonitorRepository) UpdateSchedulerConfig(_ context.Context, _ string, config checkexecution.SchedulerConfig, now time.Time) (dynamodbrecord.SchedulerConfigRecord, error) {
	r.scheduler = dynamodbrecord.SchedulerConfigRecord{Config: config, UpdatedAt: now.UTC().Format(time.RFC3339)}
	return r.scheduler, nil
}

func (r *fakeMonitorRepository) ListMonitorAuditEvents(_ context.Context, tenantID, serviceID, monitorID string) ([]auditEventView, error) {
	if _, ok := r.monitors[monitorKey(serviceID, monitorID)]; !ok {
		return nil, nil
	}
	_ = tenantID
	return append([]auditEventView(nil), r.audit[monitorKey(serviceID, monitorID)]...), nil
}

func (r *fakeMonitorRepository) ListServiceAuditEvents(_ context.Context, tenantID, serviceID string) ([]auditEventView, error) {
	service, ok := r.services[serviceKey(serviceID)]
	if !ok || service.TenantID != tenantID {
		return nil, nil
	}
	return append([]auditEventView(nil), r.audit[serviceKey(serviceID)]...), nil
}

func (r *fakeMonitorRepository) CreateEscalationPolicy(_ context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	r.policies[policy.PolicyID] = policy
	return policy, nil
}

func (r *fakeMonitorRepository) ListEscalationPolicies(_ context.Context, tenantID string) ([]escalation.EscalationPolicy, error) {
	out := make([]escalation.EscalationPolicy, 0, len(r.policies))
	for _, policy := range r.policies {
		if policy.TenantID == tenantID {
			out = append(out, policy)
		}
	}
	return out, nil
}

func (r *fakeMonitorRepository) GetEscalationPolicy(_ context.Context, tenantID, policyID string) (*escalation.EscalationPolicy, error) {
	policy, ok := r.policies[policyID]
	if !ok || policy.TenantID != tenantID {
		return nil, nil
	}
	copy := policy
	return &copy, nil
}

func (r *fakeMonitorRepository) UpdateEscalationPolicy(_ context.Context, policy escalation.EscalationPolicy) (escalation.EscalationPolicy, error) {
	r.policies[policy.PolicyID] = policy
	return policy, nil
}

func (r *fakeMonitorRepository) DeleteEscalationPolicy(_ context.Context, tenantID, policyID string) error {
	policy, ok := r.policies[policyID]
	if ok && policy.TenantID == tenantID {
		delete(r.policies, policyID)
	}
	return nil
}

func (r *fakeMonitorRepository) ServiceReferencesEscalationPolicy(_ context.Context, tenantID, policyID string) (bool, error) {
	for _, service := range r.services {
		if service.TenantID == tenantID && strings.EqualFold(service.EscalationPolicyID, policyID) {
			return true, nil
		}
	}
	return false, nil
}

func (r *fakeMonitorRepository) CreateNotificationChannel(_ context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	r.channels[channel.ChannelID] = channel
	return channel, nil
}

func (r *fakeMonitorRepository) ListNotificationChannels(_ context.Context, tenantID string) ([]escalation.NotificationChannel, error) {
	out := make([]escalation.NotificationChannel, 0, len(r.channels))
	for _, channel := range r.channels {
		if channel.TenantID == tenantID {
			out = append(out, channel)
		}
	}
	return out, nil
}

func (r *fakeMonitorRepository) GetNotificationChannel(_ context.Context, tenantID, channelID string) (*escalation.NotificationChannel, error) {
	channel, ok := r.channels[channelID]
	if !ok || channel.TenantID != tenantID {
		return nil, nil
	}
	copy := channel
	return &copy, nil
}

func (r *fakeMonitorRepository) UpdateNotificationChannel(_ context.Context, channel escalation.NotificationChannel) (escalation.NotificationChannel, error) {
	r.channels[channel.ChannelID] = channel
	return channel, nil
}

func (r *fakeMonitorRepository) DeleteNotificationChannel(_ context.Context, tenantID, channelID string) error {
	channel, ok := r.channels[channelID]
	if ok && channel.TenantID == tenantID {
		delete(r.channels, channelID)
	}
	return nil
}

func (r *fakeMonitorRepository) ChannelsReferencedByRoutes(_ context.Context, tenantID, channelID string) ([]routeReference, error) {
	references := []routeReference{}
	for _, policy := range r.policies {
		if policy.TenantID == tenantID && policyReferencesChannel(policy, channelID) {
			references = append(references, routeReference{PolicyID: policy.PolicyID, Name: policy.Name})
		}
	}
	return references, nil
}

func (r *fakeMonitorRepository) GetEscalationState(_ context.Context, tenantID, incidentID string) (*escalation.EscalationState, error) {
	for _, state := range r.escalationStates {
		if state.TenantID == tenantID && strings.EqualFold(state.IncidentID, incidentID) {
			copy := state
			return &copy, nil
		}
	}
	return nil, nil
}

func addRecord(t *testing.T, items map[string]map[string]sharedaws.AttributeValue, record any) {
	t.Helper()
	item, err := sharedaws.MarshalMap(record)
	if err != nil {
		t.Fatalf("MarshalMap: %v", err)
	}
	pk := item["PK"].(*sharedaws.AttributeValueMemberS).Value
	sk := item["SK"].(*sharedaws.AttributeValueMemberS).Value
	items[pk+"|"+sk] = item
}

func addRawItem(items map[string]map[string]sharedaws.AttributeValue, pk, sk, entityType string) {
	items[pk+"|"+sk] = map[string]sharedaws.AttributeValue{
		"PK":         &sharedaws.AttributeValueMemberS{Value: pk},
		"SK":         &sharedaws.AttributeValueMemberS{Value: sk},
		"EntityType": &sharedaws.AttributeValueMemberS{Value: entityType},
	}
}

func deletedKeysFromTransactions(inputs []*sharedaws.DynamoDBTransactWriteItemsInput) map[string]struct{} {
	keys := map[string]struct{}{}
	for _, input := range inputs {
		for _, item := range input.TransactItems {
			if item.Delete == nil {
				continue
			}
			pk := item.Delete.Key["PK"].(*sharedaws.AttributeValueMemberS).Value
			sk := item.Delete.Key["SK"].(*sharedaws.AttributeValueMemberS).Value
			keys[pk+"|"+sk] = struct{}{}
		}
	}
	return keys
}

func assertDeletedKey(t *testing.T, keys map[string]struct{}, pk, sk string) {
	t.Helper()
	if _, ok := keys[pk+"|"+sk]; !ok {
		t.Fatalf("expected delete key %s|%s", pk, sk)
	}
}

func assertNotDeletedKey(t *testing.T, keys map[string]struct{}, pk, sk string) {
	t.Helper()
	if _, ok := keys[pk+"|"+sk]; ok {
		t.Fatalf("did not expect delete key %s|%s", pk, sk)
	}
}

func transactionPutsAction(inputs []*sharedaws.DynamoDBTransactWriteItemsInput, action string) bool {
	for _, input := range inputs {
		for _, item := range input.TransactItems {
			if item.Put == nil {
				continue
			}
			value, ok := item.Put.Item["Action"].(*sharedaws.AttributeValueMemberS)
			if ok && value.Value == action {
				return true
			}
		}
	}
	return false
}

func putServiceStatusFromTransactions(t *testing.T, inputs []*sharedaws.DynamoDBTransactWriteItemsInput) dynamodbrecord.ServiceStatusRecord {
	t.Helper()
	for _, input := range inputs {
		for _, item := range input.TransactItems {
			if item.Put == nil {
				continue
			}
			entityType, ok := item.Put.Item["EntityType"].(*sharedaws.AttributeValueMemberS)
			if !ok || entityType.Value != dynamodbschema.EntityServiceStatus {
				continue
			}
			var status dynamodbrecord.ServiceStatusRecord
			if err := sharedaws.UnmarshalMap(item.Put.Item, &status); err != nil {
				t.Fatalf("UnmarshalMap: %v", err)
			}
			return status
		}
	}
	t.Fatal("service status put not found")
	return dynamodbrecord.ServiceStatusRecord{}
}

func TestCreateService(t *testing.T) {
	repo := newFakeMonitorRepository()
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services", Body: `{"name":"Auth"}`, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	if len(repo.services) != 1 {
		t.Fatal("service not stored")
	}
	var created monitorconfig.Service
	for _, s := range repo.services {
		created = s
		break
	}
	if created.ServiceID == "" {
		t.Fatal("serviceId not generated")
	}
	if created.LifecycleState != monitorconfig.ServiceLifecycleDraft {
		t.Fatalf("new service should have draft lifecycle, got %v", created.LifecycleState)
	}
}

func TestCreateMonitorUnderService(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleDraft}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/monitors", PathParameters: map[string]string{"serviceId": "auth"}, Body: `{"name":"Homepage","type":"http","intervalSeconds":60,"probeLocations":["iad"],"enabled":true,"http":{"target":"https://example.com","method":"GET","timeoutMs":5000}}`, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	if len(repo.monitors) != 1 {
		t.Fatal("monitor not stored")
	}
	var created monitorconfig.Monitor
	for _, m := range repo.monitors {
		created = m
		break
	}
	if created.ServiceID != "auth" || created.MonitorID == "" {
		t.Fatalf("created monitor = %+v", created)
	}
	if repo.services[serviceKey("auth")].LifecycleState != monitorconfig.ServiceLifecycleDraft {
		t.Fatal("draft service should stay draft after first enabled monitor")
	}
}

func TestDeleteLastMonitorFromActiveServiceConflicts(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleActive}
	repo.monitors[monitorKey("auth", "public-http")] = monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: true, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/monitors/public-http", PathParameters: map[string]string{"serviceId": "auth", "monitorId": "public-http"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusConflict)
	}
}

func TestDeleteDraftService(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleDraft}
	repo.monitors[monitorKey("auth", "public-http")] = monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: false, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}
	if _, ok := repo.services[serviceKey("auth")]; ok {
		t.Fatal("service was not deleted")
	}
	if _, ok := repo.monitors[monitorKey("auth", "public-http")]; ok {
		t.Fatal("child monitor was not deleted")
	}
}

func TestDeleteActiveServiceConflicts(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleActive}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusConflict)
	}
}

func TestDeleteMissingService(t *testing.T) {
	handler := newMonitorHandler(newFakeMonitorRepository(), defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/missing", PathParameters: map[string]string{"serviceId": "missing"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestDeleteMonitor(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleActive, MonitorCount: 2, EnabledCount: 2}
	repo.monitors[monitorKey("auth", "public-http")] = monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: true, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	repo.monitors[monitorKey("auth", "admin-http")] = monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "admin-http", Name: "Admin", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: true, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://admin.example.com", Method: "GET", TimeoutMs: 5000}}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/monitors/public-http", PathParameters: map[string]string{"serviceId": "auth", "monitorId": "public-http"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNoContent)
	}
	if _, ok := repo.monitors[monitorKey("auth", "public-http")]; ok {
		t.Fatal("monitor was not deleted")
	}
	service := repo.services[serviceKey("auth")]
	if service.MonitorCount != 1 || service.EnabledCount != 1 || service.LifecycleState != monitorconfig.ServiceLifecycleActive {
		t.Fatalf("service summary = %+v", service)
	}
}

func TestDeleteMissingMonitor(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleDraft}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/monitors/missing", PathParameters: map[string]string{"serviceId": "auth", "monitorId": "missing"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestUpdateMonitorDefaultsLegacyThresholds(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleDraft}
	repo.monitors[monitorKey("auth", "public-http")] = monitorconfig.Monitor{
		TenantID:        defaultTenantID,
		ServiceID:       "auth",
		MonitorID:       "public-http",
		Name:            "Homepage",
		Type:            monitorconfig.MonitorTypeHTTP,
		IntervalSeconds: 60,
		ProbeLocations:  []string{"iad"},
		Enabled:         true,
		HTTP:            &monitorconfig.HTTPConfiguration{Target: "https://old.example.com", Method: "GET", TimeoutMs: 5000},
	}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/services/auth/monitors/public-http",
		PathParameters: map[string]string{"serviceId": "auth", "monitorId": "public-http"},
		Body:           `{"name":"Homepage","intervalSeconds":60,"probeLocations":["iad"],"http":{"target":"https://new.example.com","method":"GET","timeoutMs":5000,"expectedStatusCodes":[200]}}`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPatch}},
	}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d body=%s", response.StatusCode, http.StatusOK, response.Body)
	}
	stored := repo.monitors[monitorKey("auth", "public-http")]
	if stored.FailureThreshold != 1 || stored.RecoveryThreshold != 1 {
		t.Fatalf("thresholds = %d/%d, want 1/1", stored.FailureThreshold, stored.RecoveryThreshold)
	}
}

func TestGetServiceIncludesNestedMonitors(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleActive, MonitorCount: 1, EnabledCount: 1, RollupStatus: rollupUp}
	repo.monitors[monitorKey("auth", "public-http")] = monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: true, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	repo.statuses[monitorKey("auth", "public-http")] = resultstatus.MonitorStatus{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", CurrentStatus: "UP", LastCheckedAt: time.Date(2026, 5, 25, 1, 0, 0, 0, time.UTC), LastOutcome: checkexecution.OutcomeSuccess}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	var body struct {
		Status string          `json:"status"`
		Data   serviceResponse `json:"data"`
	}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Status != "success" {
		t.Fatalf("body.Status = %q, want success", body.Status)
	}
	if len(body.Data.Monitors) != 1 || body.Data.Monitors[0].ServiceID != "auth" {
		t.Fatalf("body.Data.Monitors = %+v", body.Data.Monitors)
	}
}

func TestNotificationChannelCRUD(t *testing.T) {
	repo := newFakeMonitorRepository()
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	createReq := events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/notification-channels",
		Body:           `{"name":"Primary Telegram","type":"telegram","target":"chat-1","config":{"botToken":"secret"}}`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}
	createResp, err := handler.handleRequest(context.Background(), createReq)
	if err != nil {
		t.Fatalf("create handleRequest returned error: %v", err)
	}
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d, want %d", createResp.StatusCode, http.StatusCreated)
	}
	var created struct {
		Status string                      `json:"status"`
		Data   notificationChannelResponse `json:"data"`
	}
	if err := json.Unmarshal([]byte(createResp.Body), &created); err != nil {
		t.Fatalf("json.Unmarshal create: %v", err)
	}
	if created.Status != "success" {
		t.Fatalf("created.Status = %q, want success", created.Status)
	}
	if !strings.Contains(string(created.Data.Config), "***REDACTED***") || strings.Contains(string(created.Data.Config), "secret") {
		t.Fatalf("created config not redacted: %s", created.Data.Config)
	}

	listResp, err := handler.handleRequest(context.Background(), events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/notification-channels", RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}}})
	if err != nil {
		t.Fatalf("list handleRequest returned error: %v", err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list status = %d, want %d", listResp.StatusCode, http.StatusOK)
	}
	var listed struct {
		Status string                           `json:"status"`
		Data   listNotificationChannelsResponse `json:"data"`
	}
	if err := json.Unmarshal([]byte(listResp.Body), &listed); err != nil {
		t.Fatalf("json.Unmarshal list: %v", err)
	}
	if len(listed.Data.Channels) != 1 || listed.Data.Channels[0].ChannelID != created.Data.ChannelID {
		t.Fatalf("listed channels = %+v", listed.Data.Channels)
	}

	updateReq := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/notification-channels/" + created.Data.ChannelID, PathParameters: map[string]string{"channelId": created.Data.ChannelID}, Body: `{"name":"Renamed Telegram"}`, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPut}}}
	updateResp, err := handler.handleRequest(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("update handleRequest returned error: %v", err)
	}
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status = %d, want %d", updateResp.StatusCode, http.StatusOK)
	}
	if repo.channels[created.Data.ChannelID].Name != "Renamed Telegram" {
		t.Fatalf("channel name = %q", repo.channels[created.Data.ChannelID].Name)
	}

	deleteReq := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/notification-channels/" + created.Data.ChannelID, PathParameters: map[string]string{"channelId": created.Data.ChannelID}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}
	deleteResp, err := handler.handleRequest(context.Background(), deleteReq)
	if err != nil {
		t.Fatalf("delete handleRequest returned error: %v", err)
	}
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", deleteResp.StatusCode, http.StatusNoContent)
	}
}

func TestCreateNotificationChannelValidation(t *testing.T) {
	handler := newMonitorHandler(newFakeMonitorRepository(), defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/notification-channels", Body: `{"name":"Pager","type":"telegram","target":"chat-1","config":{}}`, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	var envelope struct {
		Status string `json:"status"`
		Reason struct {
			Code    string         `json:"code"`
			Details map[string]any `json:"details"`
		} `json:"reason"`
	}
	if err := json.Unmarshal([]byte(response.Body), &envelope); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if envelope.Status != "error" || envelope.Reason.Code != "VALIDATION_FAILED" {
		t.Fatalf("envelope = %+v", envelope)
	}
	if got := envelope.Reason.Details["field"]; got != "config.botToken" {
		t.Fatalf("details.field = %v, want config.botToken", got)
	}
	if got := envelope.Reason.Details["reason"]; got != "required" {
		t.Fatalf("details.reason = %v, want required", got)
	}
}

func TestDeleteNotificationChannelBlockedWhenReferenced(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.channels["CH_1"] = escalation.NotificationChannel{TenantID: defaultTenantID, ChannelID: "CH_1", Name: "Primary", Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"secret"}`)}
	repo.policies["POL_1"] = escalation.EscalationPolicy{TenantID: defaultTenantID, PolicyID: "POL_1", Name: "Route", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1"}}}, OffHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1"}}}}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/notification-channels/CH_1", PathParameters: map[string]string{"channelId": "CH_1"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusConflict)
	}
	if !strings.Contains(response.Body, "referencingRoutes") {
		t.Fatalf("body = %s", response.Body)
	}
}

func TestCreateEscalationPolicy(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.channels["CH_1"] = escalation.NotificationChannel{TenantID: defaultTenantID, ChannelID: "CH_1", Name: "Telegram", Type: escalation.ChannelTypeTelegram, Target: "ops-primary", Config: json.RawMessage(`{"botToken":"token"}`)}
	repo.channels["CH_2"] = escalation.NotificationChannel{TenantID: defaultTenantID, ChannelID: "CH_2", Name: "Email", Type: escalation.ChannelTypeEmail, Target: "ops@example.com", Config: json.RawMessage(`{"apiKey":"key","fromEmail":"from@example.com"}`)}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/escalation-policies",
		Body:           `{"name":"Primary On Call","businessHoursPath":{"steps":[{"delayMinutes":0,"channelId":"CH_1"}]},"offHoursPath":{"steps":[{"delayMinutes":5,"channelId":"CH_2"}]}}`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusCreated)
	}
	var body struct {
		Status string                   `json:"status"`
		Data   escalationPolicyResponse `json:"data"`
	}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Status != "success" {
		t.Fatalf("body.Status = %q, want success", body.Status)
	}
	if body.Data.PolicyID == "" {
		t.Fatal("PolicyID is empty")
	}
	if body.Data.Name != "Primary On Call" {
		t.Fatalf("Name = %q, want Primary On Call", body.Data.Name)
	}
	if _, ok := repo.policies[body.Data.PolicyID]; !ok {
		t.Fatalf("policy %q not stored in repo", body.Data.PolicyID)
	}
}

func TestCreateEscalationPolicyRejectsEmptyPath(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.channels["CH_1"] = escalation.NotificationChannel{TenantID: defaultTenantID, ChannelID: "CH_1", Name: "Telegram", Type: escalation.ChannelTypeTelegram, Target: "ops", Config: json.RawMessage(`{"botToken":"token"}`)}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/escalation-policies",
		Body:           `{"name":"Primary On Call","businessHoursPath":{"steps":[]},"offHoursPath":{"steps":[{"delayMinutes":0,"channelId":"CH_1"}]}}`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
}

func TestDeleteEscalationPolicyBlockedWhenReferenced(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.policies["POL_1"] = escalation.EscalationPolicy{TenantID: defaultTenantID, PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1", DelayMinutes: 0}}}, OffHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1", DelayMinutes: 0}}}}
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", EscalationPolicyID: "POL_1"}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/escalation-policies/POL_1", PathParameters: map[string]string{"policyId": "POL_1"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodDelete}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusConflict)
	}
}

func TestGetServiceEscalationPolicy(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.policies["POL_1"] = escalation.EscalationPolicy{TenantID: defaultTenantID, PolicyID: "POL_1", Name: "Primary", BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_1", DelayMinutes: 0}}}, OffHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_2", DelayMinutes: 0}}}}
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", EscalationPolicyID: "POL_1"}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/escalation-policy", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	var body struct {
		Status string                   `json:"status"`
		Data   escalationPolicyResponse `json:"data"`
	}
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if body.Data.PolicyID != "POL_1" {
		t.Fatalf("PolicyID = %q, want POL_1", body.Data.PolicyID)
	}
}

func TestGetEscalationPolicyMigratesLegacyInlineChannelsOnce(t *testing.T) {
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}
	repo := newDynamoMonitorRepository(client, "table-name")
	repo.now = func() time.Time { return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC) }
	legacy := escalation.EscalationPolicy{
		TenantID:          defaultTenantID,
		PolicyID:          "POL_1",
		Name:              "Legacy",
		CreatedAt:         "2026-05-24T00:00:00Z",
		UpdatedAt:         "2026-05-24T00:00:00Z",
		BusinessHoursPath: escalation.EscalationPath{Steps: []escalation.EscalationStep{{DelayMinutes: 0, Channels: []escalation.ChannelConfig{{Type: escalation.ChannelTypeTelegram, Target: "chat-1", Config: json.RawMessage(`{"botToken":"secret"}`)}}}}},
		OffHoursPath:      escalation.EscalationPath{Steps: []escalation.EscalationStep{{ChannelID: "CH_EXISTING", DelayMinutes: 5}}},
	}
	addRecord(t, client.items, newEscalationPolicyItemRecord(legacy))

	first, err := repo.GetEscalationPolicy(context.Background(), defaultTenantID, "POL_1")
	if err != nil {
		t.Fatalf("first GetEscalationPolicy returned error: %v", err)
	}
	second, err := repo.GetEscalationPolicy(context.Background(), defaultTenantID, "POL_1")
	if err != nil {
		t.Fatalf("second GetEscalationPolicy returned error: %v", err)
	}
	if first.BusinessHoursPath.Steps[0].ChannelID == "" || second.BusinessHoursPath.Steps[0].ChannelID != first.BusinessHoursPath.Steps[0].ChannelID {
		t.Fatalf("migrated channel IDs = %q / %q", first.BusinessHoursPath.Steps[0].ChannelID, second.BusinessHoursPath.Steps[0].ChannelID)
	}
	channelCount := 0
	for _, item := range client.items {
		entity, ok := item["EntityType"].(*sharedaws.AttributeValueMemberS)
		if ok && entity.Value == dynamodbschema.EntityNotificationChannel {
			channelCount++
		}
	}
	if channelCount != 1 {
		t.Fatalf("notification channel count = %d, want 1", channelCount)
	}
}

func TestCreateMonitorWritesServiceMonitorRefKeys(t *testing.T) {
	monitor := monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: true, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	client := &fakeDynamoClient{}
	repo := newDynamoMonitorRepository(client, "table-name")
	repo.now = func() time.Time { return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC) }
	repo.client = client
	service := monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleDraft, CreatedAt: "2026-05-25T00:00:00Z", UpdatedAt: "2026-05-25T00:00:00Z"}
	serviceItem, _ := sharedaws.MarshalMap(dynamodbrecord.NewServiceItemRecord(service))
	client.items = map[string]map[string]sharedaws.AttributeValue{"SERVICE#DEFAULT#AUTH|META": serviceItem}
	client.queryOutput = &sharedaws.DynamoDBQueryOutput{}

	if _, err := repo.CreateMonitor(context.Background(), monitor); err != nil {
		t.Fatalf("CreateMonitor returned error: %v", err)
	}
	if client.transactInput == nil || len(client.transactInput.TransactItems) == 0 {
		t.Fatal("transact input not captured")
	}
	item := client.transactInput.TransactItems[1].Put.Item
	pk := item["PK"].(*sharedaws.AttributeValueMemberS).Value
	sk := item["SK"].(*sharedaws.AttributeValueMemberS).Value
	if pk != "SERVICE#DEFAULT#AUTH" || sk != "MONITOR#PUBLIC-HTTP" {
		t.Fatalf("service monitor ref key = %s/%s", pk, sk)
	}
	if got := sharedaws.ToString(client.transactInput.TransactItems[0].Put.TableName); got != "table-name" {
		t.Fatalf("table name = %q, want table-name", got)
	}
}

func TestDeleteServiceDeletesActiveConfigAndWritesAudit(t *testing.T) {
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}
	repo := newDynamoMonitorRepository(client, "table-name")
	repo.now = func() time.Time { return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC) }
	service := monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleArchived, CreatedAt: "2026-05-25T00:00:00Z", UpdatedAt: "2026-05-25T00:00:00Z"}
	monitor := monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: false, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	addRecord(t, client.items, dynamodbrecord.NewServiceItemRecord(service))
	addRecord(t, client.items, dynamodbrecord.NewServiceRefItemRecord(service))
	addRecord(t, client.items, dynamodbrecord.NewServiceStatusItemRecord(service, "2026-05-25T00:00:00Z"))
	addRecord(t, client.items, dynamodbrecord.NewMonitorItemRecord(monitor))
	addRecord(t, client.items, dynamodbrecord.NewServiceMonitorRefItemRecord(monitor))
	addRecord(t, client.items, newDefaultMonitorStatusRecord(monitor, "2026-05-25T00:00:00Z"))
	addRawItem(client.items, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "RUN#2026-05-25T00:00:00Z#RUN_1", dynamodbschema.EntityCheckRun)
	addRecord(t, client.items, dynamodbrecord.NewIncidentMonitorItemRecord(dynamodbrecord.IncidentRecord{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", IncidentID: "INC_1", Summary: "Down", Status: incidentStatusOpen, OpenedAt: "2026-05-25T00:00:00Z", UpdatedAt: "2026-05-25T00:00:00Z"}))
	oldAudit := dynamodbrecord.NewAuditEventRecord(time.Date(2026, 5, 24, 0, 0, 0, 0, time.UTC), "AUD_test", defaultTenantID, "MONITOR_CREATED", "auth", "public-http")
	addRecord(t, client.items, oldAudit)

	deleted, err := repo.DeleteService(context.Background(), defaultTenantID, "auth")
	if err != nil {
		t.Fatalf("DeleteService returned error: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteService returned deleted=false")
	}
	deletedKeys := deletedKeysFromTransactions(client.transactInputs)
	assertDeletedKey(t, deletedKeys, dynamodbschema.ServicePK(defaultTenantID, "auth"), "META")
	assertDeletedKey(t, deletedKeys, dynamodbschema.ServicePK(defaultTenantID, "auth"), "STATUS")
	assertDeletedKey(t, deletedKeys, dynamodbschema.TenantPK(defaultTenantID), dynamodbschema.ServiceRefSK("auth"))
	assertDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "META")
	assertDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "STATUS")
	assertNotDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "RUN#2026-05-25T00:00:00Z#RUN_1")
	assertNotDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "INCIDENT#2026-05-25T00:00:00Z#INC_1")
	assertNotDeletedKey(t, deletedKeys, oldAudit.PK, oldAudit.SK)
	if !transactionPutsAction(client.transactInputs, "SERVICE_DELETED") {
		t.Fatal("SERVICE_DELETED audit event was not written")
	}
}

func TestDeleteMonitorDeletesConfigRecalculatesServiceAndWritesAudit(t *testing.T) {
	client := &fakeDynamoClient{items: map[string]map[string]sharedaws.AttributeValue{}}
	repo := newDynamoMonitorRepository(client, "table-name")
	repo.now = func() time.Time { return time.Date(2026, 5, 25, 0, 0, 0, 0, time.UTC) }
	service := monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleActive, CreatedAt: "2026-05-25T00:00:00Z", UpdatedAt: "2026-05-25T00:00:00Z"}
	deletedMonitor := monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "public-http", Name: "Homepage", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: true, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://example.com", Method: "GET", TimeoutMs: 5000}}
	remainingMonitor := monitorconfig.Monitor{TenantID: defaultTenantID, ServiceID: "auth", MonitorID: "admin-http", Name: "Admin", Type: monitorconfig.MonitorTypeHTTP, IntervalSeconds: 60, ProbeLocations: []string{"iad"}, Enabled: false, HTTP: &monitorconfig.HTTPConfiguration{Target: "https://admin.example.com", Method: "GET", TimeoutMs: 5000}}
	addRecord(t, client.items, dynamodbrecord.NewServiceItemRecord(service))
	addRecord(t, client.items, dynamodbrecord.NewServiceStatusItemRecord(service, "2026-05-25T00:00:00Z"))
	addRecord(t, client.items, dynamodbrecord.NewMonitorItemRecord(deletedMonitor))
	addRecord(t, client.items, dynamodbrecord.NewServiceMonitorRefItemRecord(deletedMonitor))
	addRecord(t, client.items, newDefaultMonitorStatusRecord(deletedMonitor, "2026-05-25T00:00:00Z"))
	addRecord(t, client.items, dynamodbrecord.NewMonitorItemRecord(remainingMonitor))
	addRecord(t, client.items, dynamodbrecord.NewServiceMonitorRefItemRecord(remainingMonitor))
	addRecord(t, client.items, newDefaultMonitorStatusRecord(remainingMonitor, "2026-05-25T00:00:00Z"))
	addRawItem(client.items, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "RUN#2026-05-25T00:00:00Z#RUN_1", dynamodbschema.EntityCheckRun)

	deleted, err := repo.DeleteMonitor(context.Background(), defaultTenantID, "auth", "public-http")
	if err != nil {
		t.Fatalf("DeleteMonitor returned error: %v", err)
	}
	if !deleted {
		t.Fatal("DeleteMonitor returned deleted=false")
	}
	deletedKeys := deletedKeysFromTransactions(client.transactInputs)
	assertDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "META")
	assertDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "STATUS")
	assertDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "META")
	assertDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "STATUS")
	assertDeletedKey(t, deletedKeys, dynamodbschema.ServicePK(defaultTenantID, "auth"), dynamodbschema.ServiceMonitorRefSK("public-http"))
	assertNotDeletedKey(t, deletedKeys, dynamodbschema.MonitorPK(defaultTenantID, "auth", "public-http"), "RUN#2026-05-25T00:00:00Z#RUN_1")
	if !transactionPutsAction(client.transactInputs, "MONITOR_DELETED") {
		t.Fatal("MONITOR_DELETED audit event was not written")
	}
	status := putServiceStatusFromTransactions(t, client.transactInputs)
	if status.MonitorCount != 1 || status.EnabledMonitorCount != 0 || status.LifecycleState != string(monitorconfig.ServiceLifecycleDraft) {
		t.Fatalf("service status = %+v", status)
	}
}

func TestArchiveActiveService(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleActive}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/archive", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if repo.services[serviceKey("auth")].LifecycleState != monitorconfig.ServiceLifecycleArchived {
		t.Fatalf("lifecycle = %v, want archived", repo.services[serviceKey("auth")].LifecycleState)
	}
}

func TestArchiveDraftService(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleDraft}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/archive", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if repo.services[serviceKey("auth")].LifecycleState != monitorconfig.ServiceLifecycleArchived {
		t.Fatalf("lifecycle = %v, want archived", repo.services[serviceKey("auth")].LifecycleState)
	}
}

func TestArchiveAlreadyArchivedServiceIsIdempotent(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleArchived}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/archive", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
}

func TestArchiveMissingService(t *testing.T) {
	handler := newMonitorHandler(newFakeMonitorRepository(), defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/missing/archive", PathParameters: map[string]string{"serviceId": "missing"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestReactivateArchivedServiceWithEnabledMonitors(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleArchived, EnabledCount: 1}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/reactivate", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if repo.services[serviceKey("auth")].LifecycleState != monitorconfig.ServiceLifecycleActive {
		t.Fatalf("lifecycle = %v, want active", repo.services[serviceKey("auth")].LifecycleState)
	}
}

func TestReactivateArchivedServiceWithNoEnabledMonitors(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleArchived, EnabledCount: 0}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/reactivate", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusOK)
	}
	if repo.services[serviceKey("auth")].LifecycleState != monitorconfig.ServiceLifecycleDraft {
		t.Fatalf("lifecycle = %v, want draft", repo.services[serviceKey("auth")].LifecycleState)
	}
}

func TestReactivateNonArchivedService(t *testing.T) {
	repo := newFakeMonitorRepository()
	repo.services[serviceKey("auth")] = monitorconfig.Service{TenantID: defaultTenantID, ServiceID: "auth", Name: "Auth", LifecycleState: monitorconfig.ServiceLifecycleActive}
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/auth/reactivate", PathParameters: map[string]string{"serviceId": "auth"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusConflict)
	}
}

func TestReactivateMissingService(t *testing.T) {
	handler := newMonitorHandler(newFakeMonitorRepository(), defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/missing/reactivate", PathParameters: map[string]string{"serviceId": "missing"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusNotFound)
	}
}

func TestUpdateSchedulerConfigRequiresStopControlWhenEnabling(t *testing.T) {
	handler := newMonitorHandler(newFakeMonitorRepository(), defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/admin/scheduler-config", Body: `{"recurringEnabled":true}`, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPatch}}}

	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.StatusCode, http.StatusBadRequest)
	}
	envelope := decodeEnvelope(t, response.Body)
	if envelope.Reason.Code != "VALIDATION_FAILED" {
		t.Fatalf("Code = %s, want VALIDATION_FAILED", envelope.Reason.Code)
	}
	if envelope.Reason.Details["field"] != "stopControlMode" {
		t.Fatalf("details.field = %v, want stopControlMode", envelope.Reason.Details["field"])
	}
}

type typedErrorEnvelope struct {
	Status string `json:"status"`
	Reason struct {
		Code    string         `json:"code"`
		Details map[string]any `json:"details"`
	} `json:"reason"`
}

func decodeEnvelope(t *testing.T, body string) typedErrorEnvelope {
	t.Helper()
	var envelope typedErrorEnvelope
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	return envelope
}

func TestHandlerRoutesTypedSentinelToWireCode(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode string
		wantHTTP int
	}{
		{"service already exists", errServiceAlreadyExists, "SERVICE_ALREADY_EXISTS", http.StatusConflict},
		{"cannot delete active service", errCannotDeleteActiveService, "SERVICE_ACTIVE", http.StatusConflict},
		{"monitor already exists", errMonitorAlreadyExists, "MONITOR_ALREADY_EXISTS", http.StatusConflict},
		{"cannot delete last monitor from active service", errCannotDeleteLastMonitorFromActiveService, "LAST_MONITOR", http.StatusConflict},
		{"incident not actionable", errIncidentNotActionable, "INCIDENT_NOT_ACTIONABLE", http.StatusConflict},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			status, env := sharederrors.Respond(tc.err)
			if status != tc.wantHTTP {
				t.Fatalf("status = %d, want %d", status, tc.wantHTTP)
			}
			if env.Reason == nil || env.Reason.Code != tc.wantCode {
				t.Fatalf("env.Reason = %+v, want code %s", env.Reason, tc.wantCode)
			}
		})
	}
}

func TestErrMissingTableNameReachesWireAsInternalWithNilDetails(t *testing.T) {
	status, env := sharederrors.Respond(errMissingTableName)
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if env.Reason == nil || env.Reason.Code != "INTERNAL" {
		t.Fatalf("env.Reason = %+v", env.Reason)
	}
	if env.Reason.Details != nil {
		t.Fatalf("INTERNAL details leaked: %v", env.Reason.Details)
	}
}

func TestNonTypedErrorReachesWireAsInternalWithNilDetails(t *testing.T) {
	status, env := sharederrors.Respond(stderrors.New("boom"))
	if status != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", status)
	}
	if env.Reason.Code != "INTERNAL" {
		t.Fatalf("Code = %s", env.Reason.Code)
	}
	if env.Reason.Details != nil {
		t.Fatalf("Details leaked: %v", env.Reason.Details)
	}
}

type typedRepoStub struct {
	monitorRepository
	forced error
}

func (s typedRepoStub) ListServices(context.Context, string) ([]monitorconfig.Service, error) {
	return nil, s.forced
}

func TestHandlerReachesInternalForUntypedRepositoryError(t *testing.T) {
	handler := newMonitorHandler(typedRepoStub{forced: stderrors.New("storage exploded")}, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services", RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}}}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", response.StatusCode)
	}
	envelope := decodeEnvelope(t, response.Body)
	if envelope.Reason.Code != "INTERNAL" {
		t.Fatalf("Code = %s, want INTERNAL", envelope.Reason.Code)
	}
	if envelope.Reason.Details != nil {
		t.Fatalf("Details leaked: %v", envelope.Reason.Details)
	}
}

func TestHandlerCodeWireIdentity(t *testing.T) {
	cases := map[sharederrors.Code]string{
		sharederrors.CodeNotFound:              "NOT_FOUND",
		sharederrors.CodeInvalidJSON:           "INVALID_JSON",
		sharederrors.CodeValidationFailed:      "VALIDATION_FAILED",
		sharederrors.CodeImmutableField:        "IMMUTABLE_FIELD",
		sharederrors.CodeInlineChannelConfig:   "INLINE_CHANNEL_CONFIG",
		sharederrors.CodeServiceNotFound:       "SERVICE_NOT_FOUND",
		sharederrors.CodeServiceAlreadyExists:  "SERVICE_ALREADY_EXISTS",
		sharederrors.CodeServiceActive:         "SERVICE_ACTIVE",
		sharederrors.CodeServiceNotArchived:    "SERVICE_NOT_ARCHIVED",
		sharederrors.CodeServiceHasNoPolicy:    "SERVICE_HAS_NO_POLICY",
		sharederrors.CodeMonitorNotFound:       "MONITOR_NOT_FOUND",
		sharederrors.CodeMonitorAlreadyExists:  "MONITOR_ALREADY_EXISTS",
		sharederrors.CodeMonitorDisabled:       "MONITOR_DISABLED",
		sharederrors.CodeMonitorStatusNotFound: "MONITOR_STATUS_NOT_FOUND",
		sharederrors.CodeLastMonitor:           "LAST_MONITOR",
		sharederrors.CodeIncidentNotFound:      "INCIDENT_NOT_FOUND",
		sharederrors.CodeIncidentNotActionable: "INCIDENT_NOT_ACTIONABLE",
		sharederrors.CodePolicyNotFound:        "POLICY_NOT_FOUND",
		sharederrors.CodePolicyReferenced:      "POLICY_REFERENCED",
		sharederrors.CodeChannelNotFound:       "CHANNEL_NOT_FOUND",
		sharederrors.CodeInternal:              "INTERNAL",
	}
	for code, want := range cases {
		status, env := sharederrors.Respond(sharederrors.New(code, nil))
		if env.Reason.Code != want {
			t.Fatalf("code %s serializes as %s, want %s", code, env.Reason.Code, want)
		}
		if status == 0 {
			t.Fatalf("status missing for %s", code)
		}
	}
}

func TestArchiveServiceSurfacesTypedNotFoundOnWire(t *testing.T) {
	repo := newFakeMonitorRepository()
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/services/missing/archive", PathParameters: map[string]string{"serviceId": "missing"}, RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}}}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.StatusCode)
	}
	envelope := decodeEnvelope(t, response.Body)
	if envelope.Reason.Code != "SERVICE_NOT_FOUND" {
		t.Fatalf("Code = %s, want SERVICE_NOT_FOUND", envelope.Reason.Code)
	}
}

func TestNotFoundForUnknownRouteUsesTypedNotFound(t *testing.T) {
	handler := newMonitorHandler(newFakeMonitorRepository(), defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{RawPath: "/api/v1/unknown", RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodGet}}}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", response.StatusCode)
	}
	envelope := decodeEnvelope(t, response.Body)
	if envelope.Reason.Code != "NOT_FOUND" {
		t.Fatalf("Code = %s, want NOT_FOUND", envelope.Reason.Code)
	}
}

func TestEscalationPolicyValidationSurfacesFieldPath(t *testing.T) {
	repo := newFakeMonitorRepository()
	handler := newMonitorHandler(repo, defaultProbeLocationCatalog(), defaultTenantID)
	request := events.APIGatewayV2HTTPRequest{
		RawPath:        "/api/v1/escalation-policies",
		Body:           `{"name":"Primary","businessHoursPath":{"steps":[{"delayMinutes":0,"channelId":"CH_MISSING"}]},"offHoursPath":{"steps":[{"delayMinutes":5,"channelId":"CH_MISSING"}]}}`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{HTTP: events.APIGatewayV2HTTPRequestContextHTTPDescription{Method: http.MethodPost}},
	}
	response, err := handler.handleRequest(context.Background(), request)
	if err != nil {
		t.Fatalf("handleRequest returned error: %v", err)
	}
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", response.StatusCode)
	}
	envelope := decodeEnvelope(t, response.Body)
	if envelope.Reason.Code != "VALIDATION_FAILED" {
		t.Fatalf("Code = %s, want VALIDATION_FAILED", envelope.Reason.Code)
	}
	if got := envelope.Reason.Details["field"]; got != "businessHoursPath.steps[0].channelId" {
		t.Fatalf("field = %v, want businessHoursPath.steps[0].channelId", got)
	}
}
